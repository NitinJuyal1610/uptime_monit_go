package service

import (
	"nitinjuyal1610/uptimeMonitor/internal/repository"

	"github.com/jackc/pgx/v5/pgxpool"
)

// hold all services
type Services struct {
	UrlService  *UrlService
	StatService *StatService
	AuthService *AuthService
}

func NewServices(db *pgxpool.Pool) *Services {

	urlRepo := repository.NewUrlRepository(db)
	statRepo := repository.NewStatRepository(db)
	userRepo := repository.NewUserRepository(db)
	urlService := NewUrlService(urlRepo, statRepo)
	statService := NewStatsService(statRepo)
	authService := NewAuthService(userRepo)

	return &Services{
		UrlService:  urlService,
		StatService: statService,
		AuthService: authService,
	}
}
