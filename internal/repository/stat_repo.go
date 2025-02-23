package repository

import (
	"database/sql"
	"fmt"
	"nitinjuyal1610/uptimeMonitor/internal/models"
	"strings"
)

type StatRepository interface {
	GetStatsByMonitorId(monitorId int) ([]*models.UrlStats, error)
	BulkCreate(urlStats []*models.UrlStats) ([]int, error)
}

type StatRepositoryPg struct {
	db *sql.DB
}

func NewStatRepository(db *sql.DB) StatRepository {
	return &StatRepositoryPg{db}
}

func (sr *StatRepositoryPg) GetStatsByMonitorId(monitorId int) ([]*models.UrlStats, error) {

	statsQuery := `
		SELECT 
			id,
			status_code,
			response_time,
			is_up,
			monitor_id,
			timestamp,
			created_at
		FROM url_stats WHERE monitor_id = $1
	`
	rows, err := sr.db.Query(statsQuery, monitorId)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var urlStats []*models.UrlStats

	for rows.Next() {
		var statItem models.UrlStats
		if err := rows.Scan(&statItem.ID, &statItem.StatusCode, &statItem.ResponseTime, &statItem.IsUp, &statItem.MonitorId, &statItem.Timestamp, &statItem.CreatedAt); err != nil {
			return nil, err
		}
		urlStats = append(urlStats, &statItem)
	}
	return urlStats, nil
}

func (sr *StatRepositoryPg) BulkCreate(urlStats []*models.UrlStats) ([]int, error) {

	if len(urlStats) == 0 {
		return nil, nil
	}

	valueStrings := make([]string, len(urlStats))
	valueArgs := make([]interface{}, 0, len(urlStats)*4)

	for i, stat := range urlStats {
		timeSeconds := stat.ResponseTime
		valueStrings[i] = fmt.Sprintf("($%d, $%d, $%d, $%d)", i*4+1, i*4+2, i*4+3, i*4+4)
		valueArgs = append(valueArgs, stat.MonitorId, stat.StatusCode, timeSeconds, stat.IsUp)
	}

	query := fmt.Sprintf(`
		INSERT INTO url_stats (monitor_id, status_code, response_time, is_up)
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
