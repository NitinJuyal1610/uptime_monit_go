package models

type UptimeStat struct {
	Date             string  `json:"date"`
	Url              string  `json:"url"`
	UptimePercentage float64 `json:"uptime_percentage"`
	SuccessfulChecks int     `json:"successful_checks"`
	FailedChecks     int     `json:"failed_checks"`
}
