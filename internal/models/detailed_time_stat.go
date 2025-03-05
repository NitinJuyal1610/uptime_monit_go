package models

import "time"

type DetailedTimeStat struct {
	Timestamp    time.Time `json:"timestamp"`
	Ttfb         float64   `json:"ttfb"`
	Url          string    `json:"url"`
	ResponseTime float64   `json:"response_time"`
	ContentSize  int64     `json:"content_size"`
	RequestType  string    `json:"request_type"`
}
