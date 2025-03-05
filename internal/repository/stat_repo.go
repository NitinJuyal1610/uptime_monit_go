package repository

import (
	"database/sql"
	"fmt"
	"nitinjuyal1610/uptimeMonitor/internal/models"
	"strings"
)

type StatRepository interface {
	GetStatsByMonitorId(monitorId int) (*models.MonitorStats, error)
	GetAvgResponseData(monitorId int, startDate string, endDate string) ([]*models.ResponseTimeStat, error)
	BulkCreate(monitorChecks []*models.MonitorCheck) ([]int, error)
	GetUptimeData(monitorId int, startDate string, endDate string) ([]*models.UptimeStat, error)
}

type StatRepositoryPg struct {
	db *sql.DB
}

func NewStatRepository(db *sql.DB) StatRepository {
	return &StatRepositoryPg{db}
}

func (sr *StatRepositoryPg) GetAvgResponseData(monitorId int, startDate string, endDate string) ([]*models.ResponseTimeStat, error) {
	rsDateQuery := `
		SELECT
			DATE(timestamp) as date,
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

	rows, err := sr.db.Query(rsDateQuery, monitorId, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*models.ResponseTimeStat

	for rows.Next() {
		var rts models.ResponseTimeStat

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

func (sr *StatRepositoryPg) GetUptimeData(monitorId int, startDate string, endDate string) ([]*models.UptimeStat, error) {

	uptimeQuery := `
		SELECT
			COUNT(*) FILTER (WHERE is_up = true) AS successful_checks,
			COUNT(*) FILTER (WHERE is_up = false) AS failed_checks,
			COALESCE (ROUND(
				(COUNT(CASE WHEN is_up THEN 1 END) * 100.0 / NULLIF(COUNT(*), 0)), 2
			),0) AS uptime_percentage,
			DATE(mc.timestamp) as date,
			um.url
		FROM monitor_checks mc
		INNER JOIN url_monitors um on um.id = mc.monitor_id
		WHERE mc.monitor_id = $1 
			AND DATE(mc.timestamp) BETWEEN $2 AND $3
		GROUP BY DATE(mc.timestamp),um.url
	`

	rows, err := sr.db.Query(uptimeQuery, monitorId, startDate, endDate)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*models.UptimeStat

	for rows.Next() {
		var uts models.UptimeStat

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

func (sr *StatRepositoryPg) GetStatsByMonitorId(monitorId int) (*models.MonitorStats, error) {
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

	row := sr.db.QueryRow(statsQuery, monitorId)

	var monitorStats models.MonitorStats

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

func (sr *StatRepositoryPg) BulkCreate(monitorChecks []*models.MonitorCheck) ([]int, error) {

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

	rows, err := sr.db.Query(query, valueArgs...)
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
