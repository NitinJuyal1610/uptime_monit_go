package models

import "time"

type Status string

const (
	StatusUnknown  Status = "UNKNOWN"
	StatusUp       Status = "UP"
	StatusDown     Status = "DOWN"
	StatusSlow     Status = "SLOW"
	StatusTimeout  Status = "TIMEOUT"
	StatusError    Status = "ERROR"
	StatusSSLError Status = "SSL_ERROR"
	StatusDNSFail  Status = "DNS_FAILURE"
	StatusPending  Status = "PENDING"
	StatusPaused   Status = "PAUSED"
	StatusDeleted  Status = "DELETED"
)

type UrlMonitors struct {
	ID                  int       `json:"id"`
	Url                 string    `json:"url"`
	Status              Status    `json:"status"`
	FrequencyMinutes    int       `json:"frequency_minutes"`
	TimeoutSeconds      int       `json:"timeout_seconds"`
	LastChecked         time.Time `json:"last_checked"`
	ExpectedStatusCode  int       `json:"expected_status_code"`
	CollectDetailedData bool      `json:"collect_detailed_data"`
	MaxFailThreshold    int       `json:"max_fail_threshold"`
	ConsecutiveFails    int       `json:"consecutive_fails"`
	AlertEmail          string    `json:"alert_email"`
	UserId              int       `json:"user_id"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}
