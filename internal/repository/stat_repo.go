package repository

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"nitinjuyal1610/uptimeMonitor/internal/models"
	"nitinjuyal1610/uptimeMonitor/pkg/types"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

type StatRepository interface {
	GetStatsByMonitorId(ctx context.Context, monitorId int) (*types.MonitorStats, error)
	GetAvgResponseData(ctx context.Context, monitorId int, startDate string, endDate string) ([]*types.ResponseTimeStat, error)
	BulkCreate(ctx context.Context, monitorChecks []*models.MonitorCheck) ([]int, error)
	GetUptimeData(ctx context.Context, monitorId int, startDate string, endDate string) ([]*types.UptimeStat, error)
	GetDetailedTimeData(ctx context.Context, monitorId int, startDate string, endDate string) ([]*types.DetailedTimeStat, error)
}

type StatRepositoryPg struct {
	db *pgxpool.Pool
}

func NewStatRepository(db *pgxpool.Pool) StatRepository {
	return &StatRepositoryPg{db}
}

func (sr *StatRepositoryPg) GetAvgResponseData(ctx context.Context, monitorId int, startDate string, endDate string) ([]*types.ResponseTimeStat, error) {
	query := `
		SELECT
			TO_CHAR(DATE(timestamp),'YYYY-MM-DD') as date,
			mc.monitor_id ,
			um.url,
			COALESCE(ROUND(CAST(AVG(CASE WHEN mc.is_up = true THEN mc.response_time END) AS numeric), 3), 0.0) AS avg_response_time
		FROM
			monitor_checks mc
		INNER JOIN url_monitors um ON
			um.id = mc.monitor_id
		WHERE mc.monitor_id = $1 AND DATE(timestamp) BETWEEN $2 AND $3
		GROUP BY DATE(timestamp), mc.monitor_id , um.url 
	`

	rows, err := sr.db.Query(ctx, query, monitorId, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*types.ResponseTimeStat

	for rows.Next() {
		var rts types.ResponseTimeStat

		err := rows.Scan(&rts.Date, &rts.MonitorID, &rts.Url, &rts.AvgResponseTime)
		if err != nil {
			return nil, err
		}

		results = append(results, &rts)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

func (sr *StatRepositoryPg) GetUptimeData(ctx context.Context, monitorId int, startDate string, endDate string) ([]*types.UptimeStat, error) {

	uptimeQuery := `
		SELECT
			COUNT(*) FILTER (WHERE is_up = true) AS successful_checks,
			COUNT(*) FILTER (WHERE is_up = false) AS failed_checks,
			COALESCE (ROUND(
				(COUNT(CASE WHEN is_up THEN 1 END) * 100.0 / NULLIF(COUNT(*), 0)), 2
			),0) AS uptime_percentage,
			TO_CHAR(DATE(mc.timestamp),'YYYY-MM-DD') as date,
			um.url
		FROM monitor_checks mc
		INNER JOIN url_monitors um on um.id = mc.monitor_id
		WHERE mc.monitor_id = $1 
			AND DATE(mc.timestamp) BETWEEN $2 AND $3
		GROUP BY DATE(mc.timestamp),um.url
	`

	rows, err := sr.db.Query(ctx, uptimeQuery, monitorId, startDate, endDate)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*types.UptimeStat

	for rows.Next() {
		var uts types.UptimeStat

		err := rows.Scan(&uts.SuccessfulChecks, &uts.FailedChecks, &uts.UptimePercentage, &uts.Date, &uts.Url)
		if err != nil {
			return nil, err
		}

		results = append(results, &uts)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

func (sr *StatRepositoryPg) GetStatsByMonitorId(ctx context.Context, monitorId int) (*types.MonitorStats, error) {
	statsQuery := `
		WITH hourlyStats AS (
			SELECT  
				ROUND(CAST(AVG(CASE WHEN is_up = true THEN response_time END) AS numeric), 3) AS daily_avg_response_time,
				(COUNT(CASE WHEN is_up THEN 1 END) * 100.0 / COUNT(*)) AS daily_uptime_percentage,
				monitor_id 
			FROM monitor_checks
			WHERE monitor_id = $1
			GROUP BY monitor_id 
		)
		SELECT 
			um.id AS id,
			um.url AS url,
			um.status AS status,
			um.last_checked AS last_checked,
			COUNT(*) AS total_checks,
			COUNT(*) FILTER (WHERE is_up = true) AS successful_checks,
			COUNT(*) FILTER (WHERE is_up = false) AS failed_checks,
			COALESCE(ROUND(CAST(AVG(CASE WHEN is_up = true THEN response_time END) AS numeric), 3), 0.0) AS avg_response_time,
			ROUND(
				(COUNT(CASE WHEN is_up THEN 1 END) * 100.0 / COUNT(*)), 
				2
			) AS uptime_percentage,
			COALESCE(ROUND(hs.daily_uptime_percentage, 2), 0.0) AS daily_uptime_percentage,
			COALESCE(hs.daily_avg_response_time, 0.0) AS daily_avg_response_time
		FROM monitor_checks us 
		LEFT JOIN hourlyStats hs ON hs.monitor_id = us.monitor_id
		LEFT JOIN url_monitors um ON um.id = us.monitor_id 
		WHERE us.monitor_id = $1
		GROUP BY um.id, um.url, um.status, um.last_checked, hs.daily_uptime_percentage, hs.daily_avg_response_time, hs.monitor_id;
	`

	row := sr.db.QueryRow(ctx, statsQuery, monitorId)

	var monitorStats types.MonitorStats

	err := row.Scan(
		&monitorStats.ID,
		&monitorStats.Url,
		&monitorStats.Status,
		&monitorStats.LastChecked,
		&monitorStats.TotalChecks,
		&monitorStats.SuccessfulChecks,
		&monitorStats.FailedChecks,
		&monitorStats.AvgResponseTime,
		&monitorStats.UptimePercentage,
		&monitorStats.DailyUptimePercentage,
		&monitorStats.DailyAvgResponseTime,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &monitorStats, nil
}

func (sr *StatRepositoryPg) BulkCreate(ctx context.Context, monitorChecks []*models.MonitorCheck) ([]int, error) {

	if len(monitorChecks) == 0 {
		return nil, nil
	}

	valueStrings := make([]string, len(monitorChecks))
	valueArgs := make([]any, 0, len(monitorChecks)*4)

	for i, stat := range monitorChecks {
		timeSeconds := stat.ResponseTime
		valueStrings[i] = fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d)", i*7+1, i*7+2, i*7+3, i*7+4, i*7+5, i*7+6, i*7+7)
		valueArgs = append(valueArgs, stat.MonitorId, stat.StatusCode, timeSeconds, stat.IsUp, stat.Ttfb, stat.ContentSize, stat.RequestType)
	}

	query := fmt.Sprintf(`
		INSERT INTO monitor_checks (monitor_id, status_code, response_time, is_up, ttfb, content_size, request_type)
		VALUES %s RETURNING id`, strings.Join(valueStrings, ", "))

	rows, err := sr.db.Query(ctx, query, valueArgs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entityIds []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		entityIds = append(entityIds, id)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	return entityIds, nil
}

func (sr *StatRepositoryPg) GetDetailedTimeData(ctx context.Context, monitorId int, startDate string, endDate string) ([]*types.DetailedTimeStat, error) {
	rsDateQuery := `
		SELECT 
			mc.timestamp,
			mc.response_time,
			mc.ttfb,
			mc.content_size,
			mc.request_type,
			um.url
		FROM 
			monitor_checks mc
		INNER JOIN 
			url_monitors um 
			ON um.id = mc.monitor_id
		WHERE 
			mc.monitor_id = $1
			AND DATE(mc.timestamp) BETWEEN $2 AND $3
			AND mc.is_up = true
			AND mc.request_type=$4
	`

	rows, err := sr.db.Query(ctx, rsDateQuery, monitorId, startDate, endDate, http.MethodGet)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*types.DetailedTimeStat

	for rows.Next() {
		var rts types.DetailedTimeStat

		err := rows.Scan(&rts.Timestamp, &rts.ResponseTime, &rts.Ttfb, &rts.ContentSize, &rts.RequestType, &rts.Url)
		if err != nil {
			return nil, err
		}

		results = append(results, &rts)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}
