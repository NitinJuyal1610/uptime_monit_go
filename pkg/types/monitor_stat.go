package types

import "time"

type MonitorStats struct {
	ID                    int       `json:"id"`
	Url                   string    `json:"url"`
	Status                string    `json:"status"`
	LastChecked           time.Time `json:"last_checked"`
	TotalChecks           int       `json:"total_checks"`
	SuccessfulChecks      int       `json:"successful_checks"`
	FailedChecks          int       `json:"failed_checks"`
	AvgResponseTime       float64   `json:"avg_response_time"`
	UptimePercentage      float64   `json:"uptime_percentage"`
	DailyUptimePercentage float64   `json:"daily_uptime_percentage"`
	DailyAvgResponseTime  float64   `json:"daily_avg_response_time"`
}
