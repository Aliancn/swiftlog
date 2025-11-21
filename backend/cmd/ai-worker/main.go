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
	ws "github.com/aliancn/swiftlog/backend/internal/websocket"
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

	// Initialize repositories
	logRunRepo := repository.NewLogRunRepository(db.DB)
	groupRepo := repository.NewLogGroupRepository(db.DB)
	projectRepo := repository.NewProjectRepository(db.DB)
	settingsRepo := repository.NewSettingsRepository(db.DB)

	// Initialize task queue
	taskQueue := queue.NewQueue(redisClient)

	// Start worker
	log.Println("Starting AI Worker...")
	log.Println("AI settings will be fetched per-user from database")
	worker := NewWorker(logRunRepo, groupRepo, projectRepo, settingsRepo, lokiClient, redisClient, taskQueue)
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
	logRunRepo   *repository.LogRunRepository
	groupRepo    *repository.LogGroupRepository
	projectRepo  *repository.ProjectRepository
	settingsRepo *repository.SettingsRepository
	lokiClient   *loki.Client
	redisClient  *redis.Client
	taskQueue    *queue.Queue
}

// NewWorker creates a new AI worker
func NewWorker(
	logRunRepo *repository.LogRunRepository,
	groupRepo *repository.LogGroupRepository,
	projectRepo *repository.ProjectRepository,
	settingsRepo *repository.SettingsRepository,
	lokiClient *loki.Client,
	redisClient *redis.Client,
	taskQueue *queue.Queue,
) *Worker {
	return &Worker{
		logRunRepo:   logRunRepo,
		groupRepo:    groupRepo,
		projectRepo:  projectRepo,
		settingsRepo: settingsRepo,
		lokiClient:   lokiClient,
		redisClient:  redisClient,
		taskQueue:    taskQueue,
	}
}

// Run starts the worker loop using event-driven architecture with concurrent workers
func (w *Worker) Run(ctx context.Context) {
	// Use a fixed number of concurrent workers
	// This can be made configurable in the future
	const maxConcurrentWorkers = 10

	log.Printf("Starting %d concurrent workers for AI analysis...", maxConcurrentWorkers)

	// Create a channel to distribute tasks to workers
	taskChan := make(chan *queue.AIAnalysisTask, maxConcurrentWorkers*2)

	// Start worker goroutines
	for i := 0; i < maxConcurrentWorkers; i++ {
		go w.worker(ctx, i+1, taskChan)
	}

	// Main loop: consume tasks from Redis and distribute to workers
	log.Println("Worker dispatcher running, waiting for AI analysis tasks from queue...")
	for {
		select {
		case <-ctx.Done():
			close(taskChan)
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

			// Send task to worker channel
			select {
			case taskChan <- task:
				// Task sent successfully
			case <-ctx.Done():
				close(taskChan)
				return
			}
		}
	}
}

// worker processes tasks from the task channel
func (w *Worker) worker(ctx context.Context, workerID int, taskChan <-chan *queue.AIAnalysisTask) {
	log.Printf("Worker #%d started", workerID)
	for {
		select {
		case <-ctx.Done():
			log.Printf("Worker #%d shutting down", workerID)
			return
		case task, ok := <-taskChan:
			if !ok {
				log.Printf("Worker #%d task channel closed, shutting down", workerID)
				return
			}

			log.Printf("Worker #%d processing task for run %s (user %s)", workerID, task.RunID, task.UserID)

			// Process the task with user settings
			if err := w.processRunByID(ctx, task.RunID, task.UserID); err != nil {
				log.Printf("Worker #%d failed to process run %s: %v", workerID, task.RunID, err)
				// Notify failure
				_ = w.taskQueue.NotifyAIResult(ctx, task.RunID, "failed", err.Error())
			} else {
				log.Printf("Worker #%d completed run %s successfully", workerID, task.RunID)
				// Notify success
				_ = w.taskQueue.NotifyAIResult(ctx, task.RunID, "completed", "Analysis completed successfully")
			}
		}
	}
}

// processRunByID fetches a run by ID and processes it
func (w *Worker) processRunByID(ctx context.Context, runID, userID uuid.UUID) error {
	run, err := w.logRunRepo.GetByID(ctx, runID)
	if err != nil {
		// Mark as failed in database
		_ = w.logRunRepo.UpdateAIReport(ctx, runID, fmt.Sprintf("Error: Run not found: %v", err), models.AIStatusFailed)
		return fmt.Errorf("failed to get run: %w", err)
	}

	if err := w.processRun(ctx, run, userID); err != nil {
		// Mark as failed in database
		errorMsg := fmt.Sprintf("Error: %v", err)
		_ = w.logRunRepo.UpdateAIReport(ctx, runID, errorMsg, models.AIStatusFailed)

		// Publish AI status update event
		aiStatus := string(models.AIStatusFailed)
		_ = ws.PublishRunUpdate(ctx, w.redisClient, runID, nil, nil, &aiStatus, &errorMsg)

		return err
	}

	return nil
}

// processRun analyzes a single run using user-specific settings
func (w *Worker) processRun(ctx context.Context, run *models.LogRun, userID uuid.UUID) error {
	log.Printf("Processing run %s (status: %s, exit_code: %v) for user %s", run.ID, run.Status, run.ExitCode, userID)

	// Update status to processing
	if err := w.logRunRepo.UpdateAIStatus(ctx, run.ID, models.AIStatusProcessing); err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}

	// Publish AI status update event
	aiStatus := string(models.AIStatusProcessing)
	_ = ws.PublishRunUpdate(ctx, w.redisClient, run.ID, nil, nil, &aiStatus, nil)

	// Get the group to find the project
	group, err := w.groupRepo.GetByID(ctx, run.GroupID)
	if err != nil {
		return fmt.Errorf("failed to get group: %w", err)
	}

	// Fetch effective settings for this user/project
	effectiveSettings, err := w.settingsRepo.GetEffectiveSettings(ctx, group.ProjectID, userID)
	if err != nil {
		return fmt.Errorf("failed to get effective settings: %w", err)
	}

	// Check if AI is enabled
	if !effectiveSettings.AIEnabled {
		return fmt.Errorf("AI analysis is disabled for this user/project")
	}

	// Check API key
	if effectiveSettings.AIAPIKey == "" {
		return fmt.Errorf("AI API key not configured")
	}

	log.Printf("Using AI settings - Model: %s, BaseURL: %s, MaxTokens: %d, MaxLogLines: %d, Strategy: %s",
		effectiveSettings.AIModel, effectiveSettings.AIBaseURL, effectiveSettings.AIMaxTokens,
		effectiveSettings.AIMaxLogLines, effectiveSettings.AILogTruncateStrategy)

	// Create analyzer with user-specific settings
	analyzer := ai.NewAnalyzer(&ai.Config{
		APIKey:       effectiveSettings.AIAPIKey,
		BaseURL:      effectiveSettings.AIBaseURL,
		Model:        effectiveSettings.AIModel,
		MaxTokens:    effectiveSettings.AIMaxTokens,
		SystemPrompt: effectiveSettings.AISystemPrompt,
	})

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

	// Analyze logs with user-specific settings
	result, err := analyzer.AnalyzeLogs(ctx, logLines, exitCode, string(run.Status),
		effectiveSettings.AIMaxLogLines, string(effectiveSettings.AILogTruncateStrategy))
	if err != nil {
		return fmt.Errorf("AI analysis failed: %w", err)
	}

	log.Printf("Analysis complete for run %s (tokens used: %d)", run.ID, result.TokensUsed)

	// Save report
	if err := w.logRunRepo.UpdateAIReport(ctx, run.ID, result.Report, models.AIStatusCompleted); err != nil {
		return fmt.Errorf("failed to save report: %w", err)
	}

	// Publish AI status update event with report
	aiStatus = string(models.AIStatusCompleted)
	_ = ws.PublishRunUpdate(ctx, w.redisClient, run.ID, nil, nil, &aiStatus, &result.Report)

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
