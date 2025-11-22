package handlers

import (
	"net/http"

	"github.com/aliancn/swiftlog/backend/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ManagementHandler handles management-related API requests
type ManagementHandler struct {
	projectRepo  *repository.ProjectRepository
	groupRepo    *repository.LogGroupRepository
	logRunRepo   *repository.LogRunRepository
}

// NewManagementHandler creates a new management handler
func NewManagementHandler(
	projectRepo *repository.ProjectRepository,
	groupRepo *repository.LogGroupRepository,
	logRunRepo *repository.LogRunRepository,
) *ManagementHandler {
	return &ManagementHandler{
		projectRepo:  projectRepo,
		groupRepo:    groupRepo,
		logRunRepo:   logRunRepo,
	}
}

// UpdateProject updates a project's name
// PUT /api/v1/projects/:id
func (h *ManagementHandler) UpdateProject(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	// Verify project ownership
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
		Name string `json:"name" binding:"required,min=1,max=255"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updated, err := h.projectRepo.Update(c.Request.Context(), projectID, req.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update project"})
		return
	}

	c.JSON(http.StatusOK, updated)
}

// DeleteProject deletes a project and all related data
// DELETE /api/v1/projects/:id
func (h *ManagementHandler) DeleteProject(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	// Verify project ownership
	project, err := h.projectRepo.GetByID(c.Request.Context(), projectID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}
	if project.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	if err := h.projectRepo.Delete(c.Request.Context(), projectID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete project"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Project deleted successfully"})
}

// UpdateGroup updates a group's name
// PUT /api/v1/groups/:id
func (h *ManagementHandler) UpdateGroup(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	groupID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid group ID"})
		return
	}

	// Verify group ownership through project
	group, err := h.groupRepo.GetByID(c.Request.Context(), groupID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Group not found"})
		return
	}

	project, err := h.projectRepo.GetByID(c.Request.Context(), group.ProjectID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}
	if project.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	var req struct {
		Name string `json:"name" binding:"required,min=1,max=255"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updated, err := h.groupRepo.Update(c.Request.Context(), groupID, req.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update group"})
		return
	}

	c.JSON(http.StatusOK, updated)
}

// DeleteGroup deletes a group and all related runs
// DELETE /api/v1/groups/:id
func (h *ManagementHandler) DeleteGroup(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	groupID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid group ID"})
		return
	}

	// Verify group ownership through project
	group, err := h.groupRepo.GetByID(c.Request.Context(), groupID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Group not found"})
		return
	}

	project, err := h.projectRepo.GetByID(c.Request.Context(), group.ProjectID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}
	if project.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	if err := h.groupRepo.Delete(c.Request.Context(), groupID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete group"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Group deleted successfully"})
}

// DeleteRun deletes a single run record
// DELETE /api/v1/runs/:id
func (h *ManagementHandler) DeleteRun(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	runID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid run ID"})
		return
	}

	// Verify run ownership through group and project
	run, err := h.logRunRepo.GetByID(c.Request.Context(), runID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Run not found"})
		return
	}

	group, err := h.groupRepo.GetByID(c.Request.Context(), run.GroupID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Group not found"})
		return
	}

	project, err := h.projectRepo.GetByID(c.Request.Context(), group.ProjectID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}
	if project.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	if err := h.logRunRepo.Delete(c.Request.Context(), runID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete run"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Run deleted successfully"})
}
