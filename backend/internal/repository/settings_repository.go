package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/aliancn/swiftlog/backend/internal/models"
	"github.com/google/uuid"
)

// SettingsRepository handles database operations for settings
type SettingsRepository struct {
	db *sql.DB
}

// NewSettingsRepository creates a new settings repository
func NewSettingsRepository(db *sql.DB) *SettingsRepository {
	return &SettingsRepository{db: db}
}

// GetUserSettings retrieves user-specific settings
func (r *SettingsRepository) GetUserSettings(ctx context.Context, userID uuid.UUID) (*models.UserSettings, error) {
	settings := &models.UserSettings{}
	query := `
		SELECT id, user_id, ai_enabled, ai_base_url, ai_api_key, ai_model, ai_max_tokens,
		       ai_auto_analyze, ai_max_log_lines, ai_log_truncate_strategy,
		       ai_system_prompt, created_at, updated_at
		FROM user_settings
		WHERE user_id = $1
	`
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&settings.ID,
		&settings.UserID,
		&settings.AIEnabled,
		&settings.AIBaseURL,
		&settings.AIAPIKey,
		&settings.AIModel,
		&settings.AIMaxTokens,
		&settings.AIAutoAnalyze,
		&settings.AIMaxLogLines,
		&settings.AILogTruncateStrategy,
		&settings.AISystemPrompt,
		&settings.CreatedAt,
		&settings.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user settings not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user settings: %w", err)
	}
	return settings, nil
}

// CreateDefaultUserSettings creates default settings for a new user
func (r *SettingsRepository) CreateDefaultUserSettings(ctx context.Context, userID uuid.UUID) (*models.UserSettings, error) {
	settings := &models.UserSettings{}
	query := `
		INSERT INTO user_settings (
			user_id, ai_enabled, ai_base_url, ai_model, ai_max_tokens,
			ai_auto_analyze, ai_max_log_lines, ai_log_truncate_strategy, ai_system_prompt
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, user_id, ai_enabled, ai_base_url, ai_api_key, ai_model, ai_max_tokens,
		          ai_auto_analyze, ai_max_log_lines, ai_log_truncate_strategy,
		          ai_system_prompt, created_at, updated_at
	`
	err := r.db.QueryRowContext(ctx, query,
		userID,
		true,                                         // ai_enabled
		"https://api.openai.com/v1",                  // ai_base_url
		"gpt-4o-mini",                                // ai_model
		500,                                          // ai_max_tokens
		false,                                        // ai_auto_analyze
		1000,                                         // ai_max_log_lines
		models.TruncateTail,                          // ai_log_truncate_strategy
		"You are a helpful assistant analyzing script execution logs. Identify errors, warnings, and provide actionable recommendations.", // ai_system_prompt
	).Scan(
		&settings.ID,
		&settings.UserID,
		&settings.AIEnabled,
		&settings.AIBaseURL,
		&settings.AIAPIKey,
		&settings.AIModel,
		&settings.AIMaxTokens,
		&settings.AIAutoAnalyze,
		&settings.AIMaxLogLines,
		&settings.AILogTruncateStrategy,
		&settings.AISystemPrompt,
		&settings.CreatedAt,
		&settings.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create default user settings: %w", err)
	}
	return settings, nil
}

// UpdateUserSettings updates user-specific settings
func (r *SettingsRepository) UpdateUserSettings(ctx context.Context, settings *models.UserSettings) error {
	query := `
		UPDATE user_settings
		SET ai_enabled = $1, ai_base_url = $2, ai_api_key = $3, ai_model = $4,
		    ai_max_tokens = $5, ai_auto_analyze = $6, ai_max_log_lines = $7,
		    ai_log_truncate_strategy = $8, ai_system_prompt = $9
		WHERE user_id = $10
	`
	_, err := r.db.ExecContext(ctx, query,
		settings.AIEnabled,
		settings.AIBaseURL,
		settings.AIAPIKey,
		settings.AIModel,
		settings.AIMaxTokens,
		settings.AIAutoAnalyze,
		settings.AIMaxLogLines,
		settings.AILogTruncateStrategy,
		settings.AISystemPrompt,
		settings.UserID,
	)
	if err != nil {
		return fmt.Errorf("failed to update user settings: %w", err)
	}
	return nil
}

// GetProjectSettings retrieves project-specific settings
func (r *SettingsRepository) GetProjectSettings(ctx context.Context, projectID uuid.UUID) (*models.ProjectSettings, error) {
	settings := &models.ProjectSettings{}
	query := `
		SELECT id, project_id, ai_enabled, ai_base_url, ai_api_key, ai_model,
		       ai_max_tokens, ai_auto_analyze, ai_max_log_lines,
		       ai_log_truncate_strategy, ai_system_prompt,
		       created_at, updated_at
		FROM project_settings
		WHERE project_id = $1
	`
	err := r.db.QueryRowContext(ctx, query, projectID).Scan(
		&settings.ID,
		&settings.ProjectID,
		&settings.AIEnabled,
		&settings.AIBaseURL,
		&settings.AIAPIKey,
		&settings.AIModel,
		&settings.AIMaxTokens,
		&settings.AIAutoAnalyze,
		&settings.AIMaxLogLines,
		&settings.AILogTruncateStrategy,
		&settings.AISystemPrompt,
		&settings.CreatedAt,
		&settings.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil // No project-specific settings (use user defaults)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get project settings: %w", err)
	}
	return settings, nil
}

// UpsertProjectSettings creates or updates project-specific settings
func (r *SettingsRepository) UpsertProjectSettings(ctx context.Context, settings *models.ProjectSettings) error {
	query := `
		INSERT INTO project_settings (
			project_id, ai_enabled, ai_base_url, ai_api_key, ai_model,
			ai_max_tokens, ai_auto_analyze, ai_max_log_lines,
			ai_log_truncate_strategy, ai_system_prompt
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (project_id) DO UPDATE SET
			ai_enabled = EXCLUDED.ai_enabled,
			ai_base_url = EXCLUDED.ai_base_url,
			ai_api_key = EXCLUDED.ai_api_key,
			ai_model = EXCLUDED.ai_model,
			ai_max_tokens = EXCLUDED.ai_max_tokens,
			ai_auto_analyze = EXCLUDED.ai_auto_analyze,
			ai_max_log_lines = EXCLUDED.ai_max_log_lines,
			ai_log_truncate_strategy = EXCLUDED.ai_log_truncate_strategy,
			ai_system_prompt = EXCLUDED.ai_system_prompt
		RETURNING id, created_at, updated_at
	`
	err := r.db.QueryRowContext(ctx, query,
		settings.ProjectID,
		settings.AIEnabled,
		settings.AIBaseURL,
		settings.AIAPIKey,
		settings.AIModel,
		settings.AIMaxTokens,
		settings.AIAutoAnalyze,
		settings.AIMaxLogLines,
		settings.AILogTruncateStrategy,
		settings.AISystemPrompt,
	).Scan(&settings.ID, &settings.CreatedAt, &settings.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to upsert project settings: %w", err)
	}
	return nil
}

// DeleteProjectSettings deletes project-specific settings (revert to user settings)
func (r *SettingsRepository) DeleteProjectSettings(ctx context.Context, projectID uuid.UUID) error {
	query := `DELETE FROM project_settings WHERE project_id = $1`
	_, err := r.db.ExecContext(ctx, query, projectID)
	if err != nil {
		return fmt.Errorf("failed to delete project settings: %w", err)
	}
	return nil
}

// GetEffectiveSettings retrieves the effective settings for a project (merged user + project)
func (r *SettingsRepository) GetEffectiveSettings(ctx context.Context, projectID, userID uuid.UUID) (*models.EffectiveSettings, error) {
	// Get user settings
	user, err := r.GetUserSettings(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Get project settings (may be nil)
	project, err := r.GetProjectSettings(ctx, projectID)
	if err != nil {
		return nil, err
	}

	// Merge settings (project overrides user)
	effective := &models.EffectiveSettings{
		AIEnabled:             user.AIEnabled,
		AIBaseURL:             user.AIBaseURL,
		AIAPIKey:              nullStringToString(user.AIAPIKey),
		AIModel:               user.AIModel,
		AIMaxTokens:           user.AIMaxTokens,
		AIAutoAnalyze:         user.AIAutoAnalyze,
		AIMaxLogLines:         user.AIMaxLogLines,
		AILogTruncateStrategy: user.AILogTruncateStrategy,
		AISystemPrompt:        user.AISystemPrompt,
		Source:                "user",
	}

	// Apply project overrides if they exist
	if project != nil {
		hasOverrides := false
		if project.AIEnabled != nil {
			effective.AIEnabled = *project.AIEnabled
			hasOverrides = true
		}
		if project.AIBaseURL != nil {
			effective.AIBaseURL = *project.AIBaseURL
			hasOverrides = true
		}
		if project.AIAPIKey.Valid {
			effective.AIAPIKey = project.AIAPIKey.String
			hasOverrides = true
		}
		if project.AIModel != nil {
			effective.AIModel = *project.AIModel
			hasOverrides = true
		}
		if project.AIMaxTokens != nil {
			effective.AIMaxTokens = *project.AIMaxTokens
			hasOverrides = true
		}
		if project.AIAutoAnalyze != nil {
			effective.AIAutoAnalyze = *project.AIAutoAnalyze
			hasOverrides = true
		}
		if project.AIMaxLogLines != nil {
			effective.AIMaxLogLines = *project.AIMaxLogLines
			hasOverrides = true
		}
		if project.AILogTruncateStrategy != nil {
			effective.AILogTruncateStrategy = *project.AILogTruncateStrategy
			hasOverrides = true
		}
		if project.AISystemPrompt != nil {
			effective.AISystemPrompt = *project.AISystemPrompt
			hasOverrides = true
		}

		if hasOverrides {
			effective.Source = "merged"
		}
	}

	return effective, nil
}

func nullStringToString(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}
