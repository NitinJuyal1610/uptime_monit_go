package service

import (
	"fmt"
	"nitinjuyal1610/uptimeMonitor/internal/models"
	"nitinjuyal1610/uptimeMonitor/internal/repository"
	"strings"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/go-echarts/go-echarts/v2/render"
)

type StatService struct {
	statRepo repository.StatRepository
}

func NewStatsService(statRepo repository.StatRepository) *StatService {
	return &StatService{statRepo}
}

func (ss *StatService) GetStatSummary(monitorId int) (*models.MonitorStats, error) {
	return ss.statRepo.GetStatsByMonitorId(monitorId)
}

func formatTimeData(data []*models.ResponseTimeStat) ([]opts.LineData, []string) {

	if len(data) == 0 {
		return []opts.LineData{{Value: 0}}, []string{"No Data"}
	}
	items := make([]opts.LineData, 0, len(data))
	keys := make([]string, 0, len(data))
	for _, d := range data {
		keys = append(keys, strings.TrimSuffix(d.Date, "T00:00:00Z"))
		items = append(items, opts.LineData{
			Name:  strings.TrimSuffix(d.Date, "T00:00:00Z"),
			Value: d.AvgResponseTime,
		})
	}
	return items, keys
}

func formatUptimeData(data []*models.UptimeStat) ([]opts.LineData, []string) {

	if len(data) == 0 {
		return []opts.LineData{{Value: 0}}, []string{"No Data"}
	}
	items := make([]opts.LineData, 0, len(data))
	keys := make([]string, 0, len(data))
	for _, d := range data {
		keys = append(keys, strings.TrimSuffix(d.Date, "T00:00:00Z"))
		items = append(items, opts.LineData{
			Name:  strings.TrimSuffix(d.Date, "T00:00:00Z"),
			Value: d.UptimePercentage,
		})
	}
	return items, keys
}

func (ss *StatService) CreateAvgResponseGraph(monitorId int, startDate string, endDate string) (render.ChartSnippet, error) {
	responseInfo, err := ss.statRepo.GetAvgResponseData(monitorId, startDate, endDate)

	if err != nil {
		return render.ChartSnippet{}, err
	}

	lineChart := charts.NewLine()
	lineChart.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{
			Theme:           "dark",
			Width:           "100%",
			Height:          "400px",
			BackgroundColor: "#111827",
		}),
		charts.WithTitleOpts(opts.Title{
			Title:    "Avg Response Time",
			Subtitle: fmt.Sprintf("From %s to %s", startDate, endDate),
			Left:     "center",
			TitleStyle: &opts.TextStyle{
				FontSize: 18,
				Color:    "#E5E7EB",
			},
		}),
		charts.WithTooltipOpts(opts.Tooltip{
			Trigger: "axis",

			Formatter: opts.FuncOpts(`function (params) {
				let tooltipText = '';
				params.forEach((item) => {
					tooltipText += item.marker + ' <strong>' + item.seriesName + '</strong>: ' + item.value + ' ms<br>';
				});
				return tooltipText;
			}`),
			AxisPointer: &opts.AxisPointer{
				Type: "cross",
				Label: &opts.Label{
					BackgroundColor: "#6a7985",
				},
			},
		}),
		charts.WithLegendOpts(opts.Legend{
			Show:      opts.Bool(true),
			Left:      "right",
			TextStyle: &opts.TextStyle{Color: "#E5E7EB"},
		}),
		charts.WithXAxisOpts(opts.XAxis{
			Type:      "category",
			Name:      "Date",
			NameGap:   10,
			AxisLabel: &opts.AxisLabel{Show: opts.Bool(true), Color: "#E5E7EB"},
			SplitLine: &opts.SplitLine{Show: opts.Bool(false)},
		}),
		charts.WithYAxisOpts(opts.YAxis{
			Type:      "value",
			Name:      "Response Time (ms)",
			NameGap:   30,
			Min:       0,
			SplitLine: &opts.SplitLine{Show: opts.Bool(true), LineStyle: &opts.LineStyle{Color: "#374151"}},
			AxisLabel: &opts.AxisLabel{
				Formatter: "{value} ms",
				Color:     "#E5E7EB",
			},
		}),
		charts.WithGridOpts(opts.Grid{
			Left:         "3%",
			Right:        "5%",
			Bottom:       "10%",
			Top:          "15%",
			ContainLabel: opts.Bool(true),
		}),
	)

	graphData, keys := formatTimeData(responseInfo)

	var urlStr string
	if len(responseInfo) > 0 {
		urlStr = strings.TrimPrefix(responseInfo[0].Url, "https://")
	}
	lineChart.SetXAxis(keys).AddSeries(urlStr, graphData).
		SetSeriesOptions(
			charts.WithLineChartOpts(opts.LineChart{
				Smooth: opts.Bool(true),
			}),
			charts.WithLineStyleOpts(opts.LineStyle{
				Width: 3,
				Color: "#10B981",
			}),
			charts.WithAreaStyleOpts(opts.AreaStyle{
				Color:   "#10B98133",
				Opacity: 0.4,
			}),
		)

	return lineChart.RenderSnippet(), nil
}

func (ss *StatService) CreateUptimeGraph(monitorId int, startDate string, endDate string) (render.ChartSnippet, error) {
	responseInfo, err := ss.statRepo.GetUptimeData(monitorId, startDate, endDate)

	if err != nil {
		return render.ChartSnippet{}, err
	}

	lineChart := charts.NewLine()
	lineChart.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{
			Theme:           "dark",
			Width:           "100%",
			Height:          "400px",
			BackgroundColor: "#111827",
		}),
		charts.WithTitleOpts(opts.Title{
			Title:    "Uptime Trend",
			Subtitle: fmt.Sprintf("From %s to %s", startDate, endDate),
			Left:     "center",
			TitleStyle: &opts.TextStyle{
				FontSize: 18,
				Color:    "#E5E7EB",
			},
		}),
		charts.WithTooltipOpts(opts.Tooltip{
			Trigger: "axis",
			Formatter: opts.FuncOpts(`function (params) {
				let tooltipText = '';
				params.forEach((item) => {
					tooltipText += item.marker + ' <strong>' + item.seriesName + '</strong>: ' + item.value.toFixed(2) + ' %<br>';
				});
				return tooltipText;
			}`),
			AxisPointer: &opts.AxisPointer{
				Type: "cross",
				Label: &opts.Label{
					BackgroundColor: "#6a7985",
				},
			},
		}),
		charts.WithLegendOpts(opts.Legend{
			Show:      opts.Bool(true),
			Left:      "right",
			TextStyle: &opts.TextStyle{Color: "#E5E7EB"},
		}),
		charts.WithXAxisOpts(opts.XAxis{
			Type:      "category",
			Name:      "Date",
			NameGap:   10,
			AxisLabel: &opts.AxisLabel{Show: opts.Bool(true), Color: "#E5E7EB"},
			SplitLine: &opts.SplitLine{Show: opts.Bool(false)},
		}),
		charts.WithYAxisOpts(opts.YAxis{
			Type:      "value",
			Name:      "Uptime (%)",
			NameGap:   30,
			Min:       0,
			Max:       100,
			SplitLine: &opts.SplitLine{Show: opts.Bool(true), LineStyle: &opts.LineStyle{Color: "#374151"}},
			AxisLabel: &opts.AxisLabel{
				Formatter: "{value} %",
				Color:     "#E5E7EB",
			},
		}),
		charts.WithGridOpts(opts.Grid{
			Left:         "3%",
			Right:        "5%",
			Bottom:       "10%",
			Top:          "15%",
			ContainLabel: opts.Bool(true),
		}),
	)

	graphData, keys := formatUptimeData(responseInfo)

	var urlStr string
	if len(responseInfo) > 0 {
		urlStr = strings.TrimPrefix(responseInfo[0].Url, "https://")
	}
	lineChart.SetXAxis(keys).AddSeries(urlStr, graphData).
		SetSeriesOptions(
			charts.WithLineChartOpts(opts.LineChart{
				Smooth: opts.Bool(true),
			}),
			charts.WithLineStyleOpts(opts.LineStyle{
				Width: 3,
				Color: "#10B981", // Green color for uptime
			}),
			charts.WithAreaStyleOpts(opts.AreaStyle{
				Color:   "#10B98133",
				Opacity: 0.4,
			}),
		)

	return lineChart.RenderSnippet(), nil
}

func (ss *StatService) CreateDetailedTimeGraph(monitorId int, startDate string, endDate string) (render.ChartSnippet, error) {
	responseInfo, err := ss.statRepo.GetUptimeData(monitorId, startDate, endDate)

	if err != nil {
		return render.ChartSnippet{}, err
	}

	lineChart := charts.NewLine()
	lineChart.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{
			Theme:           "dark",
			Width:           "100%",
			Height:          "400px",
			BackgroundColor: "#111827",
		}),
		charts.WithTitleOpts(opts.Title{
			Title:    "Uptime Trend",
			Subtitle: fmt.Sprintf("From %s to %s", startDate, endDate),
			Left:     "center",
			TitleStyle: &opts.TextStyle{
				FontSize: 18,
				Color:    "#E5E7EB",
			},
		}),
		charts.WithTooltipOpts(opts.Tooltip{
			Trigger: "axis",
			Formatter: opts.FuncOpts(`function (params) {
				let tooltipText = '';
				params.forEach((item) => {
					tooltipText += item.marker + ' <strong>' + item.seriesName + '</strong>: ' + item.value.toFixed(2) + ' %<br>';
				});
				return tooltipText;
			}`),
			AxisPointer: &opts.AxisPointer{
				Type: "cross",
				Label: &opts.Label{
					BackgroundColor: "#6a7985",
				},
			},
		}),
		charts.WithLegendOpts(opts.Legend{
			Show:      opts.Bool(true),
			Left:      "right",
			TextStyle: &opts.TextStyle{Color: "#E5E7EB"},
		}),
		charts.WithXAxisOpts(opts.XAxis{
			Type:      "category",
			Name:      "Date",
			NameGap:   10,
			AxisLabel: &opts.AxisLabel{Show: opts.Bool(true), Color: "#E5E7EB"},
			SplitLine: &opts.SplitLine{Show: opts.Bool(false)},
		}),
		charts.WithYAxisOpts(opts.YAxis{
			Type:      "value",
			Name:      "Uptime (%)",
			NameGap:   30,
			Min:       0,
			Max:       100,
			SplitLine: &opts.SplitLine{Show: opts.Bool(true), LineStyle: &opts.LineStyle{Color: "#374151"}},
			AxisLabel: &opts.AxisLabel{
				Formatter: "{value} %",
				Color:     "#E5E7EB",
			},
		}),
		charts.WithGridOpts(opts.Grid{
			Left:         "3%",
			Right:        "5%",
			Bottom:       "10%",
			Top:          "15%",
			ContainLabel: opts.Bool(true),
		}),
	)

	graphData, keys := formatUptimeData(responseInfo)

	var urlStr string
	if len(responseInfo) > 0 {
		urlStr = strings.TrimPrefix(responseInfo[0].Url, "https://")
	}
	lineChart.SetXAxis(keys).AddSeries(urlStr, graphData).
		SetSeriesOptions(
			charts.WithLineChartOpts(opts.LineChart{
				Smooth: opts.Bool(true),
			}),
			charts.WithLineStyleOpts(opts.LineStyle{
				Width: 3,
				Color: "#10B981", // Green color for uptime
			}),
			charts.WithAreaStyleOpts(opts.AreaStyle{
				Color:   "#10B98133",
				Opacity: 0.4,
			}),
		)

	return lineChart.RenderSnippet(), nil
}
