package handlers

import (
	"net/http"

	"github.com/aliancn/swiftlog/backend/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ProjectsHandler handles project-related API requests
type ProjectsHandler struct {
	projectRepo *repository.ProjectRepository
	groupRepo   *repository.LogGroupRepository
}

// NewProjectsHandler creates a new projects handler
func NewProjectsHandler(projectRepo *repository.ProjectRepository, groupRepo *repository.LogGroupRepository) *ProjectsHandler {
	return &ProjectsHandler{
		projectRepo: projectRepo,
		groupRepo:   groupRepo,
	}
}

// ListProjects returns all projects for the authenticated user (or all projects if no auth)
// GET /api/v1/projects
func (h *ProjectsHandler) ListProjects(c *gin.Context) {
	// Get user ID if authenticated, otherwise return all projects
	userIDVal, exists := c.Get("user_id")

	var projects interface{}
	var err error

	if exists {
		userID := userIDVal.(uuid.UUID)
		projects, err = h.projectRepo.ListByUserID(c.Request.Context(), userID)
	} else {
		// Development mode: return all projects
		projects, err = h.projectRepo.ListAll(c.Request.Context())
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch projects"})
		return
	}

	c.JSON(http.StatusOK, projects)
}

// GetProject returns a specific project by ID
// GET /api/v1/projects/:id
func (h *ProjectsHandler) GetProject(c *gin.Context) {
	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	project, err := h.projectRepo.GetByID(c.Request.Context(), projectID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	// Verify ownership if authenticated
	userIDVal, exists := c.Get("user_id")
	if exists {
		userID := userIDVal.(uuid.UUID)
		if project.UserID != userID {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			return
		}
	}

	c.JSON(http.StatusOK, project)
}

// CreateProject creates a new project
// POST /api/v1/projects
func (h *ProjectsHandler) CreateProject(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	var req struct {
		Name string `json:"name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	project, err := h.projectRepo.Create(c.Request.Context(), userID, req.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create project"})
		return
	}

	c.JSON(http.StatusCreated, project)
}

// GetProjectGroups returns all log groups for a specific project
// GET /api/v1/projects/:id/groups
func (h *ProjectsHandler) GetProjectGroups(c *gin.Context) {
	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	// Verify project exists
	project, err := h.projectRepo.GetByID(c.Request.Context(), projectID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	// Verify ownership if authenticated
	userIDVal, exists := c.Get("user_id")
	if exists {
		userID := userIDVal.(uuid.UUID)
		if project.UserID != userID {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			return
		}
	}

	// Get groups for this project
	groups, err := h.groupRepo.ListByProjectID(c.Request.Context(), projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch groups"})
		return
	}

	c.JSON(http.StatusOK, groups)
}
