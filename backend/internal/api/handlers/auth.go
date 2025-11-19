package handlers

import (
	"net/http"

	"github.com/aliancn/swiftlog/backend/internal/auth"
	"github.com/aliancn/swiftlog/backend/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// AuthHandler handles authentication-related API requests
type AuthHandler struct {
	userRepo     *repository.UserRepository
	settingsRepo *repository.SettingsRepository
	tokenService *auth.TokenService
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(
	userRepo *repository.UserRepository,
	settingsRepo *repository.SettingsRepository,
	tokenService *auth.TokenService,
) *AuthHandler {
	return &AuthHandler{
		userRepo:     userRepo,
		settingsRepo: settingsRepo,
		tokenService: tokenService,
	}
}

// Login authenticates a user and returns a session token
// POST /api/v1/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Get user from database
	user, err := h.userRepo.GetByUsername(c.Request.Context(), req.Username)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	// Verify password
	if err := auth.VerifyPassword(req.Password, user.PasswordHash); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	// Create API token for session
	rawToken, apiToken, err := h.tokenService.CreateToken(c.Request.Context(), user.ID, "web-session")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create session"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": rawToken,
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"is_admin": user.IsAdmin,
		},
		"token_info": apiToken,
	})
}

// Register creates a new user account
// POST /api/v1/auth/register
func (h *AuthHandler) Register(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Hash password
	passwordHash, err := auth.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create user
	user, err := h.userRepo.Create(c.Request.Context(), req.Username, passwordHash, false)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	// Create default settings for new user
	_, err = h.settingsRepo.CreateDefaultUserSettings(c.Request.Context(), user.ID)
	if err != nil {
		// Log error but don't fail registration
		// User can configure settings later
		c.Request.Context().Value("logger")
	}

	// Create API token for session
	rawToken, apiToken, err := h.tokenService.CreateToken(c.Request.Context(), user.ID, "web-session")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create session"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"token": rawToken,
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"is_admin": user.IsAdmin,
		},
		"token_info": apiToken,
	})
}

// GetCurrentUser returns the currently authenticated user
// GET /api/v1/auth/me
func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	user, err := h.userRepo.GetByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":       user.ID,
		"username": user.Username,
		"is_admin": user.IsAdmin,
		"created_at": user.CreatedAt,
	})
}

// ListTokens returns all API tokens for the current user
// GET /api/v1/auth/tokens
func (h *AuthHandler) ListTokens(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	tokens, err := h.tokenService.ListTokensByUserID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tokens"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"tokens": tokens})
}

// CreateToken creates a new API token for the current user
// POST /api/v1/auth/tokens
func (h *AuthHandler) CreateToken(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	var req struct {
		Name string `json:"name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	rawToken, apiToken, err := h.tokenService.CreateToken(c.Request.Context(), userID, req.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create token"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"token": rawToken,
		"token_info": apiToken,
	})
}

// DeleteToken deletes an API token
// DELETE /api/v1/auth/tokens/:id
func (h *AuthHandler) DeleteToken(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	tokenID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid token ID"})
		return
	}

	// Verify token belongs to user
	token, err := h.tokenService.GetTokenByID(c.Request.Context(), tokenID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Token not found"})
		return
	}

	if token.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	if err := h.tokenService.RevokeToken(c.Request.Context(), tokenID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Token deleted successfully"})
}

// ListUsers returns all users (admin only)
// GET /api/v1/auth/users
func (h *AuthHandler) ListUsers(c *gin.Context) {
	// Check if user is admin
	userID := c.MustGet("user_id").(uuid.UUID)
	currentUser, err := h.userRepo.GetByID(c.Request.Context(), userID)
	if err != nil || !currentUser.IsAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
		return
	}

	users, err := h.userRepo.ListAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"users": users})
}
