package models

type NotifyAlert struct {
	MonitorId        int
	AlertEmail       string
	ConsecutiveFails int
	Url              string
}
