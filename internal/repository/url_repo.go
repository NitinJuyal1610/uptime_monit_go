package repository

import (
	"database/sql"
	"fmt"
	"nitinjuyal1610/uptimeMonitor/internal/models"
)

type UrlRepository interface {
	Create(urlMonitor *models.UrlMonitors) (int, error)
	GetAll(status string) ([]*models.UrlMonitors, error)
	GetById(id int) (*models.UrlMonitors, error)
	GetDueMonitorURLs() ([]*models.UrlMonitors, error)
}

type UrlRepositoryPg struct {
	db *sql.DB
}

func NewUrlRepository(db *sql.DB) UrlRepository {
	return &UrlRepositoryPg{db}
}

func (u *UrlRepositoryPg) Create(urlMonitor *models.UrlMonitors) (int, error) {
	var entityId int
	createQuery := `
	INSERT INTO url_monitors (
		url, 
		frequency_minutes, 
		status, 
		timeout_seconds, 
		last_checked, 
		expected_status_code
	) 
	VALUES ($1, $2, $3, $4, $5, $6) 
	RETURNING id
`

	err := u.db.QueryRow(
		createQuery,
		urlMonitor.Url,
		urlMonitor.FrequencyMinutes,
		urlMonitor.Status,
		urlMonitor.TimeoutSeconds,
		urlMonitor.LastChecked,
		urlMonitor.ExpectedStatusCode,
	).Scan(&entityId)

	if err != nil {
		return 0, err
	}

	return entityId, nil
}

func (u *UrlRepositoryPg) GetAll(status string) ([]*models.UrlMonitors, error) {
	query := `
		SELECT 
			id, url, status, frequency_minutes, timeout_seconds, 
			last_checked, expected_status_code, created_at, updated_at 
		FROM url_monitors`

	var args []any
	if status != "" {
		query += ` WHERE status=$1;`

		args = append(args, status)
	}

	rows, err := u.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	var monitors []*models.UrlMonitors

	for rows.Next() {
		var monitor models.UrlMonitors

		err := rows.Scan(
			&monitor.ID,
			&monitor.Url,
			&monitor.Status,
			&monitor.FrequencyMinutes,
			&monitor.TimeoutSeconds,
			&monitor.LastChecked,
			&monitor.ExpectedStatusCode,
			&monitor.CreatedAt,
			&monitor.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		monitors = append(monitors, &monitor)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over rows: %w", err)
	}

	return monitors, nil
}

func (u *UrlRepositoryPg) GetById(id int) (*models.UrlMonitors, error) {
	query := `
		SELECT 
			id, url, status, frequency_minutes, timeout_seconds, 
			last_checked, expected_status_code, created_at, updated_at 
		FROM url_monitors WHERE id = $1`

	var monitor models.UrlMonitors
	err := u.db.QueryRow(query, id).Scan(
		&monitor.ID,
		&monitor.Url,
		&monitor.Status,
		&monitor.FrequencyMinutes,
		&monitor.TimeoutSeconds,
		&monitor.LastChecked,
		&monitor.ExpectedStatusCode,
		&monitor.CreatedAt,
		&monitor.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to scan row: %w", err)
	}

	return &monitor, nil
}

func (u *UrlRepositoryPg) GetByUrl(id int) (*models.UrlMonitors, error) {
	query := `
		SELECT 
			id, url, status, frequency_minutes, timeout_seconds, 
			last_checked, expected_status_code, created_at, updated_at 
		FROM url_monitors WHERE id = $1`

	var monitor models.UrlMonitors
	err := u.db.QueryRow(query, id).Scan(
		&monitor.ID,
		&monitor.Url,
		&monitor.Status,
		&monitor.FrequencyMinutes,
		&monitor.TimeoutSeconds,
		&monitor.LastChecked,
		&monitor.ExpectedStatusCode,
		&monitor.CreatedAt,
		&monitor.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to scan row: %w", err)
	}

	return &monitor, nil
}

func (u *UrlRepositoryPg) GetDueMonitorURLs() ([]*models.UrlMonitors, error) {
	//
	query := `
	 SELECT 
	    url, 
		frequency_minutes, 
		status, 
		timeout_seconds, 
		last_checked, 
		expected_status_code
	 WHERE status = 'ACTIVE' AND NOW() - last_checked >= (frequency_in_minutes * INTERVAL '1 minute');
	`

	rows, err := u.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	var dueMonitorUrls []*models.UrlMonitors
	for rows.Next() {
		var dueMonitorUrl models.UrlMonitors

		err := rows.Scan(
			&dueMonitorUrl.ID,
			&dueMonitorUrl.Url,
			&dueMonitorUrl.Status,
			&dueMonitorUrl.FrequencyMinutes,
			&dueMonitorUrl.TimeoutSeconds,
			&dueMonitorUrl.LastChecked,
			&dueMonitorUrl.ExpectedStatusCode,
			&dueMonitorUrl.CreatedAt,
			&dueMonitorUrl.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		dueMonitorUrls = append(dueMonitorUrls, &dueMonitorUrl)
	}
	return dueMonitorUrls, nil
}
