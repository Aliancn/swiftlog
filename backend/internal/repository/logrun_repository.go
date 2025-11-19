package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/aliancn/swiftlog/backend/internal/models"
	"github.com/google/uuid"
)

// LogRunRepository handles database operations for log runs
type LogRunRepository struct {
	db *sql.DB
}

// NewLogRunRepository creates a new log run repository
func NewLogRunRepository(db *sql.DB) *LogRunRepository {
	return &LogRunRepository{db: db}
}

// Create creates a new log run
func (r *LogRunRepository) Create(ctx context.Context, groupID uuid.UUID) (*models.LogRun, error) {
	run := &models.LogRun{}
	query := `
		INSERT INTO log_runs (group_id, start_time, status, ai_status)
		VALUES ($1, $2, $3, $4)
		RETURNING id, group_id, start_time, end_time, status, exit_code, ai_report, ai_status, created_at, updated_at
	`
	err := r.db.QueryRowContext(ctx, query, groupID, time.Now(), models.RunStatusRunning, models.AIStatusPending).Scan(
		&run.ID,
		&run.GroupID,
		&run.StartTime,
		&run.EndTime,
		&run.Status,
		&run.ExitCode,
		&run.AIReport,
		&run.AIStatus,
		&run.CreatedAt,
		&run.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create log run: %w", err)
	}
	return run, nil
}

// GetByID retrieves a log run by ID
func (r *LogRunRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.LogRun, error) {
	run := &models.LogRun{}
	query := `
		SELECT id, group_id, start_time, end_time, status, exit_code, ai_report, ai_status, created_at, updated_at
		FROM log_runs
		WHERE id = $1
	`
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&run.ID,
		&run.GroupID,
		&run.StartTime,
		&run.EndTime,
		&run.Status,
		&run.ExitCode,
		&run.AIReport,
		&run.AIStatus,
		&run.CreatedAt,
		&run.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("log run not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get log run: %w", err)
	}
	return run, nil
}

// UpdateStatus updates the status and exit code of a log run
func (r *LogRunRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status models.RunStatus, exitCode *int32) error {
	query := `
		UPDATE log_runs
		SET status = $1, exit_code = $2, end_time = $3
		WHERE id = $4
	`
	endTime := sql.NullTime{Time: time.Now(), Valid: true}
	var exitCodeVal sql.NullInt32
	if exitCode != nil {
		exitCodeVal = sql.NullInt32{Int32: *exitCode, Valid: true}
	}

	_, err := r.db.ExecContext(ctx, query, status, exitCodeVal, endTime, id)
	if err != nil {
		return fmt.Errorf("failed to update log run status: %w", err)
	}
	return nil
}

// ListByGroupID retrieves all log runs for a specific group
func (r *LogRunRepository) ListByGroupID(ctx context.Context, groupID uuid.UUID, limit, offset int) ([]*models.LogRun, error) {
	query := `
		SELECT id, group_id, start_time, end_time, status, exit_code, ai_report, ai_status, created_at, updated_at
		FROM log_runs
		WHERE group_id = $1
		ORDER BY start_time DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.QueryContext(ctx, query, groupID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list log runs: %w", err)
	}
	defer rows.Close()

	var runs []*models.LogRun
	for rows.Next() {
		run := &models.LogRun{}
		err := rows.Scan(
			&run.ID,
			&run.GroupID,
			&run.StartTime,
			&run.EndTime,
			&run.Status,
			&run.ExitCode,
			&run.AIReport,
			&run.AIStatus,
			&run.CreatedAt,
			&run.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan log run: %w", err)
		}
		runs = append(runs, run)
	}

	return runs, nil
}

// UpdateAIReport updates the AI report and status for a log run
func (r *LogRunRepository) UpdateAIReport(ctx context.Context, id uuid.UUID, report string, status models.AIStatus) error {
	query := `
		UPDATE log_runs
		SET ai_report = $1, ai_status = $2
		WHERE id = $3
	`
	_, err := r.db.ExecContext(ctx, query, report, status, id)
	if err != nil {
		return fmt.Errorf("failed to update AI report: %w", err)
	}
	return nil
}

// UpdateAIStatus updates only the AI status for a log run
func (r *LogRunRepository) UpdateAIStatus(ctx context.Context, id uuid.UUID, status models.AIStatus) error {
	query := `UPDATE log_runs SET ai_status = $1 WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, status, id)
	if err != nil {
		return fmt.Errorf("failed to update AI status: %w", err)
	}
	return nil
}

// ListPendingAIJobs retrieves runs pending AI analysis
func (r *LogRunRepository) ListPendingAIJobs(ctx context.Context, limit int) ([]*models.LogRun, error) {
	query := `
		SELECT id, group_id, start_time, end_time, status, exit_code, ai_report, ai_status, created_at, updated_at
		FROM log_runs
		WHERE ai_status = 'pending'
		  AND status IN ('completed', 'failed', 'aborted')
		ORDER BY end_time DESC
		LIMIT $1
	`
	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query pending AI jobs: %w", err)
	}
	defer rows.Close()

	var runs []*models.LogRun
	for rows.Next() {
		run := &models.LogRun{}
		err := rows.Scan(
			&run.ID,
			&run.GroupID,
			&run.StartTime,
			&run.EndTime,
			&run.Status,
			&run.ExitCode,
			&run.AIReport,
			&run.AIStatus,
			&run.CreatedAt,
			&run.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan log run: %w", err)
		}
		runs = append(runs, run)
	}

	return runs, nil
}

// GetStatusStatistics retrieves overall statistics for log runs and AI analysis
func (r *LogRunRepository) GetStatusStatistics(ctx context.Context) (*models.StatusStatistics, error) {
	stats := &models.StatusStatistics{}

	// Get run status counts
	query := `
		SELECT
			COUNT(*) FILTER (WHERE status = 'running') as running,
			COUNT(*) FILTER (WHERE status = 'completed') as completed,
			COUNT(*) FILTER (WHERE status = 'failed') as failed,
			COUNT(*) FILTER (WHERE status = 'aborted') as aborted
		FROM log_runs
	`
	err := r.db.QueryRowContext(ctx, query).Scan(
		&stats.RunningCount,
		&stats.CompletedCount,
		&stats.FailedCount,
		&stats.AbortedCount,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get run statistics: %w", err)
	}

	// Get AI status counts
	aiQuery := `
		SELECT
			COUNT(*) FILTER (WHERE ai_status = 'pending') as pending,
			COUNT(*) FILTER (WHERE ai_status = 'processing') as processing,
			COUNT(*) FILTER (WHERE ai_status = 'completed') as completed,
			COUNT(*) FILTER (WHERE ai_status = 'failed') as failed
		FROM log_runs
	`
	err = r.db.QueryRowContext(ctx, aiQuery).Scan(
		&stats.AIPendingCount,
		&stats.AIProcessingCount,
		&stats.AICompletedCount,
		&stats.AIFailedCount,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get AI statistics: %w", err)
	}

	return stats, nil
}

// ListRecentRuns retrieves the most recent log runs across all groups
func (r *LogRunRepository) ListRecentRuns(ctx context.Context, limit int) ([]*models.LogRun, error) {
	query := `
		SELECT id, group_id, start_time, end_time, status, exit_code, ai_report, ai_status, created_at, updated_at
		FROM log_runs
		ORDER BY start_time DESC
		LIMIT $1
	`
	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list recent runs: %w", err)
	}
	defer rows.Close()

	var runs []*models.LogRun
	for rows.Next() {
		run := &models.LogRun{}
		err := rows.Scan(
			&run.ID,
			&run.GroupID,
			&run.StartTime,
			&run.EndTime,
			&run.Status,
			&run.ExitCode,
			&run.AIReport,
			&run.AIStatus,
			&run.CreatedAt,
			&run.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan log run: %w", err)
		}
		runs = append(runs, run)
	}

	return runs, nil
}
