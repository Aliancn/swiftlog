package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aliancn/swiftlog/backend/internal/api/handlers"
	"github.com/aliancn/swiftlog/backend/internal/api/middleware"
	"github.com/aliancn/swiftlog/backend/internal/auth"
	"github.com/aliancn/swiftlog/backend/internal/database"
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
	dbURL := getEnv("DATABASE_URL", "postgres://swiftlog:changeme@localhost:5432/swiftlog?sslmode=disable")
	lokiURL := getEnv("LOKI_URL", "http://localhost:3100")
	redisURL := getEnv("REDIS_URL", "redis://localhost:6379")
	apiPort := getEnv("API_PORT", "8080")
	environment := getEnv("ENVIRONMENT", "development")
	jwtSecret := getEnv("JWT_SECRET", "your-secret-key-change-this-in-production")
	jwtExpiration := getEnvDuration("JWT_EXPIRATION", 24*time.Hour)

	// Set Gin mode
	if environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize database connection
	log.Println("Connecting to database...")
	db, err := initDatabase(ctx, dbURL)
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

	// Initialize repositories
	projectRepo := repository.NewProjectRepository(db.DB)
	groupRepo := repository.NewLogGroupRepository(db.DB)
	logRunRepo := repository.NewLogRunRepository(db.DB)
	userRepo := repository.NewUserRepository(db.DB)
	settingsRepo := repository.NewSettingsRepository(db.DB)
	systemConfigRepo := repository.NewSystemConfigRepository(db.DB)

	// Initialize auth services
	tokenService := auth.NewTokenService(db.DB)
	jwtService := auth.NewJWTService(jwtSecret, jwtExpiration)

	// Initialize admin user
	log.Println("Initializing admin user...")
	if err := initializeAdmin(ctx, userRepo, settingsRepo, getEnv("ADMIN_USERNAME", "admin"), getEnv("ADMIN_PASSWORD", "admin123")); err != nil {
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

	// CORS middleware
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
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
		// Auth endpoints (no auth required)
		auth := v1.Group("/auth")
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

func initDatabase(ctx context.Context, dbURL string) (*database.DB, error) {
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(2 * time.Minute)

	// Verify connection
	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &database.DB{DB: db}, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		duration, err := time.ParseDuration(value)
		if err != nil {
			log.Printf("Warning: Invalid duration for %s, using default: %v", key, defaultValue)
			return defaultValue
		}
		return duration
	}
	return defaultValue
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
