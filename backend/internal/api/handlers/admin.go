package handlers

import (
	"net/http"

	"github.com/aliancn/swiftlog/backend/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// AdminHandler handles admin-only API requests
type AdminHandler struct {
	userRepo         *repository.UserRepository
	systemConfigRepo *repository.SystemConfigRepository
}

// NewAdminHandler creates a new admin handler
func NewAdminHandler(
	userRepo *repository.UserRepository,
	systemConfigRepo *repository.SystemConfigRepository,
) *AdminHandler {
	return &AdminHandler{
		userRepo:         userRepo,
		systemConfigRepo: systemConfigRepo,
	}
}

// ============================================================================
// User Management
// ============================================================================

// ListUsers returns all users (admin only)
// GET /api/v1/admin/users
func (h *AdminHandler) ListUsers(c *gin.Context) {
	users, err := h.userRepo.ListAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"users": users})
}

// GetUserStats returns user statistics
// GET /api/v1/admin/users/stats
func (h *AdminHandler) GetUserStats(c *gin.Context) {
	users, err := h.userRepo.ListAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}

	stats := gin.H{
		"total":    len(users),
		"active":   0,
		"inactive": 0,
		"admins":   0,
	}

	for _, user := range users {
		if user.IsActive {
			stats["active"] = stats["active"].(int) + 1
		} else {
			stats["inactive"] = stats["inactive"].(int) + 1
		}
		if user.IsAdmin {
			stats["admins"] = stats["admins"].(int) + 1
		}
	}

	c.JSON(http.StatusOK, stats)
}

// UpdateUserStatus updates a user's active status
// PUT /api/v1/admin/users/:id/status
func (h *AdminHandler) UpdateUserStatus(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req struct {
		IsActive bool `json:"is_active"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Prevent disabling own account
	currentUserID := c.MustGet("user_id").(uuid.UUID)
	if userID == currentUserID && !req.IsActive {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot disable your own account"})
		return
	}

	if err := h.userRepo.UpdateStatus(c.Request.Context(), userID, req.IsActive); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User status updated successfully"})
}

// UpdateUserAdminStatus updates a user's admin status
// PUT /api/v1/admin/users/:id/admin
func (h *AdminHandler) UpdateUserAdminStatus(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req struct {
		IsAdmin bool `json:"is_admin"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Prevent removing own admin privileges
	currentUserID := c.MustGet("user_id").(uuid.UUID)
	if userID == currentUserID && !req.IsAdmin {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot remove your own admin privileges"})
		return
	}

	if err := h.userRepo.UpdateAdminStatus(c.Request.Context(), userID, req.IsAdmin); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update admin status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Admin status updated successfully"})
}

// DeleteUser deletes a user account
// DELETE /api/v1/admin/users/:id
func (h *AdminHandler) DeleteUser(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Prevent deleting own account
	currentUserID := c.MustGet("user_id").(uuid.UUID)
	if userID == currentUserID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot delete your own account"})
		return
	}

	if err := h.userRepo.Delete(c.Request.Context(), userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

// ============================================================================
// System Configuration
// ============================================================================

// GetSystemConfig returns a single system configuration
// GET /api/v1/admin/config/:key
func (h *AdminHandler) GetSystemConfig(c *gin.Context) {
	key := c.Param("key")
	config, err := h.systemConfigRepo.Get(c.Request.Context(), key)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Configuration not found"})
		return
	}

	c.JSON(http.StatusOK, config)
}

// ListSystemConfig returns all system configurations
// GET /api/v1/admin/config
func (h *AdminHandler) ListSystemConfig(c *gin.Context) {
	configs, err := h.systemConfigRepo.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch configurations"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"configs": configs})
}

// UpdateSystemConfig updates a system configuration
// PUT /api/v1/admin/config/:key
func (h *AdminHandler) UpdateSystemConfig(c *gin.Context) {
	key := c.Param("key")
	userID := c.MustGet("user_id").(uuid.UUID)

	var req struct {
		Value string `json:"value" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	config, err := h.systemConfigRepo.Set(c.Request.Context(), key, req.Value, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update configuration"})
		return
	}

	c.JSON(http.StatusOK, config)
}

// DeleteSystemConfig deletes a system configuration
// DELETE /api/v1/admin/config/:key
func (h *AdminHandler) DeleteSystemConfig(c *gin.Context) {
	key := c.Param("key")

	// Prevent deleting critical configs
	criticalConfigs := []string{"allow_registration"}
	for _, critical := range criticalConfigs {
		if key == critical {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot delete critical configuration"})
			return
		}
	}

	if err := h.systemConfigRepo.Delete(c.Request.Context(), key); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete configuration"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Configuration deleted successfully"})
}

// GetSystemStats returns overall system statistics
// GET /api/v1/admin/stats
func (h *AdminHandler) GetSystemStats(c *gin.Context) {
	// Get user count
	userCount, err := h.userRepo.Count(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch statistics"})
		return
	}

	// Get registration setting
	allowRegistration := h.systemConfigRepo.IsRegistrationAllowed(c.Request.Context())

	stats := gin.H{
		"users": gin.H{
			"total": userCount,
		},
		"system": gin.H{
			"allow_registration": allowRegistration,
		},
	}

	c.JSON(http.StatusOK, stats)
}
