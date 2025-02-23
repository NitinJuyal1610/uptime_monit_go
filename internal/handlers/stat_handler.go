package handler

import (
	"fmt"
	"net/http"
	service "nitinjuyal1610/uptimeMonitor/internal/services"
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

func (s *StatHandler) GetUrlStats(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	monitorId, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}
	_, err = s.statService.GetStats(monitorId)
	if err != nil {
		errStr := fmt.Sprintf("Failed to fetch monitor stats %v", err)
		http.Error(w, errStr, http.StatusBadRequest)
		return
	}

	s.templateManager.Render(w, "stat-summary.html", struct{ MonitorId int }{MonitorId: monitorId})

}
