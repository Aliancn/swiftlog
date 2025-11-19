package handlers

import (
	"database/sql"
	"net/http"

	"github.com/aliancn/swiftlog/backend/internal/models"
	"github.com/aliancn/swiftlog/backend/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// SettingsHandler handles settings-related API requests
type SettingsHandler struct {
	settingsRepo *repository.SettingsRepository
	projectRepo  *repository.ProjectRepository
}

// NewSettingsHandler creates a new settings handler
func NewSettingsHandler(
	settingsRepo *repository.SettingsRepository,
	projectRepo  *repository.ProjectRepository,
) *SettingsHandler {
	return &SettingsHandler{
		settingsRepo: settingsRepo,
		projectRepo:  projectRepo,
	}
}

// GetUserSettings returns current user's settings
// GET /api/v1/settings
func (h *SettingsHandler) GetUserSettings(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	settings, err := h.settingsRepo.GetUserSettings(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user settings"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"settings":     settings,
		"has_api_key": settings.AIAPIKey.Valid && settings.AIAPIKey.String != "",
	})
}

// UpdateUserSettings updates current user's settings
// PUT /api/v1/settings
func (h *SettingsHandler) UpdateUserSettings(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	var req struct {
		AIEnabled             bool                     `json:"ai_enabled"`
		AIBaseURL             string                   `json:"ai_base_url" binding:"required"`
		AIAPIKey              *string                  `json:"ai_api_key"` // null = don't update
		AIModel               string                   `json:"ai_model" binding:"required"`
		AIMaxTokens           int                      `json:"ai_max_tokens" binding:"required,min=1"`
		AIAutoAnalyze         bool                     `json:"ai_auto_analyze"`
		AIMaxLogLines         int                      `json:"ai_max_log_lines" binding:"required,min=1"`
		AILogTruncateStrategy models.TruncateStrategy `json:"ai_log_truncate_strategy" binding:"required"`
		AISystemPrompt        string                   `json:"ai_system_prompt" binding:"required"`
		AIMaxConcurrent       int                      `json:"ai_max_concurrent" binding:"required,min=1,max=10"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get current settings to preserve API key if not provided
	current, err := h.settingsRepo.GetUserSettings(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch current settings"})
		return
	}

	settings := &models.UserSettings{
		UserID:                userID,
		AIEnabled:             req.AIEnabled,
		AIBaseURL:             req.AIBaseURL,
		AIAPIKey:              current.AIAPIKey, // Keep existing key
		AIModel:               req.AIModel,
		AIMaxTokens:           req.AIMaxTokens,
		AIAutoAnalyze:         req.AIAutoAnalyze,
		AIMaxLogLines:         req.AIMaxLogLines,
		AILogTruncateStrategy: req.AILogTruncateStrategy,
		AISystemPrompt:        req.AISystemPrompt,
		AIMaxConcurrent:       req.AIMaxConcurrent,
	}

	// Update API key if provided
	if req.AIAPIKey != nil {
		if *req.AIAPIKey == "" {
			settings.AIAPIKey = sql.NullString{Valid: false}
		} else {
			settings.AIAPIKey = sql.NullString{String: *req.AIAPIKey, Valid: true}
		}
	}

	if err := h.settingsRepo.UpdateUserSettings(c.Request.Context(), settings); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update settings"})
		return
	}

	// Fetch updated settings
	updated, _ := h.settingsRepo.GetUserSettings(c.Request.Context(), userID)
	c.JSON(http.StatusOK, gin.H{
		"settings":     updated,
		"has_api_key": updated.AIAPIKey.Valid && updated.AIAPIKey.String != "",
	})
}

// GetProjectSettings returns project-specific settings
// GET /api/v1/projects/:id/settings
func (h *SettingsHandler) GetProjectSettings(c *gin.Context) {
	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	// Verify project exists and user has access
	userIDVal, exists := c.Get("user_id")
	if exists {
		userID := userIDVal.(uuid.UUID)
		project, err := h.projectRepo.GetByID(c.Request.Context(), projectID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
			return
		}
		if project.UserID != userID {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			return
		}
	}

	settings, err := h.settingsRepo.GetProjectSettings(c.Request.Context(), projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch project settings"})
		return
	}

	if settings == nil {
		c.JSON(http.StatusOK, gin.H{
			"settings":     nil,
			"has_api_key": false,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"settings":     settings,
		"has_api_key": settings.AIAPIKey.Valid && settings.AIAPIKey.String != "",
	})
}

// UpdateProjectSettings updates project-specific settings
// PUT /api/v1/projects/:id/settings
func (h *SettingsHandler) UpdateProjectSettings(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	// Verify project exists and user has access
	project, err := h.projectRepo.GetByID(c.Request.Context(), projectID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}
	if project.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	var req struct {
		AIEnabled             *bool                     `json:"ai_enabled"`
		AIBaseURL             *string                   `json:"ai_base_url"`
		AIAPIKey              *string                   `json:"ai_api_key"`
		AIModel               *string                   `json:"ai_model"`
		AIMaxTokens           *int                      `json:"ai_max_tokens"`
		AIAutoAnalyze         *bool                     `json:"ai_auto_analyze"`
		AIMaxLogLines         *int                      `json:"ai_max_log_lines"`
		AILogTruncateStrategy *models.TruncateStrategy `json:"ai_log_truncate_strategy"`
		AISystemPrompt        *string                   `json:"ai_system_prompt"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	settings := &models.ProjectSettings{
		ProjectID:             projectID,
		AIEnabled:             req.AIEnabled,
		AIBaseURL:             req.AIBaseURL,
		AIModel:               req.AIModel,
		AIMaxTokens:           req.AIMaxTokens,
		AIAutoAnalyze:         req.AIAutoAnalyze,
		AIMaxLogLines:         req.AIMaxLogLines,
		AILogTruncateStrategy: req.AILogTruncateStrategy,
		AISystemPrompt:        req.AISystemPrompt,
	}

	// Handle API key
	if req.AIAPIKey != nil {
		if *req.AIAPIKey == "" {
			settings.AIAPIKey = sql.NullString{Valid: false}
		} else {
			settings.AIAPIKey = sql.NullString{String: *req.AIAPIKey, Valid: true}
		}
	}

	if err := h.settingsRepo.UpsertProjectSettings(c.Request.Context(), settings); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update project settings"})
		return
	}

	// Fetch updated settings
	updated, _ := h.settingsRepo.GetProjectSettings(c.Request.Context(), projectID)
	c.JSON(http.StatusOK, gin.H{
		"settings":     updated,
		"has_api_key": updated != nil && updated.AIAPIKey.Valid && updated.AIAPIKey.String != "",
	})
}

// DeleteProjectSettings deletes project-specific settings
// DELETE /api/v1/projects/:id/settings
func (h *SettingsHandler) DeleteProjectSettings(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	// Verify project exists and user has access
	project, err := h.projectRepo.GetByID(c.Request.Context(), projectID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}
	if project.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	if err := h.settingsRepo.DeleteProjectSettings(c.Request.Context(), projectID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete project settings"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Project settings deleted, reverted to user defaults"})
}

// GetEffectiveSettings returns the effective settings for a project (merged)
// GET /api/v1/projects/:id/settings/effective
func (h *SettingsHandler) GetEffectiveSettings(c *gin.Context) {
	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	// Verify project exists and user has access
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userID := userIDVal.(uuid.UUID)
	project, err := h.projectRepo.GetByID(c.Request.Context(), projectID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}
	if project.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	settings, err := h.settingsRepo.GetEffectiveSettings(c.Request.Context(), projectID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch effective settings"})
		return
	}

	c.JSON(http.StatusOK, settings)
}
