package repository

import (
	"database/sql"
	"fmt"
	"nitinjuyal1610/uptimeMonitor/internal/models"
	"strings"
)

type StatRepository interface {
	GetStatsByMonitorId(monitorId int) (*models.MonitorStats, error)
	BulkCreate(monitorChecks []*models.MonitorCheck) ([]int, error)
}

type StatRepositoryPg struct {
	db *sql.DB
}

func NewStatRepository(db *sql.DB) StatRepository {
	return &StatRepositoryPg{db}
}

func (sr *StatRepositoryPg) GetStatsByMonitorId(monitorId int) (*models.MonitorStats, error) {
	statsQuery := `
		WITH hourlyStats AS (
			SELECT  
				ROUND(CAST(AVG(CASE WHEN is_up = true THEN response_time END) AS numeric), 2) AS daily_avg_response_time,
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
			COALESCE(ROUND(CAST(AVG(CASE WHEN is_up = true THEN response_time END) AS numeric), 2), 0.0) AS avg_response_time,
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
	valueArgs := make([]interface{}, 0, len(monitorChecks)*4)

	for i, stat := range monitorChecks {
		timeSeconds := stat.ResponseTime
		valueStrings[i] = fmt.Sprintf("($%d, $%d, $%d, $%d)", i*4+1, i*4+2, i*4+3, i*4+4)
		valueArgs = append(valueArgs, stat.MonitorId, stat.StatusCode, timeSeconds, stat.IsUp)
	}

	query := fmt.Sprintf(`
		INSERT INTO monitor_checks (monitor_id, status_code, response_time, is_up)
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
