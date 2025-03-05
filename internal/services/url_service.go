package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptrace"
	"nitinjuyal1610/uptimeMonitor/internal/models"
	"nitinjuyal1610/uptimeMonitor/internal/repository"
	"sync"
	"time"
)

type UrlService struct {
	urlRepo    repository.UrlRepository
	statRepo   repository.StatRepository
	httpClient *http.Client
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
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
			DisableKeepAlives:   true,
		},
	}
	return &UrlService{urlRepo: urlRepo, httpClient: httpClient, statRepo: statRepo}
}

func (us *UrlService) CreateUrl(urlMonitor *models.UrlMonitors) (int, error) {
	return us.urlRepo.Create(urlMonitor)
}

func (us *UrlService) GetAllUrl(status string, keyword string) ([]*models.UrlMonitors, error) {
	return us.urlRepo.GetAll(status, keyword)
}

func (us *UrlService) GetMonitorById(id int) (*models.UrlMonitors, error) {
	return us.urlRepo.GetById(id)
}

func (us *UrlService) ProcessDueMonitorURLs() error {
	// Channels for stats and Concurrent api calls limit
	// Fetch URLs to process
	urls, err := us.urlRepo.GetDueMonitorURLs()
	if err != nil {
		return fmt.Errorf("failed to get urls : %w", err)
	}
	statChan := make(chan *RawStats, len(urls))
	limitChan := make(chan bool, 10)

	fmt.Printf("%d urls to be processed \n", len(urls))
	var wg sync.WaitGroup
	for _, url := range urls {
		wg.Add(1)
		go func(targetUrl string, monitorId int, expectedStatusCode int, collectDetailedData bool) {
			//semaphore increment
			limitChan <- true
			defer func() {
				<-limitChan
			}()
			ctx, cancel := context.WithTimeout(context.Background(), time.Duration(url.TimeoutSeconds)*time.Second)
			defer cancel()
			defer wg.Done()
			stats, _ := us.fetchStatsFromUrl(ctx, targetUrl, collectDetailedData)
			stats.MonitorId = monitorId
			stats.ExpectedStatusCode = expectedStatusCode
			statChan <- stats
		}(url.Url, url.ID, url.ExpectedStatusCode, url.CollectDetailedData)
	}

	go func() {
		wg.Wait()
		close(statChan)
		close(limitChan)
	}()

	return us.saveResultsToDB(statChan)
}

func (us *UrlService) saveResultsToDB(statChan <-chan *RawStats) error {
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
		if err := us.processBatch(batch); err != nil {
			return fmt.Errorf("failed to process batch: %w", err)
		}
	}
	return nil
}

func (us *UrlService) processBatch(batch []*RawStats) error {
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
	}

	if _, err := us.statRepo.BulkCreate(monitorChecks); err != nil {
		return fmt.Errorf("failed to save stats to DB: %w", err)
	}

	if err := us.urlRepo.BulkUpdate(monitorUpdates); err != nil {
		return fmt.Errorf("failed to update monitors: %w", err)
	}

	return nil
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

func (us *UrlService) fetchStatsFromUrl(ctx context.Context, url string, collectDetailedData bool) (*RawStats, error) {
	var (
		start        = time.Now()
		responseTime time.Duration
		ttfb         time.Duration
		contentSize  int64
		statusCode   int
	)

	method := http.MethodHead
	trace := httptrace.ClientTrace{
		GotFirstResponseByte: func() {
			ttfb = time.Since(start)
		},
	}
	if collectDetailedData {
		method = http.MethodGet
	}
	req, err := http.NewRequestWithContext(httptrace.WithClientTrace(ctx, &trace), method, url, nil)
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
		}, err
	}
	defer resp.Body.Close()

	var rawStats = &RawStats{
		RequestType: method,
	}
	if collectDetailedData {
		//copy body
		writtenBytes, err := io.Copy(io.Discard, resp.Body)
		if err != nil {
			fmt.Println(err)
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
