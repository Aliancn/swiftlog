package handlers

import (
	"net/http"
	"strconv"

	"github.com/aliancn/swiftlog/backend/internal/queue"
	"github.com/aliancn/swiftlog/backend/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// StatusHandler handles system status-related API requests
type StatusHandler struct {
	logRunRepo *repository.LogRunRepository
	taskQueue  *queue.Queue
}

// NewStatusHandler creates a new status handler
func NewStatusHandler(
	logRunRepo *repository.LogRunRepository,
	taskQueue *queue.Queue,
) *StatusHandler {
	return &StatusHandler{
		logRunRepo: logRunRepo,
		taskQueue:  taskQueue,
	}
}

// GetStatistics returns statistics for the authenticated user
// GET /api/v1/status/statistics
func (h *StatusHandler) GetStatistics(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	stats, err := h.logRunRepo.GetStatusStatisticsByUser(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch statistics"})
		return
	}

	// Get queue length
	queueLength, err := h.taskQueue.GetQueueLength(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch queue length"})
		return
	}

	response := gin.H{
		"run_statistics": gin.H{
			"running":   stats.RunningCount,
			"completed": stats.CompletedCount,
			"failed":    stats.FailedCount,
			"aborted":   stats.AbortedCount,
			"total":     stats.RunningCount + stats.CompletedCount + stats.FailedCount + stats.AbortedCount,
		},
		"ai_statistics": gin.H{
			"pending":    stats.AIPendingCount,
			"processing": stats.AIProcessingCount,
			"completed":  stats.AICompletedCount,
			"failed":     stats.AIFailedCount,
			"total":      stats.AIPendingCount + stats.AIProcessingCount + stats.AICompletedCount + stats.AIFailedCount,
		},
		"queue_length": queueLength,
	}

	c.JSON(http.StatusOK, response)
}

// GetRecentRuns returns recent log runs for the authenticated user
// GET /api/v1/status/recent
func (h *StatusHandler) GetRecentRuns(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	// Parse limit parameter
	// Relaxed for better usability
	limit := 50
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 200 {
			limit = l
		}
	}

	runs, err := h.logRunRepo.ListRecentRunsByUser(c.Request.Context(), userID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch recent runs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  runs,
		"total": len(runs),
		"limit": limit,
	})
}

// ArchiveCompletedRuns archives all completed/failed/aborted runs for the authenticated user
// POST /api/v1/status/archive-completed
func (h *StatusHandler) ArchiveCompletedRuns(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	rowsAffected, err := h.logRunRepo.ArchiveCompletedRunsByUser(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to archive completed runs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":        "Completed runs archived successfully",
		"archived_count": rowsAffected,
	})
}
