package handlers

import (
	"net/http"

	"github.com/aliancn/swiftlog/backend/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GroupsHandler handles log group-related API requests
type GroupsHandler struct {
	groupRepo   *repository.LogGroupRepository
	projectRepo *repository.ProjectRepository
}

// NewGroupsHandler creates a new groups handler
func NewGroupsHandler(groupRepo *repository.LogGroupRepository, projectRepo *repository.ProjectRepository) *GroupsHandler {
	return &GroupsHandler{
		groupRepo:   groupRepo,
		projectRepo: projectRepo,
	}
}

// GetGroup returns a specific log group by ID
// GET /api/v1/groups/:id
func (h *GroupsHandler) GetGroup(c *gin.Context) {
	groupID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid group ID"})
		return
	}

	group, err := h.groupRepo.GetByID(c.Request.Context(), groupID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Group not found"})
		return
	}

	// Verify ownership if authenticated
	userIDVal, exists := c.Get("user_id")
	if exists {
		userID := userIDVal.(uuid.UUID)
		// Get the project to verify ownership
		project, err := h.projectRepo.GetByID(c.Request.Context(), group.ProjectID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
			return
		}
		if project.UserID != userID {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			return
		}
	}

	c.JSON(http.StatusOK, group)
}
