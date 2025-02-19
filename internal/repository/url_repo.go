package repository

import (
	"database/sql"
	"fmt"
	"nitinjuyal1610/uptimeMonitor/internal/models"
	"strings"
)

type UrlRepository interface {
	Create(urlMonitor *models.UrlMonitors) (int, error)
	GetAll(status string) ([]*models.UrlMonitors, error)
	GetById(id int) (*models.UrlMonitors, error)
	GetDueMonitorURLs() ([]*models.UrlMonitors, error)
	Update(id int, urlMonitor *models.UrlMonitors) error
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
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("not found")
		}
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

	query := `
	 SELECT
	 	id, 
	    url, 
		status, 
		frequency_minutes, 
		timeout_seconds, 
		last_checked, 
		expected_status_code,
		created_at
	 FROM url_monitors
	 WHERE status = 'ACTIVE' 
	 AND 
	    (
        last_checked IS NULL 
        OR (
            NOW() >= (last_checked + (frequency_minutes * INTERVAL '1 minute'))
        )
    );
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
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		dueMonitorUrls = append(dueMonitorUrls, &dueMonitorUrl)
	}
	return dueMonitorUrls, nil
}

func (u *UrlRepositoryPg) Update(id int, urlMonitor *models.UrlMonitors) error {

	existing, err := u.GetById(id)
	if err != nil {
		return fmt.Errorf("error checking existing record: %w", err)
	}
	if existing == nil {
		return fmt.Errorf("record with ID %d not found", id)
	}

	fields := []string{}
	args := []interface{}{}
	argIndex := 1

	if urlMonitor.Status != "" {
		fields = append(fields, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, urlMonitor.Status)
		argIndex++
	}
	if urlMonitor.FrequencyMinutes > 0 {
		fields = append(fields, fmt.Sprintf("frequency_minutes = $%d", argIndex))
		args = append(args, urlMonitor.FrequencyMinutes)
		argIndex++
	}
	if urlMonitor.TimeoutSeconds > 0 {
		fields = append(fields, fmt.Sprintf("timeout_seconds = $%d", argIndex))
		args = append(args, urlMonitor.TimeoutSeconds)
		argIndex++
	}
	if urlMonitor.ExpectedStatusCode > 0 {
		fields = append(fields, fmt.Sprintf("expected_status_code = $%d", argIndex))
		args = append(args, urlMonitor.ExpectedStatusCode)
		argIndex++
	}

	if !urlMonitor.LastChecked.IsZero() {
		fields = append(fields, fmt.Sprintf("last_checked = $%d", argIndex))
		args = append(args, urlMonitor.LastChecked)
		argIndex++
	}

	if len(fields) == 0 {
		return fmt.Errorf("no fields provided for update")
	}

	query := fmt.Sprintf("UPDATE url_monitors SET %s WHERE id = $%d", strings.Join(fields, ", "), argIndex)
	args = append(args, id)

	result, err := u.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("failed to update url monitor: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("no records updated, possibly due to no changes")
	}

	return nil
}
