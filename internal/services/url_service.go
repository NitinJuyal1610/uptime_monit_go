package service

import (
	"nitinjuyal1610/uptimeMonitor/internal/models"
	url "nitinjuyal1610/uptimeMonitor/internal/repository"
)

type UrlService struct {
	urlRepo url.UrlRepository
}

func NewUrlService(urlRepo url.UrlRepository) *UrlService {
	return &UrlService{urlRepo: urlRepo}
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
	// fetch urls to process

	//spawn workers

	//worker processing logic in different folder

	return nil
}
