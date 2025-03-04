package handler

import (
	"fmt"
	"net/http"
	"net/url"
	"nitinjuyal1610/uptimeMonitor/internal/models"
	service "nitinjuyal1610/uptimeMonitor/internal/services"
	"nitinjuyal1610/uptimeMonitor/pkg/utils"
	templates "nitinjuyal1610/uptimeMonitor/web"
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
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

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
	}

	entityId, err := uh.urlService.CreateUrl(urlMonitor)
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

	//get query param for status
	status := r.URL.Query().Get("status")
	keyword := r.URL.Query().Get("q")
	values, err := uh.urlService.GetAllUrl(status, keyword)

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

	monitor, err := uh.urlService.GetMonitorById(id)

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
