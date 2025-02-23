package models

import "time"

type MonitorCheck struct {
	ID           int       `json:"id"`
	MonitorId    int       `json:"monitor_id"`
	StatusCode   int       `json:"status_code"`
	ResponseTime float64   `json:"response_time"`
	IsUp         bool      `json:"is_up"`
	Timestamp    time.Time `json:"timestamp"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
