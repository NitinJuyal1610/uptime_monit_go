package handler

import (
	"fmt"
	"net/http"
	service "nitinjuyal1610/uptimeMonitor/internal/services"
	"nitinjuyal1610/uptimeMonitor/pkg/utils"
	templates "nitinjuyal1610/uptimeMonitor/web"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type StatHandler struct {
	statService     *service.StatService
	templateManager *templates.TemplateManager
}

func NewStatHandler(s *service.StatService, tm *templates.TemplateManager) *StatHandler {
	return &StatHandler{
		statService:     s,
		templateManager: tm,
	}
}

type MonitorStatsResponse struct {
	Url                   string  `json:"url"`
	Status                string  `json:"status"`
	LastChecked           string  `json:"last_checked"`
	TotalChecks           int     `json:"total_checks"`
	SuccessfulChecks      int     `json:"successful_checks"`
	FailedChecks          int     `json:"failed_checks"`
	TotalIncidents        int     `json:"total_incidnets"`
	AvgResponseTime       float64 `json:"avg_response_time"`
	UptimePercentage      float64 `json:"uptime_percentage"`
	DailyUptimePercentage float64 `json:"daily_uptime_percentage"`
	DailyAvgResponseTime  float64 `json:"daily_avg_response_time"`
}

func (s *StatHandler) GetMonitorStats(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	monitorId, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}
	result, err := s.statService.GetStats(monitorId)
	if err != nil {
		errStr := fmt.Sprintf("Failed to fetch monitor stats %v", err)
		http.Error(w, errStr, http.StatusBadRequest)
		return
	}

	var res MonitorStatsResponse
	if result != nil {
		res = MonitorStatsResponse{
			Url:                   result.Url,
			Status:                result.Status,
			LastChecked:           utils.TimeAgo(result.LastChecked),
			TotalChecks:           result.TotalChecks,
			SuccessfulChecks:      result.SuccessfulChecks,
			FailedChecks:          result.FailedChecks,
			TotalIncidents:        result.FailedChecks,
			AvgResponseTime:       result.AvgResponseTime,
			UptimePercentage:      result.UptimePercentage,
			DailyUptimePercentage: result.DailyUptimePercentage,
			DailyAvgResponseTime:  result.DailyAvgResponseTime,
		}
	}

	s.templateManager.Render(w, "stat-summary.html", res)

}
