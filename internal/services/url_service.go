package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptrace"
	"nitinjuyal1610/uptimeMonitor/internal/models"
	"nitinjuyal1610/uptimeMonitor/internal/repository"
	"nitinjuyal1610/uptimeMonitor/pkg/types"
	"os"
	"sync"
	"time"

	"gopkg.in/gomail.v2"
)

type UrlService struct {
	urlRepo     repository.UrlRepository
	statRepo    repository.StatRepository
	httpClient  *http.Client
	emailClient *gomail.Dialer
}

type RawStats struct {
	StatusCode         int           `json:"status_code"`
	ResponseTime       time.Duration `json:"response_time"`
	Ttfb               time.Duration `json:"ttfb"`
	ContentSize        int64         `json:"content_size"`
	IsUp               bool          `json:"is_up"`
	MonitorId          int           `json:"monitor_id"`
	ExpectedStatusCode int           `json:"expected_status_code"`
	RequestType        string        `json:"request_type"`
}

func NewUrlService(urlRepo repository.UrlRepository, statRepo repository.StatRepository) *UrlService {
	httpClient := &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 50,
			IdleConnTimeout:     90 * time.Second,
		},
	}

	//
	gmailUser := os.Getenv("GMAIL_USER")
	gmailPass := os.Getenv("GMAIL_PASS")

	if gmailUser == "" || gmailPass == "" {
		log.Fatal("Missing required environment variables: GMAIL_USER or GMAIL_PASS")
	}

	emailClient := gomail.NewDialer("smtp.gmail.com", 587, gmailUser, gmailPass)

	return &UrlService{urlRepo: urlRepo, httpClient: httpClient, statRepo: statRepo, emailClient: emailClient}
}

func (us *UrlService) CreateUrl(ctx context.Context, urlMonitor *models.UrlMonitors) (int, error) {
	return us.urlRepo.Create(ctx, urlMonitor)
}

func (us *UrlService) GetAllUrl(ctx context.Context, status string, keyword string, userId int) ([]*models.UrlMonitors, error) {
	return us.urlRepo.GetAll(ctx, status, keyword, userId)
}

func (us *UrlService) GetMonitorById(ctx context.Context, id int) (*models.UrlMonitors, error) {
	return us.urlRepo.GetById(ctx, id)
}

func (us *UrlService) ProcessDueMonitorURLs(ctx context.Context) error {
	// Channels for stats and Concurrent api calls limit
	// Fetch URLs to process
	monitors, err := us.urlRepo.GetDueMonitors(ctx)
	if err != nil {
		return fmt.Errorf("failed to get urls : %w", err)
	}
	statChan := make(chan *RawStats, len(monitors))
	limitChan := make(chan bool, 15)
	notifyChan := make(chan *types.NotifyAlert, len(monitors))
	fmt.Printf("%d urls to be processed \n", len(monitors))
	var wg sync.WaitGroup
	for _, url := range monitors {
		wg.Add(1)
		go func(targetUrl string, monitorId int, expectedStatusCode int, collectDetailedData bool, alertEmail string, maxFailThreshold int, consecutiveFails int) {
			defer wg.Done()
			//semaphore increment
			limitChan <- true
			defer func() {
				<-limitChan
			}()
			ctx, cancel := context.WithTimeout(context.Background(), time.Duration(url.TimeoutSeconds)*time.Second)
			defer cancel()
			stats, err := us.fetchStatsFromUrl(ctx, targetUrl, collectDetailedData)

			if err == nil && stats != nil && !stats.IsUp && url.MaxFailThreshold > 0 && (url.ConsecutiveFails+1 >= url.MaxFailThreshold) {
				notifyChan <- &types.NotifyAlert{
					Url:              targetUrl,
					MonitorId:        monitorId,
					ConsecutiveFails: consecutiveFails,
					AlertEmail:       alertEmail,
				}
			}
			if err == nil {
				stats.MonitorId = monitorId
				stats.ExpectedStatusCode = expectedStatusCode
				statChan <- stats
			}
		}(url.Url, url.ID, url.ExpectedStatusCode, url.CollectDetailedData, url.AlertEmail, url.MaxFailThreshold, url.ConsecutiveFails)
	}

	go func() {
		wg.Wait()
		close(statChan)
		close(limitChan)
		close(notifyChan)
	}()

	go us.handleNotifications(notifyChan)
	return us.saveResultsToDB(ctx, statChan)
}

func (us *UrlService) handleNotifications(notifyChan <-chan *types.NotifyAlert) {
	for notifyItem := range notifyChan {
		m := gomail.NewMessage()
		m.SetHeader("From", "n.juyal99@gmail.com")
		m.SetHeader("To", notifyItem.AlertEmail)
		m.SetHeader("Subject", "⚠ Monitor Alert: Website Down")

		body := fmt.Sprintf(`
			<html>
			<body style="font-family: Arial, sans-serif; color: #333;">
				<h2 style="color: #d32f2f;">🚨 Website Down Alert!</h2>
				<p><strong>URL:</strong> <a href="%s">%s</a></p>
				<p><strong>Consecutive Failures:</strong> %d</p>
				<p>Please check the website status immediately.</p>
				<hr>
				<p style="font-size: 12px; color: #777;">This is an automated message from your monitoring system.</p>
			</body>
			</html>
		`, notifyItem.Url, notifyItem.Url, notifyItem.ConsecutiveFails+1)

		m.SetBody("text/html", body)

		if err := us.emailClient.DialAndSend(m); err != nil {
			log.Printf("Failed to send notification to %s: %v", notifyItem.AlertEmail, err)
		}
	}
}

func (us *UrlService) saveResultsToDB(ctx context.Context, statChan <-chan *RawStats) error {
	const batchSize = 100
	for {
		batch := make([]*RawStats, 0, batchSize)
		for range batchSize {
			stat, ok := <-statChan
			if !ok {
				// Channel closed
				break
			}
			batch = append(batch, stat)
		}
		if len(batch) == 0 {
			break // No more stats
		}
		if err := us.processBatch(ctx, batch); err != nil {
			return fmt.Errorf("failed to process batch: %w", err)
		}
	}
	return nil
}

func (us *UrlService) processBatch(ctx context.Context, batch []*RawStats) error {
	if len(batch) == 0 {
		return nil
	}

	monitorChecks := make([]*models.MonitorCheck, len(batch))
	monitorUpdates := make(map[int]*models.UrlMonitors)

	currentTime := time.Now().UTC().Truncate(time.Minute)
	for i, raw := range batch {
		status := determineStatus(raw)

		monitorChecks[i] = &models.MonitorCheck{
			MonitorId:    raw.MonitorId,
			StatusCode:   raw.StatusCode,
			ResponseTime: raw.ResponseTime.Seconds(),
			IsUp:         raw.IsUp,
			RequestType:  raw.RequestType,
			Timestamp:    currentTime,
		}

		if raw.Ttfb != 0 {
			monitorChecks[i].Ttfb = raw.Ttfb.Seconds()
			monitorChecks[i].ContentSize = raw.ContentSize
		}

		monitorUpdates[raw.MonitorId] = &models.UrlMonitors{
			LastChecked: currentTime,
			Status:      status,
		}

		if !raw.IsUp {
			monitorUpdates[raw.MonitorId].ConsecutiveFails += 1
		} else {
			monitorUpdates[raw.MonitorId].ConsecutiveFails = 0
		}
	}

	if _, err := us.statRepo.BulkCreate(ctx, monitorChecks); err != nil {
		return fmt.Errorf("failed to save stats to DB: %w", err)
	}

	if err := us.urlRepo.BulkUpdate(ctx, monitorUpdates); err != nil {
		return fmt.Errorf("failed to update monitors: %w", err)
	}

	return nil
}

func (us *UrlService) fetchStatsFromUrl(ctx context.Context, url string, collectDetailedData bool) (*RawStats, error) {
	var (
		start        = time.Now()
		responseTime time.Duration
		ttfb         time.Duration
		contentSize  int64
		statusCode   int
	)

	method := http.MethodHead
	trace := &httptrace.ClientTrace{
		DNSStart:     func(info httptrace.DNSStartInfo) { fmt.Println("DNS lookup start:", url, info.Host, time.Since(start)) },
		DNSDone:      func(info httptrace.DNSDoneInfo) { fmt.Println("DNS lookup done", url, time.Since(start)) },
		ConnectStart: func(network, addr string) { fmt.Println("Connecting to:", addr, url, time.Since(start)) },
		ConnectDone: func(network, addr string, err error) {
			if err != nil {
				fmt.Println("Connection failed:", err)
			} else {
				fmt.Println("Connected:", addr, url, time.Since(start))
			}
		},
		GotFirstResponseByte: func() { fmt.Println("First byte received", url, time.Since((start))); ttfb = time.Since(start) },
	}

	if collectDetailedData {
		method = http.MethodGet
	}

	req, err := http.NewRequestWithContext(httptrace.WithClientTrace(ctx, trace), method, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := us.httpClient.Do(req)

	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			fmt.Println("Request timed out:", err)
			statusCode = 408
		} else if errors.Is(err, context.Canceled) {
			fmt.Println("Request was canceled:", err)
			statusCode = 499
		} else {
			fmt.Println("Failed to fetch response:", err)
			statusCode = 500
		}

		return &RawStats{
			RequestType:  method,
			StatusCode:   statusCode,
			ResponseTime: time.Since(start),
			IsUp:         false,
		}, nil
	}
	defer resp.Body.Close()

	var rawStats = &RawStats{
		RequestType: method,
	}

	if collectDetailedData {
		//copy body
		writtenBytes, err := io.Copy(io.Discard, resp.Body)
		if err != nil {
			log.Printf("Failed to Copy Bytes.... %v", err)
		}
		contentSize = writtenBytes
		rawStats.ContentSize = contentSize
		rawStats.Ttfb = ttfb
	}
	responseTime = time.Since(start)
	statusCode = resp.StatusCode

	rawStats.StatusCode = statusCode
	rawStats.ResponseTime = responseTime
	rawStats.IsUp = (statusCode >= 200 && statusCode < 400)

	fmt.Printf("Raw stats %s -> %+v\n", url, rawStats)
	return rawStats, nil
}

func (us *UrlService) UpdateMonitorStatus(ctx context.Context, id int, status string) error {
	updateMontior := &models.UrlMonitors{
		Status: models.Status(status),
	}
	return us.urlRepo.Update(ctx, id, updateMontior)
}

func determineStatus(raw *RawStats) models.Status {
	switch {
	case raw.IsUp && raw.StatusCode == raw.ExpectedStatusCode:
		return models.StatusUp
	case raw.IsUp && raw.StatusCode != raw.ExpectedStatusCode:
		return models.StatusError
	case !raw.IsUp && raw.StatusCode == http.StatusRequestTimeout:
		return models.StatusTimeout
	case !raw.IsUp:
		return models.StatusDown
	default:
		return models.StatusUnknown
	}
}
