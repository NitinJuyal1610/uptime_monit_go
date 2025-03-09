package service

import (
	"fmt"
	"nitinjuyal1610/uptimeMonitor/internal/models"
	"nitinjuyal1610/uptimeMonitor/internal/repository"
	"nitinjuyal1610/uptimeMonitor/pkg/utils"
	"strings"
	"time"

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
			Type:  "value",
			Name:  "Time (ms)",
			Min:   0,
			Scale: opts.Bool(true),
			AxisLabel: &opts.AxisLabel{
				Formatter: "{value} ms",
				Color:     "#E5E7EB",
			},
			SplitLine: &opts.SplitLine{Show: opts.Bool(true), LineStyle: &opts.LineStyle{Color: "#374151"}},
		}),
		charts.WithGridOpts(opts.Grid{
			Left:         "5%",
			Right:        "10%",
			Bottom:       "15%",
			Top:          "20%",
			ContainLabel: opts.Bool(true),
		}),
	)

	graphData, keys := utils.FormatTimeData(responseInfo)

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

func (ss *StatService) CreateUptimeTrend(monitorId int, startDate string, endDate string) ([]*models.UptimeStat, error) {
	responseInfo, err := ss.statRepo.GetUptimeData(monitorId, startDate, endDate)

	if err != nil {
		return nil, err
	}
	for i := range responseInfo {
		responseInfo[i].BarColor = utils.DetermineBarColor(responseInfo[i].UptimePercentage)
	}
	return utils.FillMissingUptimeStats(responseInfo), nil
	// if err != nil {
	// 	return render.ChartSnippet{}, err
	// }

	// lineChart := charts.NewLine()
	// lineChart.SetGlobalOptions(
	// 	charts.WithInitializationOpts(opts.Initialization{
	// 		Theme:           "dark",
	// 		Width:           "100%",
	// 		Height:          "400px",
	// 		BackgroundColor: "#111827",
	// 	}),
	// 	charts.WithTitleOpts(opts.Title{
	// 		Title:    "Uptime Trend",
	// 		Subtitle: fmt.Sprintf("From %s to %s", startDate, endDate),
	// 		Left:     "center",
	// 		TitleStyle: &opts.TextStyle{
	// 			FontSize: 18,
	// 			Color:    "#E5E7EB",
	// 		},
	// 	}),
	// 	charts.WithTooltipOpts(opts.Tooltip{
	// 		Trigger: "axis",
	// 		Formatter: opts.FuncOpts(`function (params) {
	// 			let tooltipText = '';
	// 			params.forEach((item) => {
	// 				tooltipText += item.marker + ' <strong>' + item.seriesName + '</strong>: ' + item.value.toFixed(2) + ' %<br>';
	// 			});
	// 			return tooltipText;
	// 		}`),
	// 		AxisPointer: &opts.AxisPointer{
	// 			Type: "cross",
	// 			Label: &opts.Label{
	// 				BackgroundColor: "#6a7985",
	// 			},
	// 		},
	// 	}),
	// 	charts.WithLegendOpts(opts.Legend{
	// 		Show:      opts.Bool(true),
	// 		Left:      "right",
	// 		TextStyle: &opts.TextStyle{Color: "#E5E7EB"},
	// 	}),
	// 	charts.WithXAxisOpts(opts.XAxis{
	// 		Type:      "category",
	// 		Name:      "Date",
	// 		NameGap:   10,
	// 		AxisLabel: &opts.AxisLabel{Show: opts.Bool(true), Color: "#E5E7EB"},
	// 		SplitLine: &opts.SplitLine{Show: opts.Bool(false)},
	// 	}),
	// 	charts.WithYAxisOpts(opts.YAxis{
	// 		Type:      "value",
	// 		Name:      "Uptime (%)",
	// 		NameGap:   30,
	// 		Min:       0,
	// 		Max:       100,
	// 		SplitLine: &opts.SplitLine{Show: opts.Bool(true), LineStyle: &opts.LineStyle{Color: "#374151"}},
	// 		AxisLabel: &opts.AxisLabel{
	// 			Formatter: "{value} %",
	// 			Color:     "#E5E7EB",
	// 		},
	// 	}),
	// 	charts.WithGridOpts(opts.Grid{
	// 		Left:         "3%",
	// 		Right:        "5%",
	// 		Bottom:       "10%",
	// 		Top:          "15%",
	// 		ContainLabel: opts.Bool(true),
	// 	}),
	// )

	// graphData, keys := formatUptimeData(responseInfo)

	// var urlStr string
	// if len(responseInfo) > 0 {
	// 	urlStr = strings.TrimPrefix(responseInfo[0].Url, "https://")
	// }
	// lineChart.SetXAxis(keys).AddSeries(urlStr, graphData).
	// 	SetSeriesOptions(
	// 		charts.WithLineChartOpts(opts.LineChart{
	// 			Smooth: opts.Bool(true),
	// 		}),
	// 		charts.WithLineStyleOpts(opts.LineStyle{
	// 			Width: 3,
	// 			Color: "#10B981", // Green color for uptime
	// 		}),
	// 		charts.WithAreaStyleOpts(opts.AreaStyle{
	// 			Color:   "#10B98133",
	// 			Opacity: 0.4,
	// 		}),
	// 	)

	// return lineChart.RenderSnippet(), nil
}

func (ss *StatService) CreateDetailedTimeGraph(monitorId int, startDate string, endDate string) (render.ChartSnippet, error) {
	responseInfo, err := ss.statRepo.GetDetailedTimeData(monitorId, startDate, endDate)
	if err != nil {
		return render.ChartSnippet{}, err
	}

	if len(responseInfo) == 0 {
		return render.ChartSnippet{}, nil
	}

	var (
		keys             []string
		ttfbData         []opts.LineData
		responseTimeData []opts.LineData
	)

	for _, stat := range responseInfo {
		keys = append(keys, stat.Timestamp.In(time.FixedZone("IST", 19800)).Format("Jan 02 3:04 PM"))
		ttfbData = append(ttfbData, opts.LineData{Value: stat.Ttfb * 1000})
		responseTimeData = append(responseTimeData, opts.LineData{Value: stat.ResponseTime * 1000})
	}

	lineChart := charts.NewLine()
	lineChart.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{
			Theme:           "dark",
			Width:           "100%",
			Height:          "500px",
			BackgroundColor: "#111827",
		}),
		charts.WithTitleOpts(opts.Title{
			Title:    "Performance Metrics",
			Subtitle: fmt.Sprintf("TTFB and Response Time (%s to %s)", startDate, endDate),
			Left:     "center",
			TitleStyle: &opts.TextStyle{
				FontSize: 18,
				Color:    "#E5E7EB",
			},
		}),
		charts.WithTooltipOpts(opts.Tooltip{
			Trigger: "axis",
			Formatter: opts.FuncOpts(`function (params) {
                let tooltipText = params[0].axisValue + '<br/>';
                params.forEach((item) => {
                    tooltipText += item.marker + ' <strong>' + item.seriesName + '</strong>: ' + 
                                  item.value.toFixed(2) + ' ms<br>';
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
			Type:    "category",
			Name:    "Timestamp",
			NameGap: 10,
			AxisLabel: &opts.AxisLabel{
				Show:     opts.Bool(true),
				Color:    "#E5E7EB",
				Rotate:   0,      // Remove rotation for better readability
				Interval: "auto", // Let the chart automatically determine label intervals
				Formatter: opts.FuncOpts(`function(value) {
                    return value; 
                }`),
			},
			SplitLine: &opts.SplitLine{Show: opts.Bool(false)},
			AxisTick: &opts.AxisTick{
				Show:           opts.Bool(true),
				AlignWithLabel: opts.Bool(true),
			},
		}),
		charts.WithYAxisOpts(opts.YAxis{
			Type:  "value",
			Name:  "Time (ms)",
			Min:   0,
			Scale: opts.Bool(true),
			AxisLabel: &opts.AxisLabel{
				Formatter: "{value} ms",
				Color:     "#E5E7EB",
			},
			SplitLine: &opts.SplitLine{Show: opts.Bool(true), LineStyle: &opts.LineStyle{Color: "#374151"}},
		}),
		charts.WithGridOpts(opts.Grid{
			Left:         "5%",
			Right:        "10%",
			Bottom:       "15%",
			Top:          "20%",
			ContainLabel: opts.Bool(true),
		}),
		// Add data zooming capability
		charts.WithDataZoomOpts(opts.DataZoom{
			Type:       "slider",
			Start:      0,
			End:        100,
			XAxisIndex: []int{0},
		}),
	)

	lineChart.SetXAxis(keys).
		AddSeries("TTFB", ttfbData,
			charts.WithLineChartOpts(opts.LineChart{Smooth: opts.Bool(true)}),
			charts.WithLineStyleOpts(opts.LineStyle{
				Width: 3,
				Color: "#10B981",
			}),
			charts.WithAreaStyleOpts(opts.AreaStyle{
				Color:   "#10B98133",
				Opacity: 0.4,
			}),
		).
		AddSeries("Response Time", responseTimeData,
			charts.WithLineChartOpts(opts.LineChart{Smooth: opts.Bool(true)}),
			charts.WithLineStyleOpts(opts.LineStyle{
				Width: 3,
				Color: "#3B82F6",
			}),
			charts.WithAreaStyleOpts(opts.AreaStyle{
				Color:   "#3B82F633",
				Opacity: 0.4,
			}),
		)

	return lineChart.RenderSnippet(), nil
}
