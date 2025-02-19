package service

import (
	"database/sql"
	"nitinjuyal1610/uptimeMonitor/internal/repository"
)

// hold all services
type Services struct {
	UrlService *UrlService
}

func NewServices(db *sql.DB) *Services {

	urlRepo := repository.NewUrlRepository(db)
	statRepo := repository.NewStatRepository(db)
	urlService := NewUrlService(urlRepo, statRepo)

	return &Services{
		UrlService: urlService,
	}
}
