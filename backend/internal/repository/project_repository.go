package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/aliancn/swiftlog/backend/internal/models"
	"github.com/google/uuid"
)

// ProjectRepository handles database operations for projects
type ProjectRepository struct {
	db *sql.DB
}

// NewProjectRepository creates a new project repository
func NewProjectRepository(db *sql.DB) *ProjectRepository {
	return &ProjectRepository{db: db}
}

// Create creates a new project
func (r *ProjectRepository) Create(ctx context.Context, userID uuid.UUID, name string) (*models.Project, error) {
	project := &models.Project{}
	query := `
		INSERT INTO projects (user_id, name)
		VALUES ($1, $2)
		RETURNING id, user_id, name, created_at
	`
	err := r.db.QueryRowContext(ctx, query, userID, name).Scan(
		&project.ID,
		&project.UserID,
		&project.Name,
		&project.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create project: %w", err)
	}
	return project, nil
}

// GetOrCreate gets an existing project or creates it if it doesn't exist
func (r *ProjectRepository) GetOrCreate(ctx context.Context, userID uuid.UUID, name string) (*models.Project, error) {
	// Try to get existing project
	project, err := r.GetByUserAndName(ctx, userID, name)
	if err == nil {
		return project, nil
	}

	// Create new project if not found
	return r.Create(ctx, userID, name)
}

// GetByID retrieves a project by ID
func (r *ProjectRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Project, error) {
	project := &models.Project{}
	query := `SELECT id, user_id, name, created_at FROM projects WHERE id = $1`
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&project.ID,
		&project.UserID,
		&project.Name,
		&project.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("project not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}
	return project, nil
}

// GetByUserAndName retrieves a project by user ID and name
func (r *ProjectRepository) GetByUserAndName(ctx context.Context, userID uuid.UUID, name string) (*models.Project, error) {
	project := &models.Project{}
	query := `SELECT id, user_id, name, created_at FROM projects WHERE user_id = $1 AND name = $2`
	err := r.db.QueryRowContext(ctx, query, userID, name).Scan(
		&project.ID,
		&project.UserID,
		&project.Name,
		&project.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("project not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}
	return project, nil
}

// ListByUserID retrieves all projects for a user
func (r *ProjectRepository) ListByUserID(ctx context.Context, userID uuid.UUID) ([]*models.Project, error) {
	query := `
		SELECT id, user_id, name, created_at
		FROM projects
		WHERE user_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}
	defer rows.Close()

	var projects []*models.Project
	for rows.Next() {
		project := &models.Project{}
		err := rows.Scan(
			&project.ID,
			&project.UserID,
			&project.Name,
			&project.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan project: %w", err)
		}
		projects = append(projects, project)
	}

	return projects, nil
}

// ListAll retrieves all projects (for development/admin use)
func (r *ProjectRepository) ListAll(ctx context.Context) ([]*models.Project, error) {
	query := `
		SELECT id, user_id, name, created_at
		FROM projects
		ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}
	defer rows.Close()

	var projects []*models.Project
	for rows.Next() {
		project := &models.Project{}
		err := rows.Scan(
			&project.ID,
			&project.UserID,
			&project.Name,
			&project.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan project: %w", err)
		}
		projects = append(projects, project)
	}

	return projects, nil
}
