package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aliancn/swiftlog/backend/internal/auth"
	"github.com/aliancn/swiftlog/backend/internal/database"
	"github.com/aliancn/swiftlog/backend/internal/repository"
	ws "github.com/aliancn/swiftlog/backend/internal/websocket"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Allow all origins in development (should be restricted in production)
		return true
	},
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Load configuration from environment
	dbURL := getEnv("DATABASE_URL", "postgres://swiftlog:changeme@localhost:5432/swiftlog?sslmode=disable")
	redisURL := getEnv("REDIS_URL", "redis://localhost:6379")
	wsPort := getEnv("WS_PORT", "8081")
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

	// Initialize Redis client
	log.Println("Connecting to Redis...")
	redisClient, err := initRedis(ctx, redisURL)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redisClient.Close()

	// Initialize repositories
	logRunRepo := repository.NewLogRunRepository(db.DB)
	groupRepo := repository.NewLogGroupRepository(db.DB)
	projectRepo := repository.NewProjectRepository(db.DB)

	// Initialize auth token service
	tokenService := auth.NewTokenService(db.DB)

	// Create WebSocket hub
	hub := ws.NewHub(ctx, redisClient)
	go hub.Run()

	// Create Gin router
	router := gin.Default()

	// CORS middleware
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "healthy"})
	})

	// WebSocket endpoint
	router.GET("/ws/runs/:run_id", func(c *gin.Context) {
		handleWebSocket(c, hub, tokenService, logRunRepo, groupRepo, projectRepo)
	})

	// Start server
	log.Printf("Starting WebSocket server on port %s...", wsPort)
	go func() {
		if err := router.Run(":" + wsPort); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down gracefully...")
	cancel()
	time.Sleep(1 * time.Second)
	log.Println("Server stopped")
}

func handleWebSocket(
	c *gin.Context,
	hub *ws.Hub,
	tokenService *auth.TokenService,
	logRunRepo *repository.LogRunRepository,
	groupRepo *repository.LogGroupRepository,
	projectRepo *repository.ProjectRepository,
) {
	// Extract token from query parameter
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing token"})
		return
	}

	// Validate token
	userID, err := tokenService.ValidateToken(c.Request.Context(), token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	// Parse run ID
	runID, err := uuid.Parse(c.Param("run_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid run ID"})
		return
	}

	// Verify user has access to this run
	run, err := logRunRepo.GetByID(c.Request.Context(), runID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Run not found"})
		return
	}

	group, err := groupRepo.GetByID(c.Request.Context(), run.GroupID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify ownership"})
		return
	}

	project, err := projectRepo.GetByID(c.Request.Context(), group.ProjectID)
	if err != nil || project.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	// Upgrade to WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}

	// Create client and register with hub
	client := ws.NewClient(hub, conn, runID)
	client.Register()
	client.Start()
}

func initDatabase(ctx context.Context, dbURL string) (*database.DB, error) {
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(2 * time.Minute)

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &database.DB{DB: db}, nil
}

func initRedis(ctx context.Context, redisURL string) (*redis.Client, error) {
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Redis URL: %w", err)
	}

	client := redis.NewClient(opt)

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping Redis: %w", err)
	}

	return client, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
