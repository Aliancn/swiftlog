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

	// Initialize auth token service
	tokenService := auth.NewTokenService(db.DB)

	// Initialize admin user
	log.Println("Initializing admin user...")
	if err := initializeAdmin(ctx, userRepo, settingsRepo, getEnv("ADMIN_USERNAME", "admin"), getEnv("ADMIN_PASSWORD", "admin123")); err != nil {
		log.Printf("Warning: Failed to initialize admin user: %v", err)
	}

	// Initialize handlers
	projectsHandler := handlers.NewProjectsHandler(projectRepo, groupRepo)
	groupsHandler := handlers.NewGroupsHandler(groupRepo, projectRepo)
	runsHandler := handlers.NewRunsHandler(logRunRepo, groupRepo, projectRepo, lokiClient, taskQueue)
	authHandler := handlers.NewAuthHandler(userRepo, settingsRepo, tokenService)
	statusHandler := handlers.NewStatusHandler(logRunRepo, taskQueue)
	settingsHandler := handlers.NewSettingsHandler(settingsRepo, projectRepo)

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
		}

		// Public read-only endpoints (no auth required for development)
		v1.GET("/projects", projectsHandler.ListProjects)
		v1.GET("/projects/:id", projectsHandler.GetProject)
		v1.GET("/projects/:id/groups", projectsHandler.GetProjectGroups)
		v1.GET("/groups/:id", groupsHandler.GetGroup)
		v1.GET("/groups/:id/runs", runsHandler.ListRuns)
		v1.GET("/runs/:id", runsHandler.GetRun)
		v1.GET("/runs/:id/logs", runsHandler.GetRunLogs)

		// Status endpoints (no auth required for development)
		v1.GET("/status/statistics", statusHandler.GetStatistics)
		v1.GET("/status/recent", statusHandler.GetRecentRuns)

		// Protected endpoints (auth required)
		protected := v1.Group("")
		protected.Use(middleware.AuthMiddleware(tokenService))
		{
			// Project management
			protected.POST("/projects", projectsHandler.CreateProject)
			protected.POST("/runs/:id/analyze", runsHandler.TriggerAIAnalysis)

			// User management
			protected.GET("/auth/me", authHandler.GetCurrentUser)
			protected.GET("/auth/users", authHandler.ListUsers)

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
