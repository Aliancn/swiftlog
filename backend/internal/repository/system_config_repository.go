package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"

	"github.com/aliancn/swiftlog/backend/internal/models"
	"github.com/google/uuid"
)

// SystemConfigRepository handles database operations for system configuration
type SystemConfigRepository struct {
	db *sql.DB
}

// NewSystemConfigRepository creates a new system config repository
func NewSystemConfigRepository(db *sql.DB) *SystemConfigRepository {
	return &SystemConfigRepository{db: db}
}

// Get retrieves a configuration value by key
func (r *SystemConfigRepository) Get(ctx context.Context, key string) (*models.SystemConfig, error) {
	config := &models.SystemConfig{}
	query := `
		SELECT id, key, value, description, updated_at, updated_by
		FROM system_config
		WHERE key = $1
	`
	err := r.db.QueryRowContext(ctx, query, key).Scan(
		&config.ID,
		&config.Key,
		&config.Value,
		&config.Description,
		&config.UpdatedAt,
		&config.UpdatedBy,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("config not found: %s", key)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get config: %w", err)
	}
	return config, nil
}

// GetValue retrieves just the value for a configuration key
func (r *SystemConfigRepository) GetValue(ctx context.Context, key string) (string, error) {
	config, err := r.Get(ctx, key)
	if err != nil {
		return "", err
	}
	return config.Value, nil
}

// GetBool retrieves a boolean configuration value
func (r *SystemConfigRepository) GetBool(ctx context.Context, key string) (bool, error) {
	value, err := r.GetValue(ctx, key)
	if err != nil {
		return false, err
	}
	return value == "true" || value == "1", nil
}

// GetInt retrieves an integer configuration value
func (r *SystemConfigRepository) GetInt(ctx context.Context, key string) (int, error) {
	value, err := r.GetValue(ctx, key)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(value)
}

// Set updates or creates a configuration value
func (r *SystemConfigRepository) Set(ctx context.Context, key, value string, updatedBy uuid.UUID) (*models.SystemConfig, error) {
	config := &models.SystemConfig{}
	query := `
		INSERT INTO system_config (key, value, updated_by)
		VALUES ($1, $2, $3)
		ON CONFLICT (key) DO UPDATE
		SET value = EXCLUDED.value,
		    updated_by = EXCLUDED.updated_by,
		    updated_at = NOW()
		RETURNING id, key, value, description, updated_at, updated_by
	`
	err := r.db.QueryRowContext(ctx, query, key, value, uuid.NullUUID{UUID: updatedBy, Valid: true}).Scan(
		&config.ID,
		&config.Key,
		&config.Value,
		&config.Description,
		&config.UpdatedAt,
		&config.UpdatedBy,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to set config: %w", err)
	}
	return config, nil
}

// List retrieves all configuration entries
func (r *SystemConfigRepository) List(ctx context.Context) ([]*models.SystemConfig, error) {
	query := `
		SELECT id, key, value, description, updated_at, updated_by
		FROM system_config
		ORDER BY key
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list configs: %w", err)
	}
	defer rows.Close()

	var configs []*models.SystemConfig
	for rows.Next() {
		config := &models.SystemConfig{}
		err := rows.Scan(
			&config.ID,
			&config.Key,
			&config.Value,
			&config.Description,
			&config.UpdatedAt,
			&config.UpdatedBy,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan config: %w", err)
		}
		configs = append(configs, config)
	}

	return configs, nil
}

// Delete removes a configuration entry
func (r *SystemConfigRepository) Delete(ctx context.Context, key string) error {
	query := `DELETE FROM system_config WHERE key = $1`
	result, err := r.db.ExecContext(ctx, query, key)
	if err != nil {
		return fmt.Errorf("failed to delete config: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("config not found: %s", key)
	}

	return nil
}

// IsRegistrationAllowed checks if new user registration is enabled
func (r *SystemConfigRepository) IsRegistrationAllowed(ctx context.Context) bool {
	allowed, err := r.GetBool(ctx, "allow_registration")
	if err != nil {
		// Default to true if config not found (for backward compatibility)
		return true
	}
	return allowed
}
