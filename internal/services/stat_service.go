package service

import (
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

func (ss *StatService) GetAvgResponseGraph(monitorId int, startDate string, endDate string) (render.ChartSnippet, error) {
	responseInfo, err := ss.statRepo.GetResponseTimeByDateRange(monitorId, startDate, endDate)

	if err != nil {
		return render.ChartSnippet{}, err
	}

	lineChart := charts.NewLine()
	lineChart.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{
			Theme:  "dark",
			Width:  "900px",
			Height: "500px",
		}),
		charts.WithTitleOpts(opts.Title{
			Title:    "API Response Time Trend",
			Subtitle: "Last 30 Days",
			Left:     "center",
			TitleStyle: &opts.TextStyle{
				FontSize: 18,
			},
		}),
		charts.WithTooltipOpts(opts.Tooltip{
			Trigger: "axis",
			Formatter: opts.FuncOpts(`function (params) {let tooltipText = '';
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
			Show: opts.Bool(true),
			Left: "right",
		}),
		charts.WithXAxisOpts(opts.XAxis{
			Type:      "category",
			Name:      "Date",
			NameGap:   30,
			AxisLabel: &opts.AxisLabel{Show: opts.Bool(true)},
			SplitLine: &opts.SplitLine{Show: opts.Bool(false)},
		}),
		charts.WithYAxisOpts(opts.YAxis{
			Type:      "value",
			Name:      "Response Time (ms)",
			NameGap:   30,
			Min:       0,
			SplitLine: &opts.SplitLine{Show: opts.Bool(true)},
			AxisLabel: &opts.AxisLabel{
				Formatter: "{value} ms",
			},
		}),
		charts.WithGridOpts(opts.Grid{
			Left:         "5%",
			Right:        "5%",
			Bottom:       "15%",
			Top:          "15%",
			ContainLabel: opts.Bool(true),
		}),
	)

	graphData, keys := formatTimeData(responseInfo)

	lineChart.SetXAxis(keys).AddSeries("https://github.com", graphData).
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
