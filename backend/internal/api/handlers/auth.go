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
	userRepo         *repository.UserRepository
	settingsRepo     *repository.SettingsRepository
	tokenService     *auth.TokenService
	jwtService       *auth.JWTService
	systemConfigRepo *repository.SystemConfigRepository
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(
	userRepo *repository.UserRepository,
	settingsRepo *repository.SettingsRepository,
	tokenService *auth.TokenService,
	jwtService *auth.JWTService,
	systemConfigRepo *repository.SystemConfigRepository,
) *AuthHandler {
	return &AuthHandler{
		userRepo:         userRepo,
		settingsRepo:     settingsRepo,
		tokenService:     tokenService,
		jwtService:       jwtService,
		systemConfigRepo: systemConfigRepo,
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

	// Check if user is active
	if !user.IsActive {
		c.JSON(http.StatusForbidden, gin.H{"error": "Account is disabled. Please contact administrator."})
		return
	}

	// Update last login timestamp
	_ = h.userRepo.UpdateLastLogin(c.Request.Context(), user.ID)

	// Generate JWT token for web session
	jwtToken, err := h.jwtService.GenerateToken(user.ID, user.Username, user.IsAdmin)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create session"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": jwtToken,
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"is_admin": user.IsAdmin,
		},
	})
}

// Register creates a new user account
// POST /api/v1/auth/register
func (h *AuthHandler) Register(c *gin.Context) {
	// Check if registration is allowed
	if !h.systemConfigRepo.IsRegistrationAllowed(c.Request.Context()) {
		c.JSON(http.StatusForbidden, gin.H{"error": "User registration is currently disabled"})
		return
	}

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

	// Generate JWT token for web session
	jwtToken, err := h.jwtService.GenerateToken(user.ID, user.Username, user.IsAdmin)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create session"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"token": jwtToken,
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"is_admin": user.IsAdmin,
		},
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

// GetRegistrationStatus returns whether registration is allowed (public endpoint)
// GET /api/v1/auth/registration-status
func (h *AuthHandler) GetRegistrationStatus(c *gin.Context) {
	allowed := h.systemConfigRepo.IsRegistrationAllowed(c.Request.Context())
	c.JSON(http.StatusOK, gin.H{
		"registration_allowed": allowed,
	})
}
