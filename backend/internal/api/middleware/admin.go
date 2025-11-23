package middleware

import (
	"net/http"

	"github.com/aliancn/swiftlog/backend/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// AdminMiddleware creates a Gin middleware that requires admin privileges
// PERFORMANCE: Uses is_admin from JWT claims to avoid database query
// Note: is_active is verified by ActiveUserMiddleware which should run before this
func AdminMiddleware(userRepo *repository.UserRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get is_admin from JWT claims (performance optimization - no DB query)
		// Note: user_id is available in context if needed for future use
		isAdminVal, exists := c.Get("is_admin")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authentication"})
			c.Abort()
			return
		}

		isAdmin, ok := isAdminVal.(bool)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid admin status"})
			c.Abort()
			return
		}

		// Check if user is admin
		if !isAdmin {
			c.JSON(http.StatusForbidden, gin.H{"error": "Admin privileges required"})
			c.Abort()
			return
		}

		// OPTIONAL: For critical admin operations, verify current admin status from database
		// This is a security vs performance trade-off
		// Uncomment if you need real-time admin status verification:
		/*
		user, err := userRepo.GetByID(c.Request.Context(), userID.(uuid.UUID))
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
			c.Abort()
			return
		}
		if !user.IsAdmin || !user.IsActive {
			c.JSON(http.StatusForbidden, gin.H{"error": "Admin privileges revoked"})
			c.Abort()
			return
		}
		c.Set("user", user)
		*/

		c.Next()
	}
}

// ActiveUserMiddleware checks if the authenticated user is active
func ActiveUserMiddleware(userRepo *repository.UserRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user ID from context (set by AuthMiddleware)
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			c.Abort()
			return
		}

		// Get user from database
		user, err := userRepo.GetByID(c.Request.Context(), userID.(uuid.UUID))
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
			c.Abort()
			return
		}

		// Check if user is active
		if !user.IsActive {
			c.JSON(http.StatusForbidden, gin.H{"error": "Account is disabled"})
			c.Abort()
			return
		}

		c.Next()
	}
}
