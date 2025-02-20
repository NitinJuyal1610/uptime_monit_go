package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
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
	IsUp               bool          `json:"is_up"`
	MonitorId          int           `json:"monitor_id"`
	ExpectedStatusCode int           `json:"expected_status_code"`
}

func NewUrlService(urlRepo repository.UrlRepository, statRepo repository.StatRepository) *UrlService {
	httpClient := &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
		},
		Timeout: 10 * time.Second,
	}
	return &UrlService{urlRepo: urlRepo, httpClient: httpClient, statRepo: statRepo}
}

func (us *UrlService) CreateUrl(urlMonitor *models.UrlMonitors) (int, error) {
	return us.urlRepo.Create(urlMonitor)
}

func (us *UrlService) GetAllUrl(status string) ([]*models.UrlMonitors, error) {
	return us.urlRepo.GetAll(status)
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
		go func(targetUrl string, monitorId int, expectedStatusCode int) {
			//semaphore increment
			limitChan <- true
			defer func() {
				<-limitChan
			}()
			ctx, cancel := context.WithTimeout(context.Background(), time.Duration(url.TimeoutSeconds)*time.Second)
			defer cancel()
			defer wg.Done()
			stats, _ := us.fetchStatsFromUrl(ctx, targetUrl)
			stats.MonitorId = monitorId
			stats.ExpectedStatusCode = expectedStatusCode
			statChan <- stats
		}(url.Url, url.ID, url.ExpectedStatusCode)
	}

	go func() {
		wg.Wait()
		close(statChan)
		close(limitChan)
	}()

	return us.saveResultsToDB(statChan)
}

func (us *UrlService) saveResultsToDB(statChan <-chan *RawStats) error {

	var allStats []*RawStats

	for s := range statChan {
		allStats = append(allStats, s)
	}

	if len(allStats) == 0 {
		return nil
	}

	urlStats := make([]*models.UrlStats, len(allStats))
	for i, raw := range allStats {

		var status models.Status
		switch {
		case raw.IsUp && raw.StatusCode == raw.ExpectedStatusCode:
			status = models.StatusUp
		case raw.IsUp && raw.StatusCode != raw.ExpectedStatusCode:
			status = models.StatusError
		case !raw.IsUp && raw.StatusCode == http.StatusRequestTimeout:
			status = models.StatusTimeout
		case !raw.IsUp:
			status = models.StatusDown
		default:
			status = models.StatusUnknown
		}

		urlStats[i] = &models.UrlStats{
			MonitorId:    raw.MonitorId,
			StatusCode:   raw.StatusCode,
			ResponseTime: raw.ResponseTime,
			IsUp:         raw.IsUp,
		}

		//update last
		err := us.urlRepo.Update(urlStats[i].MonitorId, &models.UrlMonitors{
			LastChecked: time.Now().UTC().Truncate(time.Minute),
			Status:      status,
		})
		if err != nil {
			return fmt.Errorf("failed to update timestamp: %v", err)
		}
	}

	_, err := us.statRepo.BulkCreate(urlStats)
	if err != nil {
		return fmt.Errorf("failed to save stats to DB: %w", err)
	}

	return nil
}

func (us *UrlService) fetchStatsFromUrl(ctx context.Context, url string) (*RawStats, error) {
	var (
		start        = time.Now()
		responseTime time.Duration
		statusCode   int
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodHead, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := us.httpClient.Do(req)
	responseTime = time.Since(start)

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
			StatusCode:   statusCode,
			ResponseTime: responseTime,
			IsUp:         false,
		}, err
	}

	defer resp.Body.Close()

	statusCode = resp.StatusCode

	rawStats := &RawStats{
		StatusCode:   statusCode,
		ResponseTime: responseTime,
		IsUp:         statusCode >= 200 && statusCode < 400,
	}

	fmt.Printf("Raw stats %s -> %+v\n", url, rawStats)
	return rawStats, nil
}
