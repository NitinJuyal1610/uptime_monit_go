package handler

import (
	"fmt"
	"html/template"
	"net/http"
	"nitinjuyal1610/uptimeMonitor/internal/models"
	service "nitinjuyal1610/uptimeMonitor/internal/services"
	"nitinjuyal1610/uptimeMonitor/pkg/utils"
	templates "nitinjuyal1610/uptimeMonitor/web"
	"strconv"
	"time"

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

type PageData struct {
	UptimeStats []*models.UptimeStat
}

func (s *StatHandler) GetMonitorStats(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	monitorId, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}
	result, err := s.statService.GetStatSummary(monitorId)
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
			AvgResponseTime:       result.AvgResponseTime * 1000,
			UptimePercentage:      result.UptimePercentage,
			DailyUptimePercentage: result.DailyUptimePercentage,
			DailyAvgResponseTime:  result.DailyAvgResponseTime * 1000,
		}
	}

	s.templateManager.Render(w, "stat-summary.html", res)
}

func (s *StatHandler) GetAvgResponseGraph(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")
	dayStr := r.URL.Query().Get("days")

	monitorId, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid monitor ID", http.StatusBadRequest)
		return
	}

	now := time.Now()
	var days int

	if dayStr != "" {
		days, err = strconv.Atoi(dayStr)
		if err != nil {
			http.Error(w, "Invalid days parameter", http.StatusBadRequest)
			return
		}
	}

	if startDate == "" || endDate == "" {
		endDate = now.Format("2006-01-02")
		startDate = now.AddDate(0, 0, -30).Format("2006-01-02")
	}

	if dayStr != "" {
		startDate = now.AddDate(0, 0, -days).Format("2006-01-02")
	}

	lineSnippet, err := s.statService.CreateAvgResponseGraph(monitorId, startDate, endDate)
	if err != nil {
		errStr := fmt.Sprintf("Failed to fetch average response graph %v", err)
		http.Error(w, errStr, http.StatusBadRequest)
		return
	}

	data := struct {
		Element template.HTML
		Script  template.HTML
	}{
		Element: template.HTML(lineSnippet.Element),
		Script:  template.HTML(lineSnippet.Script),
	}

	s.templateManager.Render(w, "chart-container.html", data)
}

func (s *StatHandler) GetUptimeGraph(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")
	dayStr := r.URL.Query().Get("days")

	monitorId, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid monitor ID", http.StatusBadRequest)
		return
	}

	now := time.Now()
	var days int

	if dayStr != "" {
		days, err = strconv.Atoi(dayStr)
		if err != nil {
			http.Error(w, "Invalid days parameter", http.StatusBadRequest)
			return
		}
	}

	if startDate == "" || endDate == "" {
		endDate = now.Format("2006-01-02")
		startDate = now.AddDate(0, 0, -30).Format("2006-01-02")
	}

	if dayStr != "" {
		startDate = now.AddDate(0, 0, -days).Format("2006-01-02")
	}

	uptimeTrend, err := s.statService.CreateUptimeTrend(monitorId, startDate, endDate)
	if err != nil {
		errStr := fmt.Sprintf("Failed to fetch uptime trend %v", err)
		http.Error(w, errStr, http.StatusBadRequest)
		return
	}

	data := PageData{
		UptimeStats: uptimeTrend,
	}
	s.templateManager.Render(w, "uptime-trend.html", data)
}

func (s *StatHandler) GetDetailedTimeGraph(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")
	dayStr := r.URL.Query().Get("days")

	monitorId, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid monitor ID", http.StatusBadRequest)
		return
	}

	now := time.Now()
	var days int

	if dayStr != "" {
		days, err = strconv.Atoi(dayStr)
		if err != nil {
			http.Error(w, "Invalid days parameter", http.StatusBadRequest)
			return
		}
	}

	if startDate == "" || endDate == "" {
		endDate = now.Format("2006-01-02")
		startDate = now.AddDate(0, 0, -30).Format("2006-01-02")
	}

	if dayStr != "" {
		startDate = now.AddDate(0, 0, -days).Format("2006-01-02")
	}

	lineSnippet, err := s.statService.CreateDetailedTimeGraph(monitorId, startDate, endDate)
	if err != nil {
		errStr := fmt.Sprintf("Failed to fetch detailed time graph %v", err)
		http.Error(w, errStr, http.StatusBadRequest)
		return
	}

	data := struct {
		Element template.HTML
		Script  template.HTML
	}{
		Element: template.HTML(lineSnippet.Element),
		Script:  template.HTML(lineSnippet.Script),
	}

	s.templateManager.Render(w, "chart-container.html", data)
}
