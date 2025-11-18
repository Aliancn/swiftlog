package handlers

import (
	"net/http"
	"strconv"

	"github.com/aliancn/swiftlog/backend/internal/loki"
	"github.com/aliancn/swiftlog/backend/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RunsHandler handles log run-related API requests
type RunsHandler struct {
	logRunRepo  *repository.LogRunRepository
	groupRepo   *repository.LogGroupRepository
	projectRepo *repository.ProjectRepository
	lokiClient  *loki.Client
}

// NewRunsHandler creates a new runs handler
func NewRunsHandler(
	logRunRepo *repository.LogRunRepository,
	groupRepo *repository.LogGroupRepository,
	projectRepo *repository.ProjectRepository,
	lokiClient *loki.Client,
) *RunsHandler {
	return &RunsHandler{
		logRunRepo:  logRunRepo,
		groupRepo:   groupRepo,
		projectRepo: projectRepo,
		lokiClient:  lokiClient,
	}
}

// ListRuns returns runs for a specific group
// GET /api/v1/groups/:id/runs
func (h *RunsHandler) ListRuns(c *gin.Context) {
	groupID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid group ID"})
		return
	}

	// Verify group ownership if authenticated
	userIDVal, exists := c.Get("user_id")
	if exists {
		userID := userIDVal.(uuid.UUID)
		group, err := h.groupRepo.GetByID(c.Request.Context(), groupID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Group not found"})
			return
		}

		project, err := h.projectRepo.GetByID(c.Request.Context(), group.ProjectID)
		if err != nil || project.UserID != userID {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			return
		}
	}

	// Parse pagination params
	limit := 50
	offset := 0
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	runs, err := h.logRunRepo.ListByGroupID(c.Request.Context(), groupID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch runs"})
		return
	}

	// Return in format expected by frontend
	c.JSON(http.StatusOK, gin.H{
		"data":   runs,
		"total":  len(runs), // TODO: Get actual total count from database
		"limit":  limit,
		"offset": offset,
	})
}

// GetRun returns a specific run by ID
// GET /api/v1/runs/:id
func (h *RunsHandler) GetRun(c *gin.Context) {
	runID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid run ID"})
		return
	}

	run, err := h.logRunRepo.GetByID(c.Request.Context(), runID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Run not found"})
		return
	}

	// Verify ownership if authenticated
	userIDVal, exists := c.Get("user_id")
	if exists {
		userID := userIDVal.(uuid.UUID)
		group, err := h.groupRepo.GetByID(c.Request.Context(), run.GroupID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify ownership"})
			return
		}

		project, err := h.projectRepo.GetByID(c.Request.Context(), group.ProjectID)
		if err != nil || project.UserID != userID {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			return
		}
	}

	c.JSON(http.StatusOK, run)
}

// GetRunLogs returns logs for a specific run from Loki
// GET /api/v1/runs/:id/logs
func (h *RunsHandler) GetRunLogs(c *gin.Context) {
	runID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid run ID"})
		return
	}

	// Verify ownership if authenticated
	userIDVal, exists := c.Get("user_id")
	if exists {
		userID := userIDVal.(uuid.UUID)
		run, err := h.logRunRepo.GetByID(c.Request.Context(), runID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Run not found"})
			return
		}

		group, err := h.groupRepo.GetByID(c.Request.Context(), run.GroupID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify ownership"})
			return
		}

		project, err := h.projectRepo.GetByID(c.Request.Context(), group.ProjectID)
		if err != nil || project.UserID != userID {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			return
		}
	}

	// Query logs from Loki
	logs, err := h.lokiClient.QueryLogs(c.Request.Context(), runID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch logs"})
		return
	}

	c.JSON(http.StatusOK, logs)
}

// TriggerAIAnalysis triggers AI analysis for a run
// POST /api/v1/runs/:id/analyze
func (h *RunsHandler) TriggerAIAnalysis(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	runID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid run ID"})
		return
	}

	// Verify ownership
	run, err := h.logRunRepo.GetByID(c.Request.Context(), runID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Run not found"})
		return
	}

	group, err := h.groupRepo.GetByID(c.Request.Context(), run.GroupID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify ownership"})
		return
	}

	project, err := h.projectRepo.GetByID(c.Request.Context(), group.ProjectID)
	if err != nil || project.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	// TODO: Publish message to Redis for AI worker to pick up
	// For now, just return success
	c.JSON(http.StatusAccepted, gin.H{
		"message": "AI analysis queued",
		"run_id":  runID.String(),
	})
}
