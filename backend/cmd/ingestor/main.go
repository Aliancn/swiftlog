package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aliancn/swiftlog/backend/internal/auth"
	"github.com/aliancn/swiftlog/backend/internal/config"
	"github.com/aliancn/swiftlog/backend/internal/crypto"
	"github.com/aliancn/swiftlog/backend/internal/ingestor"
	"github.com/aliancn/swiftlog/backend/internal/loki"
	"github.com/aliancn/swiftlog/backend/internal/queue"
	"github.com/aliancn/swiftlog/backend/internal/repository"
	pb "github.com/aliancn/swiftlog/backend/proto"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
)

func main() {
	ctx := context.Background()

	// Load configuration from environment
	dbURL := config.GetEnv("DATABASE_URL", "postgres://swiftlog:changeme@localhost:5432/swiftlog?sslmode=disable")
	lokiURL := config.GetEnv("LOKI_URL", "http://localhost:3100")
	redisURL := config.GetEnv("REDIS_URL", "redis://localhost:6379")
	grpcPort := config.GetEnv("GRPC_PORT", "50051")

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
	encryptionKey, err := config.GetEncryptionKey("ingestor")
	if err != nil {
		log.Fatalf("Failed to get encryption key: %v", err)
	}
	encryptionService, err := crypto.NewEncryptionService(encryptionKey)
	if err != nil {
		log.Fatalf("Failed to initialize encryption service: %v", err)
	}

	// Initialize repositories
	logRunRepo := repository.NewLogRunRepository(db.DB)
	projectRepo := repository.NewProjectRepository(db.DB)
	groupRepo := repository.NewLogGroupRepository(db.DB)
	settingsRepo := repository.NewSettingsRepository(db.DB, encryptionService)

	// Initialize auth token service
	tokenService := auth.NewTokenService(db.DB)

	// Initialize ingestor service
	ingestorService := ingestor.NewService(&ingestor.Config{
		LogRunRepo:    logRunRepo,
		ProjectRepo:   projectRepo,
		GroupRepo:     groupRepo,
		SettingsRepo:  settingsRepo,
		LokiClient:    lokiClient,
		RedisClient:   redisClient,
		TaskQueue:     taskQueue,
		BatchSize:     100,
		BatchInterval: 1 * time.Second,
	})

	// Create gRPC server with auth interceptors
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(auth.GRPCAuthInterceptor(tokenService)),
		grpc.StreamInterceptor(auth.GRPCAuthStreamInterceptor(tokenService)),
	)

	// Register service
	pb.RegisterLogStreamerServer(grpcServer, ingestorService)

	// Start gRPC server
	lis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", grpcPort, err)
	}

	log.Printf("Starting gRPC Ingestor service on port %s...", grpcPort)
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down gracefully...")
	grpcServer.GracefulStop()
	log.Println("Server stopped")
}
