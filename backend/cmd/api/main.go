package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/aliancn/swiftlog/backend/internal/api/handlers"
	"github.com/aliancn/swiftlog/backend/internal/api/middleware"
	"github.com/aliancn/swiftlog/backend/internal/auth"
	"github.com/aliancn/swiftlog/backend/internal/config"
	"github.com/aliancn/swiftlog/backend/internal/crypto"
	"github.com/aliancn/swiftlog/backend/internal/loki"
	"github.com/aliancn/swiftlog/backend/internal/queue"
	"github.com/aliancn/swiftlog/backend/internal/repository"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func main() {
	ctx := context.Background()

	// Load configuration from environment
	dbURL := config.GetEnv("DATABASE_URL", "postgres://swiftlog:changeme@localhost:5432/swiftlog?sslmode=disable")
	lokiURL := config.GetEnv("LOKI_URL", "http://localhost:3100")
	redisURL := config.GetEnv("REDIS_URL", "redis://localhost:6379")
	apiPort := config.GetEnv("API_PORT", "8080")
	environment := config.GetEnv("ENVIRONMENT", "development")
	corsOrigins := config.GetEnv("CORS_ORIGINS", "http://localhost:3000")
	jwtSecret := config.GetEnv("JWT_SECRET", "your-secret-key-change-this-in-production")
	// Default: 2 hours (more secure than 24h, balances security and UX)
	// Production recommendation: 15m-1h with refresh token mechanism
	jwtExpiration := config.GetEnvDuration("JWT_EXPIRATION", 2*time.Hour)

	if environment == "production" && jwtExpiration > 24*time.Hour {
		log.Printf("WARNING: JWT_EXPIRATION is set to %v (>24h) in production", jwtExpiration)
		log.Println("WARNING: Long-lived JWTs increase security risk. Consider using shorter expiration with refresh tokens.")
	}

	// Validate JWT secret strength
	if err := validateJWTSecret(jwtSecret, environment); err != nil {
		log.Fatalf("JWT secret validation failed: %v", err)
	}

	// Set Gin mode
	if environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize database connection
	log.Println("Connecting to database...")
	db, err := config.InitDatabase(ctx, dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize Loki client
	log.Println("Initializing Loki client...")
	lokiClient := loki.NewClient(&loki.Config{
		URL:     lokiURL,
		Timeout: 10 * time.Second,
	})

	// Initialize Redis client
	log.Println("Connecting to Redis...")
	redisOpt, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Fatalf("Failed to parse Redis URL: %v", err)
	}
	redisClient := redis.NewClient(redisOpt)
	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redisClient.Close()

	// Initialize task queue
	taskQueue := queue.NewQueue(redisClient)

	// Initialize encryption service for API key storage
	log.Println("Initializing encryption service...")
	encryptionKey, err := config.GetEncryptionKey("api")
	if err != nil {
		log.Fatalf("Failed to get encryption key: %v", err)
	}
	encryptionService, err := crypto.NewEncryptionService(encryptionKey)
	if err != nil {
		log.Fatalf("Failed to initialize encryption service: %v", err)
	}

	// Initialize repositories
	projectRepo := repository.NewProjectRepository(db.DB)
	groupRepo := repository.NewLogGroupRepository(db.DB)
	logRunRepo := repository.NewLogRunRepository(db.DB)
	userRepo := repository.NewUserRepository(db.DB)
	settingsRepo := repository.NewSettingsRepository(db.DB, encryptionService)
	systemConfigRepo := repository.NewSystemConfigRepository(db.DB)

	// Initialize auth services
	tokenService := auth.NewTokenService(db.DB)
	jwtService := auth.NewJWTService(jwtSecret, jwtExpiration)

	// Initialize admin user
	log.Println("Initializing admin user...")
	if err := initializeAdmin(ctx, userRepo, settingsRepo, config.GetEnv("ADMIN_USERNAME", "admin"), config.GetEnv("ADMIN_PASSWORD", "admin123")); err != nil {
		log.Printf("Warning: Failed to initialize admin user: %v", err)
	}

	// Initialize handlers
	projectsHandler := handlers.NewProjectsHandler(projectRepo, groupRepo)
	groupsHandler := handlers.NewGroupsHandler(groupRepo, projectRepo)
	runsHandler := handlers.NewRunsHandler(logRunRepo, groupRepo, projectRepo, lokiClient, taskQueue)
	authHandler := handlers.NewAuthHandler(userRepo, settingsRepo, tokenService, jwtService, systemConfigRepo)
	statusHandler := handlers.NewStatusHandler(logRunRepo, taskQueue)
	settingsHandler := handlers.NewSettingsHandler(settingsRepo, projectRepo)
	managementHandler := handlers.NewManagementHandler(projectRepo, groupRepo, logRunRepo)
	adminHandler := handlers.NewAdminHandler(userRepo, systemConfigRepo)

	// Create Gin router
	router := gin.Default()

	// Parse CORS origins from environment
	allowedOrigins := strings.Split(corsOrigins, ",")
	for i := range allowedOrigins {
		allowedOrigins[i] = strings.TrimSpace(allowedOrigins[i])
	}
	log.Printf("CORS allowed origins: %v", allowedOrigins)

	// CORS middleware
	router.Use(cors.New(cors.Config{
		AllowOrigins:     allowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length", "X-RateLimit-Limit", "X-RateLimit-Remaining", "Retry-After"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Health check endpoint (no auth required)
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "healthy"})
	})

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Auth endpoints (no auth required, but rate limited)
		auth := v1.Group("/auth")
		auth.Use(middleware.AuthRateLimitMiddleware(redisClient))
		{
			auth.POST("/login", authHandler.Login)
			auth.POST("/register", authHandler.Register)
			auth.GET("/registration-status", authHandler.GetRegistrationStatus)
		}

		// Protected endpoints (JWT auth required for Web UI)
		protected := v1.Group("")
		protected.Use(middleware.JWTAuthMiddleware(jwtService))
		protected.Use(middleware.ActiveUserMiddleware(userRepo))
		{
			// Project and data access (read)
			protected.GET("/projects", projectsHandler.ListProjects)
			protected.GET("/projects/:id", projectsHandler.GetProject)
			protected.GET("/projects/:id/groups", projectsHandler.GetProjectGroups)
			protected.GET("/groups/:id", groupsHandler.GetGroup)
			protected.GET("/groups/:id/runs", runsHandler.ListRuns)
			protected.GET("/runs/:id", runsHandler.GetRun)
			protected.GET("/runs/:id/logs", runsHandler.GetRunLogs)

			// Status endpoints
			protected.GET("/status/statistics", statusHandler.GetStatistics)
			protected.GET("/status/recent", statusHandler.GetRecentRuns)
			protected.POST("/status/archive-completed", statusHandler.ArchiveCompletedRuns)

			// Project management (write)
			protected.POST("/projects", projectsHandler.CreateProject)
			protected.POST("/runs/:id/analyze", runsHandler.TriggerAIAnalysis)

			// User management
			protected.GET("/auth/me", authHandler.GetCurrentUser)

			// Token management
			protected.GET("/auth/tokens", authHandler.ListTokens)
			protected.POST("/auth/tokens", authHandler.CreateToken)
			protected.DELETE("/auth/tokens/:id", authHandler.DeleteToken)

			// User settings management (requires auth)
			protected.GET("/settings", settingsHandler.GetUserSettings)
			protected.PUT("/settings", settingsHandler.UpdateUserSettings)

			// Project settings management (requires auth)
			protected.GET("/projects/:id/settings", settingsHandler.GetProjectSettings)
			protected.PUT("/projects/:id/settings", settingsHandler.UpdateProjectSettings)
			protected.DELETE("/projects/:id/settings", settingsHandler.DeleteProjectSettings)
			protected.GET("/projects/:id/settings/effective", settingsHandler.GetEffectiveSettings)

			// Resource management (requires auth)
			protected.PUT("/projects/:id", managementHandler.UpdateProject)
			protected.DELETE("/projects/:id", managementHandler.DeleteProject)
			protected.PUT("/groups/:id", managementHandler.UpdateGroup)
			protected.DELETE("/groups/:id", managementHandler.DeleteGroup)
			protected.DELETE("/runs/:id", managementHandler.DeleteRun)
		}

		// Admin endpoints (JWT auth + admin role required)
		admin := v1.Group("/admin")
		admin.Use(middleware.JWTAuthMiddleware(jwtService))
		admin.Use(middleware.AdminMiddleware(userRepo))
		{
			// User management
			admin.GET("/users", adminHandler.ListUsers)
			admin.GET("/users/stats", adminHandler.GetUserStats)
			admin.PUT("/users/:id/status", adminHandler.UpdateUserStatus)
			admin.PUT("/users/:id/admin", adminHandler.UpdateUserAdminStatus)
			admin.DELETE("/users/:id", adminHandler.DeleteUser)

			// System configuration
			admin.GET("/config", adminHandler.ListSystemConfig)
			admin.GET("/config/:key", adminHandler.GetSystemConfig)
			admin.PUT("/config/:key", adminHandler.UpdateSystemConfig)
			admin.DELETE("/config/:key", adminHandler.DeleteSystemConfig)

			// System statistics
			admin.GET("/stats", adminHandler.GetSystemStats)
		}
	}

	// Start server
	log.Printf("Starting API server on port %s...", apiPort)
	go func() {
		if err := router.Run(":" + apiPort); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down gracefully...")
	log.Println("Server stopped")
}

// initializeAdmin creates the admin user if no users exist
func initializeAdmin(ctx context.Context, userRepo *repository.UserRepository, settingsRepo *repository.SettingsRepository, username, password string) error {
	// Check if any users exist
	count, err := userRepo.Count(ctx)
	if err != nil {
		return fmt.Errorf("failed to count users: %w", err)
	}

	// If users exist, don't create admin
	if count > 0 {
		log.Println("Users already exist, skipping admin creation")
		return nil
	}

	// Hash password
	passwordHash, err := auth.HashPassword(password)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Create admin user
	admin, err := userRepo.Create(ctx, username, passwordHash, true)
	if err != nil {
		return fmt.Errorf("failed to create admin user: %w", err)
	}

	// Create default settings for admin user
	_, err = settingsRepo.CreateDefaultUserSettings(ctx, admin.ID)
	if err != nil {
		log.Printf("Warning: Failed to create default settings for admin user: %v", err)
	}

	log.Printf("Admin user created: %s (ID: %s)", admin.Username, admin.ID)
	return nil
}

// validateJWTSecret validates JWT secret strength
func validateJWTSecret(secret, environment string) error {
	// List of common weak secrets
	weakSecrets := []string{
		"your-secret-key-change-this-in-production",
		"secret",
		"changeme",
		"password",
		"123456",
		"admin",
		"test",
		"default",
	}

	// In production, enforce strict validation
	if environment == "production" {
		// Check against weak secret list
		for _, weak := range weakSecrets {
			if secret == weak {
				return fmt.Errorf("SECURITY: JWT_SECRET is set to a default/weak value in production. Please set a strong random secret")
			}
		}

		// Enforce minimum length (256 bits = 32 bytes)
		if len(secret) < 32 {
			return fmt.Errorf("SECURITY: JWT_SECRET must be at least 32 characters in production, got %d. Generate with: openssl rand -base64 32", len(secret))
		}

		log.Println("JWT secret validated successfully for production environment")
		return nil
	}

	// In development, warn about weak secrets
	for _, weak := range weakSecrets {
		if secret == weak {
			log.Printf("WARNING: JWT_SECRET is set to a default value (%s)", secret)
			log.Println("WARNING: This is acceptable in development but MUST be changed in production")
			log.Printf("WARNING: Generate a strong secret with: openssl rand -base64 32")
			return nil
		}
	}

	if len(secret) < 32 {
		log.Printf("WARNING: JWT_SECRET is shorter than recommended (got %d chars, recommend 32+)", len(secret))
	}

	return nil
}
