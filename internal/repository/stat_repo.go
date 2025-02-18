package repository

import (
	"database/sql"
	"nitinjuyal1610/uptimeMonitor/internal/models"
)

type StatRepository interface {
	Create(urlMonitor *models.UrlStats) (int, error)
}

type StatRepositoryPg struct {
	db *sql.DB
}

func NewStatRepository(db *sql.DB) StatRepository {
	return &StatRepositoryPg{db}
}

func (u *StatRepositoryPg) Create(urlStats *models.UrlStats) (int, error) {
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

	err := u.db.QueryRow(
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
