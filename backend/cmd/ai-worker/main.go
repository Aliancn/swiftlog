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

	"github.com/aliancn/swiftlog/backend/internal/ai"
	"github.com/aliancn/swiftlog/backend/internal/database"
	"github.com/aliancn/swiftlog/backend/internal/loki"
	"github.com/aliancn/swiftlog/backend/internal/models"
	"github.com/aliancn/swiftlog/backend/internal/queue"
	"github.com/aliancn/swiftlog/backend/internal/repository"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Load configuration from environment
	dbURL := getEnv("DATABASE_URL", "postgres://swiftlog:changeme@localhost:5432/swiftlog?sslmode=disable")
	lokiURL := getEnv("LOKI_URL", "http://localhost:3100")
	redisURL := getEnv("REDIS_URL", "redis://localhost:6379")
	openAIKey := getEnv("OPENAI_API_KEY", "")
	openAIBaseURL := getEnv("OPENAI_BASE_URL", "https://api.openai.com/v1")
	openAIModel := getEnv("OPENAI_MODEL", "gpt-4o-mini")
	maxTokens := 500

	if openAIKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable is required")
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
	redisClient, err := initRedis(ctx, redisURL)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redisClient.Close()

	// Initialize AI analyzer
	log.Println("Initializing AI analyzer...")
	log.Printf("Using OpenAI endpoint: %s", openAIBaseURL)
	analyzer := ai.NewAnalyzer(&ai.Config{
		APIKey:    openAIKey,
		BaseURL:   openAIBaseURL,
		Model:     openAIModel,
		MaxTokens: maxTokens,
	})

	// Initialize repository
	logRunRepo := repository.NewLogRunRepository(db.DB)

	// Initialize task queue
	taskQueue := queue.NewQueue(redisClient)

	// Start worker
	log.Println("Starting AI Worker...")
	worker := NewWorker(logRunRepo, lokiClient, analyzer, redisClient, taskQueue)
	go worker.Run(ctx)

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down gracefully...")
	cancel()
	time.Sleep(2 * time.Second)
	log.Println("Worker stopped")
}

// Worker processes AI analysis jobs
type Worker struct {
	logRunRepo  *repository.LogRunRepository
	lokiClient  *loki.Client
	analyzer    *ai.Analyzer
	redisClient *redis.Client
	taskQueue   *queue.Queue
}

// NewWorker creates a new AI worker
func NewWorker(
	logRunRepo *repository.LogRunRepository,
	lokiClient *loki.Client,
	analyzer *ai.Analyzer,
	redisClient *redis.Client,
	taskQueue *queue.Queue,
) *Worker {
	return &Worker{
		logRunRepo:  logRunRepo,
		lokiClient:  lokiClient,
		analyzer:    analyzer,
		redisClient: redisClient,
		taskQueue:   taskQueue,
	}
}

// Run starts the worker loop using event-driven architecture
func (w *Worker) Run(ctx context.Context) {
	log.Println("Worker running, waiting for AI analysis tasks from queue...")

	for {
		select {
		case <-ctx.Done():
			return
		default:
			// Block and wait for task from Redis queue (5 second timeout)
			task, err := w.taskQueue.ConsumeAITask(ctx, 5*time.Second)
			if err != nil {
				log.Printf("Error consuming task: %v", err)
				continue
			}

			// No task available (timeout), continue waiting
			if task == nil {
				continue
			}

			log.Printf("Received task for run %s", task.RunID)

			// Process the task
			if err := w.processRunByID(ctx, task.RunID); err != nil {
				log.Printf("Failed to process run %s: %v", task.RunID, err)
				// Notify failure
				_ = w.taskQueue.NotifyAIResult(ctx, task.RunID, "failed", err.Error())
			} else {
				// Notify success
				_ = w.taskQueue.NotifyAIResult(ctx, task.RunID, "completed", "Analysis completed successfully")
			}
		}
	}
}

// processRunByID fetches a run by ID and processes it
func (w *Worker) processRunByID(ctx context.Context, runID uuid.UUID) error {
	run, err := w.logRunRepo.GetByID(ctx, runID)
	if err != nil {
		// Mark as failed in database
		_ = w.logRunRepo.UpdateAIReport(ctx, runID, fmt.Sprintf("Error: Run not found: %v", err), models.AIStatusFailed)
		return fmt.Errorf("failed to get run: %w", err)
	}

	if err := w.processRun(ctx, run); err != nil {
		// Mark as failed in database
		_ = w.logRunRepo.UpdateAIReport(ctx, runID, fmt.Sprintf("Error: %v", err), models.AIStatusFailed)
		return err
	}

	return nil
}

// processRun analyzes a single run
func (w *Worker) processRun(ctx context.Context, run *models.LogRun) error {
	log.Printf("Processing run %s (status: %s, exit_code: %v)", run.ID, run.Status, run.ExitCode)

	// Update status to processing
	if err := w.logRunRepo.UpdateAIStatus(ctx, run.ID, models.AIStatusProcessing); err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}

	// Fetch logs from Loki
	logs, err := w.lokiClient.QueryLogs(ctx, run.ID)
	if err != nil {
		return fmt.Errorf("failed to fetch logs: %w", err)
	}

	if len(logs) == 0 {
		return fmt.Errorf("no logs found for run")
	}

	// Convert logs to string array
	logLines := make([]string, len(logs))
	for i, log := range logs {
		logLines[i] = log.Line
	}

	// Get exit code
	exitCode := int32(0)
	if run.ExitCode.Valid {
		exitCode = run.ExitCode.Int32
	}

	// Analyze logs
	result, err := w.analyzer.AnalyzeLogs(ctx, logLines, exitCode, string(run.Status))
	if err != nil {
		return fmt.Errorf("AI analysis failed: %w", err)
	}

	log.Printf("Analysis complete for run %s (tokens used: %d)", run.ID, result.TokensUsed)

	// Save report
	if err := w.logRunRepo.UpdateAIReport(ctx, run.ID, result.Report, models.AIStatusCompleted); err != nil {
		return fmt.Errorf("failed to save report: %w", err)
	}

	return nil
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
