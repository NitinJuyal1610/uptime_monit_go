package models

import "time"

type UrlMonitors struct {
	ID                 int       `json:"id"`
	Url                string    `json:"url"`
	Status             string    `json:"status"`
	FrequencyMinutes   int       `json:"frequency_minutes"`
	TimeoutSeconds     int       `json:"timeout_seconds"`
	LastChecked        time.Time `json:"last_checked"`
	ExpectedStatusCode int       `json:"expected_status_code"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}
