package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"nitinjuyal1610/uptimeMonitor/internal/models"
	service "nitinjuyal1610/uptimeMonitor/internal/services"
	"nitinjuyal1610/uptimeMonitor/pkg/utils"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
)

type UrlHandler struct {
	urlService *service.UrlService
}

type CreateURLRequest struct {
	Url                string    `json:"url"`
	Status             string    `json:"status,omitempty"`
	FrequencyMinutes   int       `json:"frequency_minutes,omitempty"`
	TimeoutSeconds     int       `json:"timeout_seconds,omitempty"`
	LastChecked        time.Time `json:"last_checked,omitempty"`
	ExpectedStatusCode int       `json:"expected_status_code,omitempty"`
}

func NewUrlHandler(u *service.UrlService) *UrlHandler {
	return &UrlHandler{
		urlService: u,
	}
}

func validateURL(inputURL string) bool {
	parsedURL, err := url.ParseRequestURI(inputURL)
	return err == nil && parsedURL.Scheme != "" && parsedURL.Host != ""
}

func (uh *UrlHandler) CreateURLMonitor(w http.ResponseWriter, r *http.Request) {

	var req CreateURLRequest
	//decode body
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Failed to parse request: %v", err), http.StatusBadRequest)
		return
	}

	// Non Empty URL
	if req.Url == "" {
		http.Error(w, "URL cannot be empty", http.StatusBadRequest)
		return
	}

	// Validate URL
	if !validateURL(req.Url) {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	// Validate Frequency
	if req.FrequencyMinutes <= 0 {
		http.Error(w, "Frequency cannot be empty", http.StatusBadRequest)
		return
	}

	urlMonitor := &models.UrlMonitors{
		Url:                req.Url,
		FrequencyMinutes:   req.FrequencyMinutes,
		LastChecked:        time.Now().UTC().Truncate(time.Minute),
		ExpectedStatusCode: req.ExpectedStatusCode | 200,
		Status:             "ACTIVE",
		TimeoutSeconds:     req.TimeoutSeconds | 5,
	}

	// Create URL using service
	entityId, err := uh.urlService.CreateUrl(urlMonitor)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create URL: %v", err), http.StatusInternalServerError)
		return
	}

	res := utils.JSONResponse{
		Message: "Created url entry successfully!",
		Data:    map[string]any{"entityId": entityId},
	}

	utils.SendJSONResponse(w, http.StatusCreated, res)
}

func (uh *UrlHandler) GetURLMonitors(w http.ResponseWriter, r *http.Request) {

	//get query param for status
	status := r.URL.Query().Get("status")
	values, err := uh.urlService.GetAllUrl(status)

	if err != nil {
		errStr := fmt.Sprintf("Failed to fetch URL Monitors %v", err)
		http.Error(w, errStr, http.StatusBadRequest)
	}

	res := utils.JSONResponse{
		Message: "Fetch URL monitors successfully",
		Data:    map[string]any{"monitors": values},
	}

	utils.SendJSONResponse(w, http.StatusAccepted, res)
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
