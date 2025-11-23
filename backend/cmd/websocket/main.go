package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/aliancn/swiftlog/backend/internal/auth"
	"github.com/aliancn/swiftlog/backend/internal/config"
	"github.com/aliancn/swiftlog/backend/internal/repository"
	ws "github.com/aliancn/swiftlog/backend/internal/websocket"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var (
	upgrader         websocket.Upgrader
	allowedOrigins   []string
	isProduction     bool
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Load configuration from environment
	dbURL := config.GetEnv("DATABASE_URL", "postgres://swiftlog:changeme@localhost:5432/swiftlog?sslmode=disable")
	redisURL := config.GetEnv("REDIS_URL", "redis://localhost:6379")
	wsPort := config.GetEnv("WS_PORT", "8081")
	environment := config.GetEnv("ENVIRONMENT", "development")
	allowedOriginsStr := config.GetEnv("ALLOWED_ORIGINS", "http://localhost:3000")

	// Parse allowed origins
	allowedOrigins = strings.Split(allowedOriginsStr, ",")
	for i := range allowedOrigins {
		allowedOrigins[i] = strings.TrimSpace(allowedOrigins[i])
	}
	isProduction = environment == "production"

	// Initialize WebSocket upgrader with secure origin checking
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			origin := r.Header.Get("Origin")

			// In development, allow all origins
			if !isProduction {
				log.Printf("DEV: Accepting WebSocket connection from origin: %s", origin)
				return true
			}

			// In production, check against allowed origins
			for _, allowed := range allowedOrigins {
				if origin == allowed {
					log.Printf("PROD: Accepting WebSocket connection from allowed origin: %s", origin)
					return true
				}
			}

			log.Printf("PROD: Rejecting WebSocket connection from unauthorized origin: %s", origin)
			return false
		},
	}

	// Set Gin mode
	if isProduction {
		gin.SetMode(gin.ReleaseMode)
		log.Printf("Running in PRODUCTION mode with allowed origins: %v", allowedOrigins)
	} else {
		log.Printf("Running in DEVELOPMENT mode - accepting all WebSocket origins")
	}

	// Initialize database connection
	log.Println("Connecting to database...")
	db, err := config.InitDatabase(ctx, dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize Redis client
	log.Println("Connecting to Redis...")
	redisClient, err := config.InitRedis(ctx, redisURL)
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
