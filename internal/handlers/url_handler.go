package handler

import (
	"fmt"
	"net/http"
	"net/url"
	middleware "nitinjuyal1610/uptimeMonitor/internal/middlewares"
	"nitinjuyal1610/uptimeMonitor/internal/models"
	service "nitinjuyal1610/uptimeMonitor/internal/services"
	"nitinjuyal1610/uptimeMonitor/pkg/utils"
	templates "nitinjuyal1610/uptimeMonitor/web"
	"slices"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
)

type UrlHandler struct {
	urlService      *service.UrlService
	templateManager *templates.TemplateManager
}

type CreateURLRequest struct {
	Url                string    `json:"url"`
	Status             string    `json:"status,omitempty"`
	FrequencyMinutes   int       `json:"frequency_minutes,omitempty"`
	TimeoutSeconds     int       `json:"timeout_seconds,omitempty"`
	LastChecked        time.Time `json:"last_checked,omitempty"`
	ExpectedStatusCode int       `json:"expected_status_code,omitempty"`
}

func NewUrlHandler(u *service.UrlService, tm *templates.TemplateManager) *UrlHandler {
	return &UrlHandler{
		urlService:      u,
		templateManager: tm,
	}
}

func validateURL(inputURL string) bool {
	parsedURL, err := url.ParseRequestURI(inputURL)
	return err == nil && parsedURL.Scheme != "" && parsedURL.Host != ""
}

func (uh *UrlHandler) CreateURLMonitor(w http.ResponseWriter, r *http.Request) {

	userId := r.Context().Value(middleware.UserKey)
	// Parse form data
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	// Extract form values
	url := r.FormValue("url")
	frequencyMinutes, _ := strconv.Atoi(r.FormValue("frequency_minutes"))
	expectedStatusCode, _ := strconv.Atoi(r.FormValue("status_code"))
	timeoutSeconds, _ := strconv.Atoi(r.FormValue("timeout_seconds"))
	collectDetailedDate, _ := strconv.ParseBool(r.FormValue("collect_detailed_data"))
	maxFailThreshold, _ := strconv.Atoi(r.FormValue("max_fail_threshold"))
	alertEmail := r.FormValue("alert_email")
	if url == "" {
		http.Error(w, "URL cannot be empty", http.StatusBadRequest)
		return
	}
	// Validate URL
	if !validateURL(url) {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}
	// Validate Frequency
	if frequencyMinutes <= 0 {
		http.Error(w, "Frequency must be greater than zero", http.StatusBadRequest)
		return
	}

	if timeoutSeconds <= 0 {
		http.Error(w, "Timeout must be greater than 0", http.StatusBadRequest)
		return
	}

	// Create URL monitor object
	urlMonitor := &models.UrlMonitors{
		Url:                 url,
		FrequencyMinutes:    frequencyMinutes,
		LastChecked:         time.Now().UTC().Truncate(time.Minute),
		ExpectedStatusCode:  expectedStatusCode,
		Status:              models.StatusUnknown,
		TimeoutSeconds:      timeoutSeconds,
		CollectDetailedData: collectDetailedDate || false,
		MaxFailThreshold:    maxFailThreshold,
		AlertEmail:          alertEmail,
		UserId:              userId.(int),
	}

	entityId, err := uh.urlService.CreateUrl(r.Context(), urlMonitor)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create URL: %v", err), http.StatusInternalServerError)
		return
	}

	fmt.Println("Successfully created a monitor with id: ", entityId)

	urlMonitor.ID = entityId
	data := map[string]any{
		"Monitor": urlMonitor,
	}
	uh.templateManager.Render(w, "monitor-item", data)
}

func (uh *UrlHandler) GetURLMonitors(w http.ResponseWriter, r *http.Request) {

	userId := r.Context().Value(middleware.UserKey).(int)

	//get query param for status
	status := r.URL.Query().Get("status")
	keyword := r.URL.Query().Get("q")
	values, err := uh.urlService.GetAllUrl(r.Context(), status, keyword, userId)

	if err != nil {
		errStr := fmt.Sprintf("Failed to fetch URL Monitors %v", err)
		http.Error(w, errStr, http.StatusBadRequest)
		return
	}

	uh.templateManager.Render(w, "monitor-list.html", values)
}

func (uh *UrlHandler) GetMonitorById(w http.ResponseWriter, r *http.Request) {
	// Extract ID from query parameters
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	monitor, err := uh.urlService.GetMonitorById(r.Context(), id)

	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to fetch info: %v", err), http.StatusBadRequest)
		return
	}

	res := utils.JSONResponse{
		Message: "Monitor Fetched Successfully.",
		Data: map[string]any{
			"monitor": monitor,
		},
	}
	utils.SendJSONResponse(w, http.StatusOK, res)
}

func (uh *UrlHandler) UpdateMonitorStatus(w http.ResponseWriter, r *http.Request) {

	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	status := r.FormValue("status")
	validStatus := []string{"PAUSED", "UNKNOWN"}

	if !slices.Contains(validStatus, status) {
		http.Error(w, "Invalid Status", http.StatusBadRequest)
		return
	}
	err = uh.urlService.UpdateMonitorStatus(r.Context(), id, status)

	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to Update monitor status: %v", err), http.StatusBadRequest)
		return
	}
	uh.templateManager.Render(w, "service-status.html", map[string]any{
		"Status":    status,
		"MonitorId": id,
	})

}
