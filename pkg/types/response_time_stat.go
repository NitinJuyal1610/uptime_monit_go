package types

type ResponseTimeStat struct {
	Date            string  `json:"date"`
	MonitorID       int     `json:"monitor_id"`
	Url             string  `json:"url"`
	AvgResponseTime float64 `json:"avg_response_time"`
}
