package repository

import (
	"database/sql"
	"fmt"
	"nitinjuyal1610/uptimeMonitor/internal/models"
	"strings"
	"time"
)

type StatRepository interface {
	Create(urlMonitor *models.UrlStats) (int, error)
	BulkCreate(urlStats []*models.UrlStats) ([]int, error)
}

type StatRepositoryPg struct {
	db *sql.DB
}

func NewStatRepository(db *sql.DB) StatRepository {
	return &StatRepositoryPg{db}
}

func (sr *StatRepositoryPg) Create(urlStats *models.UrlStats) (int, error) {
	var entityId int
	createQuery := `
	INSERT INTO url_monitors (
		monitor_id, 
		status_code, 
		response_time, 
		is_up 
	) 
	VALUES ($1, $2, $3, $4) 
	RETURNING id
`

	err := sr.db.QueryRow(
		createQuery,
		urlStats.MonitorId,
		urlStats.StatusCode,
		urlStats.ResponseTime,
		urlStats.IsUp,
	).Scan(&entityId)

	if err != nil {
		return 0, err
	}

	return entityId, nil
}

func (sr *StatRepositoryPg) BulkCreate(urlStats []*models.UrlStats) ([]int, error) {

	if len(urlStats) == 0 {
		return nil, nil
	}

	valueStrings := make([]string, len(urlStats))
	valueArgs := make([]interface{}, 0, len(urlStats)*4)

	for i, stat := range urlStats {
		timeSeconds := float64(stat.ResponseTime) / float64(time.Second)
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
