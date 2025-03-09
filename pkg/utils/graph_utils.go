package utils

import (
	"nitinjuyal1610/uptimeMonitor/internal/models"
	"strings"
	"time"

	"github.com/go-echarts/go-echarts/v2/opts"
)

func FormatTimeData(data []*models.ResponseTimeStat) ([]opts.LineData, []string) {
	if len(data) == 0 {
		return []opts.LineData{{Value: 0}}, []string{"No Data"}
	}
	items := make([]opts.LineData, 0, len(data))
	keys := make([]string, 0, len(data))
	for _, d := range data {
		keys = append(keys, strings.TrimSuffix(d.Date, "T00:00:00Z"))
		items = append(items, opts.LineData{
			Name:  strings.TrimSuffix(d.Date, "T00:00:00Z"),
			Value: d.AvgResponseTime * 1000,
		})
	}
	return items, keys
}

// func FormatUptimeData(data []*models.UptimeStat) ([]opts.LineData, []string) {

// 	if len(data) == 0 {
// 		return []opts.LineData{{Value: 0}}, []string{"No Data"}
// 	}
// 	items := make([]opts.LineData, 0, len(data))
// 	keys := make([]string, 0, len(data))
// 	for _, d := range data {
// 		keys = append(keys, strings.TrimSuffix(d.Date, "T00:00:00Z"))
// 		items = append(items, opts.LineData{
// 			Name:  strings.TrimSuffix(d.Date, "T00:00:00Z"),
// 			Value: d.UptimePercentage,
// 		})
// 	}
// 	return items, keys
// }

func DetermineBarColor(uptime float64) string {
	switch {
	case uptime == 0.0:
		return "bg-gray-400" // Missing data
	case uptime == 100.0:
		return "bg-green-500" // 100% uptime
	case uptime >= 75.0:
		return "bg-green-300" // 75-99%
	case uptime >= 50.0:
		return "bg-yellow-400" // 50-74%
	case uptime >= 25.0:
		return "bg-orange-400" // 25-49%
	default:
		return "bg-red-500" // 0-24%
	}
}

func FillMissingUptimeStats(existingStats []*models.UptimeStat) []*models.UptimeStat {
	dateMap := make(map[string]*models.UptimeStat)

	for _, stat := range existingStats {
		stat.Date = strings.TrimSuffix(stat.Date, "T00:00:00Z")
		dateMap[stat.Date] = stat
	}

	var filledStats []*models.UptimeStat
	for i := 29; i >= 0; i-- {
		date := time.Now().AddDate(0, 0, -i).Format("2006-01-02")

		if stat, exists := dateMap[date]; exists {
			filledStats = append(filledStats, stat)
		} else {
			filledStats = append(filledStats, &models.UptimeStat{
				Date:             date,
				UptimePercentage: 0,
				SuccessfulChecks: 0,
				FailedChecks:     0,
				BarColor:         "bg-gray-500",
			})
		}
	}

	return filledStats
}
