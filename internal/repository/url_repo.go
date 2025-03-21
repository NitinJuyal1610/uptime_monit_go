package repository

import (
	"context"
	"database/sql"
	"fmt"
	"nitinjuyal1610/uptimeMonitor/internal/models"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

type UrlRepository interface {
	Create(ctx context.Context, urlMonitor *models.UrlMonitors) (int, error)
	GetAll(ctx context.Context, status string, keyword string, userId int) ([]*models.UrlMonitors, error)
	GetById(ctx context.Context, id int) (*models.UrlMonitors, error)
	GetDueMonitors(ctx context.Context) ([]*models.UrlMonitors, error)
	Update(ctx context.Context, id int, urlMonitor *models.UrlMonitors) error
	BulkUpdate(ctx context.Context, updates map[int]*models.UrlMonitors) error
}

type UrlRepositoryPg struct {
	db *pgxpool.Pool
}

func NewUrlRepository(db *pgxpool.Pool) UrlRepository {
	return &UrlRepositoryPg{db}
}

func (u *UrlRepositoryPg) Create(ctx context.Context, urlMonitor *models.UrlMonitors) (int, error) {
	var entityId int

	createQuery := `
	INSERT INTO url_monitors (
		url, 
		frequency_minutes, 
		status, 
		timeout_seconds, 
		last_checked, 
		expected_status_code,
		collect_detailed_data,
		max_fail_threshold,
		alert_email,
		user_id
	) 
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9,$10) 
	RETURNING id
`

	err := u.db.QueryRow(
		ctx,
		createQuery,
		urlMonitor.Url,
		urlMonitor.FrequencyMinutes,
		urlMonitor.Status,
		urlMonitor.TimeoutSeconds,
		urlMonitor.LastChecked,
		urlMonitor.ExpectedStatusCode,
		urlMonitor.CollectDetailedData,
		urlMonitor.MaxFailThreshold,
		urlMonitor.AlertEmail,
		urlMonitor.UserId,
	).Scan(&entityId)

	if err != nil {
		return 0, err
	}

	return entityId, nil
}

func (u *UrlRepositoryPg) GetAll(ctx context.Context, status string, keyword string, userId int) ([]*models.UrlMonitors, error) {
	query := `
		SELECT 
			id, REGEXP_REPLACE(url, '^https?://', '', 'i') AS trimmed_url, status, frequency_minutes, timeout_seconds, 
			last_checked, expected_status_code, collect_detailed_data, created_at, updated_at 
		FROM url_monitors WHERE user_id = $1`

	var args []any
	args = append(args, userId)

	if status != "" {
		query += ` AND status = $` + fmt.Sprint(len(args)+1)
		args = append(args, status)
	}

	if keyword != "" {
		query += ` AND url LIKE $` + fmt.Sprint(len(args)+1)
		args = append(args, "%"+keyword+"%")
	}

	rows, err := u.db.Query(ctx, query, args...)
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
			&monitor.CollectDetailedData,
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

func (u *UrlRepositoryPg) GetById(ctx context.Context, id int) (*models.UrlMonitors, error) {
	query := `
		SELECT 
			id, url, status, frequency_minutes, timeout_seconds, 
			last_checked, expected_status_code,collect_detailed_data, created_at, updated_at 
		FROM url_monitors WHERE id = $1`

	var monitor models.UrlMonitors
	err := u.db.QueryRow(ctx, query, id).Scan(
		&monitor.ID,
		&monitor.Url,
		&monitor.Status,
		&monitor.FrequencyMinutes,
		&monitor.TimeoutSeconds,
		&monitor.LastChecked,
		&monitor.ExpectedStatusCode,
		&monitor.CollectDetailedData,
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

func (u *UrlRepositoryPg) GetByUrl(ctx context.Context, id int) (*models.UrlMonitors, error) {
	query := `
		SELECT 
			id, url, status, frequency_minutes, timeout_seconds, 
			last_checked, expected_status_code, collect_detailed_data, created_at, updated_at 
		FROM url_monitors WHERE id = $1`

	var monitor models.UrlMonitors
	err := u.db.QueryRow(ctx, query, id).Scan(
		&monitor.ID,
		&monitor.Url,
		&monitor.Status,
		&monitor.FrequencyMinutes,
		&monitor.TimeoutSeconds,
		&monitor.LastChecked,
		&monitor.ExpectedStatusCode,
		&monitor.CollectDetailedData,
		&monitor.CreatedAt,
		&monitor.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to scan row: %w", err)
	}

	return &monitor, nil
}

func (u *UrlRepositoryPg) GetDueMonitors(ctx context.Context) ([]*models.UrlMonitors, error) {

	query := `
	 SELECT
	 	id, 
	    url, 
		status, 
		frequency_minutes, 
		timeout_seconds, 
		last_checked, 
		expected_status_code,
		collect_detailed_data,
		consecutive_fails,
		max_fail_threshold,
		alert_email,
		created_at
	 FROM url_monitors
	 WHERE status NOT IN ('PAUSED','DELETED','PENDING')
	 AND 
	    (
        last_checked IS NULL 
        OR (
            NOW() >= (last_checked + (frequency_minutes * INTERVAL '1 minute'))
        )
    );
	`

	rows, err := u.db.Query(ctx, query)
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
			&dueMonitorUrl.CollectDetailedData,
			&dueMonitorUrl.ConsecutiveFails,
			&dueMonitorUrl.MaxFailThreshold,
			&dueMonitorUrl.AlertEmail,
			&dueMonitorUrl.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		dueMonitorUrls = append(dueMonitorUrls, &dueMonitorUrl)
	}
	return dueMonitorUrls, nil
}

func (u *UrlRepositoryPg) Update(ctx context.Context, id int, urlMonitor *models.UrlMonitors) error {

	existing, err := u.GetById(ctx, id)
	if err != nil {
		return fmt.Errorf("error checking existing record: %w", err)
	}
	if existing == nil {
		return fmt.Errorf("record with ID %d not found", id)
	}

	fields := []string{}
	args := []any{}
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

	result, err := u.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update url monitor: %w", err)
	}

	rowsAffected := result.RowsAffected()

	if rowsAffected == 0 {
		return fmt.Errorf("no records updated, possibly due to no changes")
	}

	return nil
}

func (u *UrlRepositoryPg) BulkUpdate(ctx context.Context, updates map[int]*models.UrlMonitors) error {
	if len(updates) == 0 {
		return nil
	}

	var placeholderValues []any
	var queryParts []string
	var monitorIDs []int

	paramIndex := 2
	for monitorID, update := range updates {
		queryParts = append(queryParts, fmt.Sprintf("WHEN $%d THEN $%d", paramIndex, paramIndex+1))
		placeholderValues = append(placeholderValues, monitorID, update.Status)
		monitorIDs = append(monitorIDs, monitorID)
		paramIndex += 2

	}

	if len(monitorIDs) == 0 {
		return nil
	}

	wherePlaceholders := make([]string, len(monitorIDs))
	for i := range monitorIDs {
		wherePlaceholders[i] = fmt.Sprintf("$%d", paramIndex+i)
	}

	query := fmt.Sprintf(`
        UPDATE url_monitors
        SET 
            last_checked = $1,
            status = CASE id %s ELSE status END
        WHERE id IN (%s)
    `,
		strings.Join(queryParts, " "),
		strings.Join(wherePlaceholders, ", "),
	)

	firstMonitor := updates[monitorIDs[0]]
	allParams := []any{firstMonitor.LastChecked}

	allParams = append(allParams, placeholderValues...)
	for _, id := range monitorIDs {
		allParams = append(allParams, id)
	}

	result, err := u.db.Exec(ctx, query, allParams...)
	if err != nil {
		return fmt.Errorf("failed to perform bulk update: %w", err)
	}

	rowsAffected := result.RowsAffected()

	if rowsAffected == 0 {
		return fmt.Errorf("no rows were updated")
	}

	return nil
}
