package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/aliancn/swiftlog/backend/internal/models"
	"github.com/google/uuid"
)

// LogGroupRepository handles database operations for log groups
type LogGroupRepository struct {
	db *sql.DB
}

// NewLogGroupRepository creates a new log group repository
func NewLogGroupRepository(db *sql.DB) *LogGroupRepository {
	return &LogGroupRepository{db: db}
}

// Create creates a new log group
func (r *LogGroupRepository) Create(ctx context.Context, projectID uuid.UUID, name string) (*models.LogGroup, error) {
	group := &models.LogGroup{}
	query := `
		INSERT INTO log_groups (project_id, name)
		VALUES ($1, $2)
		RETURNING id, project_id, name, created_at
	`
	err := r.db.QueryRowContext(ctx, query, projectID, name).Scan(
		&group.ID,
		&group.ProjectID,
		&group.Name,
		&group.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create log group: %w", err)
	}
	return group, nil
}

// GetOrCreate gets an existing log group or creates it if it doesn't exist
func (r *LogGroupRepository) GetOrCreate(ctx context.Context, projectID uuid.UUID, name string) (*models.LogGroup, error) {
	// Try to get existing group
	group, err := r.GetByProjectAndName(ctx, projectID, name)
	if err == nil {
		return group, nil
	}

	// Create new group if not found
	return r.Create(ctx, projectID, name)
}

// GetByID retrieves a log group by ID
func (r *LogGroupRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.LogGroup, error) {
	group := &models.LogGroup{}
	query := `SELECT id, project_id, name, created_at FROM log_groups WHERE id = $1`
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&group.ID,
		&group.ProjectID,
		&group.Name,
		&group.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("log group not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get log group: %w", err)
	}
	return group, nil
}

// GetByProjectAndName retrieves a log group by project ID and name
func (r *LogGroupRepository) GetByProjectAndName(ctx context.Context, projectID uuid.UUID, name string) (*models.LogGroup, error) {
	group := &models.LogGroup{}
	query := `SELECT id, project_id, name, created_at FROM log_groups WHERE project_id = $1 AND name = $2`
	err := r.db.QueryRowContext(ctx, query, projectID, name).Scan(
		&group.ID,
		&group.ProjectID,
		&group.Name,
		&group.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("log group not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get log group: %w", err)
	}
	return group, nil
}

// ListByProjectID retrieves all log groups for a project
func (r *LogGroupRepository) ListByProjectID(ctx context.Context, projectID uuid.UUID) ([]*models.LogGroup, error) {
	query := `
		SELECT id, project_id, name, created_at
		FROM log_groups
		WHERE project_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list log groups: %w", err)
	}
	defer rows.Close()

	var groups []*models.LogGroup
	for rows.Next() {
		group := &models.LogGroup{}
		err := rows.Scan(
			&group.ID,
			&group.ProjectID,
			&group.Name,
			&group.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan log group: %w", err)
		}
		groups = append(groups, group)
	}

	return groups, nil
}
