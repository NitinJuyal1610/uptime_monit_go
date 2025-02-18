package service

import (
	"nitinjuyal1610/uptimeMonitor/internal/repository"
)

type StatService struct {
	statRepo repository.StatRepository
}

func NewStatsService(statRepo repository.StatRepository) *StatService {
	return &StatService{statRepo}
}
