package service

import (
	"nitinjuyal1610/uptimeMonitor/internal/models"
	"nitinjuyal1610/uptimeMonitor/internal/repository"
)

type StatService struct {
	statRepo repository.StatRepository
}

func NewStatsService(statRepo repository.StatRepository) *StatService {
	return &StatService{statRepo}
}

func (ss *StatService) GetStats(monitorId int) ([]*models.UrlStats, error) {
	//TODO LOGIC
	return ss.statRepo.GetStatsByMonitorId(monitorId)
}
