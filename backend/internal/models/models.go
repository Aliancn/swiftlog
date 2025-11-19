package models

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// User represents an authenticated user of the platform
type User struct {
	ID           uuid.UUID `json:"id" db:"id"`
	Username     string    `json:"username" db:"username"`
	PasswordHash string    `json:"-" db:"password_hash"` // Never expose in JSON
	IsAdmin      bool      `json:"is_admin" db:"is_admin"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

// APIToken stores API tokens for authenticating the CLI and other clients
type APIToken struct {
	ID        uuid.UUID `json:"id" db:"id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	TokenHash string    `json:"-" db:"token_hash"` // Never expose in JSON
	Name      string    `json:"name" db:"name"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// Project is a top-level container for organizing logs, owned by a user
type Project struct {
	ID        uuid.UUID `json:"id" db:"id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	Name      string    `json:"name" db:"name"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// LogGroup is an organizational unit within a project
type LogGroup struct {
	ID        uuid.UUID `json:"id" db:"id"`
	ProjectID uuid.UUID `json:"project_id" db:"project_id"`
	Name      string    `json:"name" db:"name"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// RunStatus represents the execution status of a script
type RunStatus string

const (
	RunStatusRunning   RunStatus = "running"
	RunStatusCompleted RunStatus = "completed"
	RunStatusFailed    RunStatus = "failed"
	RunStatusAborted   RunStatus = "aborted"
)

// AIStatus represents the status of AI report generation
type AIStatus string

const (
	AIStatusNone       AIStatus = "none"       // AI analysis is disabled
	AIStatusPending    AIStatus = "pending"
	AIStatusProcessing AIStatus = "processing"
	AIStatusCompleted  AIStatus = "completed"
	AIStatusFailed     AIStatus = "failed"
)

// LogRun represents a single execution of a logged script
type LogRun struct {
	ID        uuid.UUID      `json:"id" db:"id"`
	GroupID   uuid.UUID      `json:"group_id" db:"group_id"`
	StartTime time.Time      `json:"start_time" db:"start_time"`
	EndTime   sql.NullTime   `json:"-" db:"end_time"`
	Status    RunStatus      `json:"status" db:"status"`
	ExitCode  sql.NullInt32  `json:"-" db:"exit_code"`
	AIReport  sql.NullString `json:"-" db:"ai_report"`
	AIStatus  AIStatus       `json:"ai_status" db:"ai_status"`
	CreatedAt time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt time.Time      `json:"updated_at" db:"updated_at"`
}

// MarshalJSON implements custom JSON serialization for LogRun
func (r LogRun) MarshalJSON() ([]byte, error) {
	type Alias LogRun
	return json.Marshal(&struct {
		*Alias
		EndTime  *time.Time `json:"end_time,omitempty"`
		ExitCode *int32     `json:"exit_code,omitempty"`
		AIReport *string    `json:"ai_report,omitempty"`
	}{
		Alias:    (*Alias)(&r),
		EndTime:  nullTimeToPtr(r.EndTime),
		ExitCode: nullInt32ToPtr(r.ExitCode),
		AIReport: nullStringToPtr(r.AIReport),
	})
}

func nullTimeToPtr(nt sql.NullTime) *time.Time {
	if nt.Valid {
		return &nt.Time
	}
	return nil
}

func nullInt32ToPtr(ni sql.NullInt32) *int32 {
	if ni.Valid {
		return &ni.Int32
	}
	return nil
}

func nullStringToPtr(ns sql.NullString) *string {
	if ns.Valid {
		return &ns.String
	}
	return nil
}

// LogLine represents a single log line (stored in Loki, not PostgreSQL)
type LogLine struct {
	Timestamp time.Time `json:"timestamp"`
	Level     string    `json:"level"` // "stdout" or "stderr"
	Content   string    `json:"content"`
}

// StatusStatistics provides overall statistics for runs and AI tasks
type StatusStatistics struct {
	// Run statistics
	RunningCount   int `json:"running_count"`
	CompletedCount int `json:"completed_count"`
	FailedCount    int `json:"failed_count"`
	AbortedCount   int `json:"aborted_count"`

	// AI analysis statistics
	AIPendingCount    int `json:"ai_pending_count"`
	AIProcessingCount int `json:"ai_processing_count"`
	AICompletedCount  int `json:"ai_completed_count"`
	AIFailedCount     int `json:"ai_failed_count"`
}

// TruncateStrategy defines how to handle log truncation
type TruncateStrategy string

const (
	TruncateHead  TruncateStrategy = "head"  // Keep first N lines
	TruncateTail  TruncateStrategy = "tail"  // Keep last N lines
	TruncateSmart TruncateStrategy = "smart" // Keep head + tail with summary
)

// UserSettings represents user-specific configuration
type UserSettings struct {
	ID     uuid.UUID `json:"id" db:"id"`
	UserID uuid.UUID `json:"user_id" db:"user_id"`

	// AI Configuration
	AIEnabled             bool             `json:"ai_enabled" db:"ai_enabled"`
	AIBaseURL             string           `json:"ai_base_url" db:"ai_base_url"`
	AIAPIKey              sql.NullString   `json:"-" db:"ai_api_key"` // Never expose in JSON
	AIModel               string           `json:"ai_model" db:"ai_model"`
	AIMaxTokens           int              `json:"ai_max_tokens" db:"ai_max_tokens"`
	AIAutoAnalyze         bool             `json:"ai_auto_analyze" db:"ai_auto_analyze"`
	AIMaxLogLines         int              `json:"ai_max_log_lines" db:"ai_max_log_lines"`
	AILogTruncateStrategy TruncateStrategy `json:"ai_log_truncate_strategy" db:"ai_log_truncate_strategy"`
	AISystemPrompt        string           `json:"ai_system_prompt" db:"ai_system_prompt"`
	AIMaxConcurrent       int              `json:"ai_max_concurrent" db:"ai_max_concurrent"`

	// Metadata
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// ProjectSettings represents project-specific configuration (overrides user settings)
type ProjectSettings struct {
	ID        uuid.UUID `json:"id" db:"id"`
	ProjectID uuid.UUID `json:"project_id" db:"project_id"`

	// AI Configuration (nullable = inherit from user settings)
	AIEnabled             *bool             `json:"ai_enabled,omitempty" db:"ai_enabled"`
	AIBaseURL             *string           `json:"ai_base_url,omitempty" db:"ai_base_url"`
	AIAPIKey              sql.NullString    `json:"-" db:"ai_api_key"`
	AIModel               *string           `json:"ai_model,omitempty" db:"ai_model"`
	AIMaxTokens           *int              `json:"ai_max_tokens,omitempty" db:"ai_max_tokens"`
	AIAutoAnalyze         *bool             `json:"ai_auto_analyze,omitempty" db:"ai_auto_analyze"`
	AIMaxLogLines         *int              `json:"ai_max_log_lines,omitempty" db:"ai_max_log_lines"`
	AILogTruncateStrategy *TruncateStrategy `json:"ai_log_truncate_strategy,omitempty" db:"ai_log_truncate_strategy"`
	AISystemPrompt        *string           `json:"ai_system_prompt,omitempty" db:"ai_system_prompt"`
	AIMaxConcurrent       *int              `json:"ai_max_concurrent,omitempty" db:"ai_max_concurrent"`

	// Metadata
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// EffectiveSettings represents the merged configuration (project overrides user)
type EffectiveSettings struct {
	// AI Configuration
	AIEnabled             bool             `json:"ai_enabled"`
	AIBaseURL             string           `json:"ai_base_url"`
	AIAPIKey              string           `json:"-"` // Never expose
	AIModel               string           `json:"ai_model"`
	AIMaxTokens           int              `json:"ai_max_tokens"`
	AIAutoAnalyze         bool             `json:"ai_auto_analyze"`
	AIMaxLogLines         int              `json:"ai_max_log_lines"`
	AILogTruncateStrategy TruncateStrategy `json:"ai_log_truncate_strategy"`
	AISystemPrompt        string           `json:"ai_system_prompt"`
	AIMaxConcurrent       int              `json:"ai_max_concurrent"`

	// Source indicator
	Source string `json:"source"` // "user", "project", "merged"
}
