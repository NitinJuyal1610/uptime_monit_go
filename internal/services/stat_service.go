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

func (ss *StatService) GetStatSummary(monitorId int) (*models.MonitorStats, error) {
	return ss.statRepo.GetStatsByMonitorId(monitorId)
}

func (ss *StatService) GetAvgResponseData(monitorId int, startDate string, endDate string) ([]*models.ResponseTimeStat, error) {
	return ss.statRepo.GetResponseTimeByDateRange(monitorId, startDate, endDate)
}
