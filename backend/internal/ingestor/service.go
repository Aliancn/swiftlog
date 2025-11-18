package ingestor

import (
	"fmt"
	"io"
	"time"

	"github.com/aliancn/swiftlog/backend/internal/auth"
	"github.com/aliancn/swiftlog/backend/internal/loki"
	"github.com/aliancn/swiftlog/backend/internal/models"
	"github.com/aliancn/swiftlog/backend/internal/repository"
	pb "github.com/aliancn/swiftlog/backend/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Service implements the LogStreamer gRPC service
type Service struct {
	pb.UnimplementedLogStreamerServer
	logRunRepo    *repository.LogRunRepository
	projectRepo   *repository.ProjectRepository
	groupRepo     *repository.LogGroupRepository
	lokiClient    *loki.Client
	batchSize     int
	batchInterval time.Duration
}

// Config holds ingestor service configuration
type Config struct {
	LogRunRepo    *repository.LogRunRepository
	ProjectRepo   *repository.ProjectRepository
	GroupRepo     *repository.LogGroupRepository
	LokiClient    *loki.Client
	BatchSize     int           // Number of log lines to batch before sending to Loki
	BatchInterval time.Duration // Maximum time to wait before sending a batch
}

// NewService creates a new ingestor service
func NewService(cfg *Config) *Service {
	if cfg.BatchSize == 0 {
		cfg.BatchSize = 100 // Default from research.md
	}
	if cfg.BatchInterval == 0 {
		cfg.BatchInterval = 1 * time.Second
	}

	return &Service{
		logRunRepo:    cfg.LogRunRepo,
		projectRepo:   cfg.ProjectRepo,
		groupRepo:     cfg.GroupRepo,
		lokiClient:    cfg.LokiClient,
		batchSize:     cfg.BatchSize,
		batchInterval: cfg.BatchInterval,
	}
}

// StreamLog implements the bidirectional streaming RPC
func (s *Service) StreamLog(stream pb.LogStreamer_StreamLogServer) error {
	ctx := stream.Context()

	// Get authenticated user ID from context
	userID, err := auth.GetUserIDFromContext(ctx)
	if err != nil {
		return status.Errorf(codes.Unauthenticated, "authentication required: %v", err)
	}

	// Receive the first message (must be metadata)
	req, err := stream.Recv()
	if err != nil {
		return status.Errorf(codes.Internal, "failed to receive first message: %v", err)
	}

	metadata := req.GetMetadata()
	if metadata == nil {
		return status.Errorf(codes.InvalidArgument, "first message must contain metadata")
	}

	// Get or create project and group
	projectName := metadata.ProjectName
	if projectName == "" {
		projectName = "default"
	}
	groupName := metadata.GroupName
	if groupName == "" {
		groupName = "default"
	}

	project, err := s.projectRepo.GetOrCreate(ctx, userID, projectName)
	if err != nil {
		return status.Errorf(codes.Internal, "failed to get/create project: %v", err)
	}

	group, err := s.groupRepo.GetOrCreate(ctx, project.ID, groupName)
	if err != nil {
		return status.Errorf(codes.Internal, "failed to get/create group: %v", err)
	}

	// Create log run
	logRun, err := s.logRunRepo.Create(ctx, group.ID)
	if err != nil {
		return status.Errorf(codes.Internal, "failed to create log run: %v", err)
	}

	// Send StreamStarted response
	err = stream.Send(&pb.StreamLogResponse{
		Event: &pb.StreamLogResponse_Started{
			Started: &pb.StreamStarted{
				RunId: logRun.ID.String(),
			},
		},
	})
	if err != nil {
		return status.Errorf(codes.Internal, "failed to send started response: %v", err)
	}

	// Start receiving log lines
	logBatch := make([]loki.LogEntry, 0, s.batchSize)
	batchTicker := time.NewTicker(s.batchInterval)
	defer batchTicker.Stop()

	flushBatch := func() error {
		if len(logBatch) == 0 {
			return nil
		}
		if err := s.lokiClient.PushLogs(ctx, logRun.ID, userID, projectName, logBatch); err != nil {
			return fmt.Errorf("failed to push logs to Loki: %w", err)
		}
		logBatch = logBatch[:0] // Clear batch
		return nil
	}

	// Process incoming log lines
	for {
		select {
		case <-ctx.Done():
			// Context cancelled, flush remaining logs
			_ = flushBatch()
			return status.Errorf(codes.Canceled, "stream cancelled")
		case <-batchTicker.C:
			// Flush batch on timer
			if err := flushBatch(); err != nil {
				return status.Errorf(codes.Internal, "failed to flush batch: %v", err)
			}
		default:
			// Receive next message
			req, err := stream.Recv()
			if err == io.EOF {
				// Client closed stream, flush remaining logs
				_ = flushBatch()
				return nil
			}
			if err != nil {
				// Stream error, mark run as aborted
				_ = flushBatch()
				_ = s.logRunRepo.UpdateStatus(ctx, logRun.ID, models.RunStatusAborted, nil)
				return status.Errorf(codes.Internal, "stream error: %v", err)
			}

			// Handle different message types
			if line := req.GetLine(); line != nil {
				// Add log line to batch
				entry := loki.LogEntry{
					Timestamp: line.Timestamp.AsTime(),
					Line:      fmt.Sprintf("[%s] %s", line.Level.String(), line.Content),
				}
				logBatch = append(logBatch, entry)

				// Flush if batch is full
				if len(logBatch) >= s.batchSize {
					if err := flushBatch(); err != nil {
						return status.Errorf(codes.Internal, "failed to flush batch: %v", err)
					}
				}
			} else if completion := req.GetCompletion(); completion != nil {
				// Script completed, flush remaining logs
				if err := flushBatch(); err != nil {
					return status.Errorf(codes.Internal, "failed to flush final batch: %v", err)
				}

				// Update run status based on exit code
				exitCode := completion.ExitCode
				var runStatus models.RunStatus
				if exitCode == 0 {
					runStatus = models.RunStatusCompleted
				} else {
					runStatus = models.RunStatusFailed
				}

				if err := s.logRunRepo.UpdateStatus(ctx, logRun.ID, runStatus, &exitCode); err != nil {
					return status.Errorf(codes.Internal, "failed to update run status: %v", err)
				}

				return nil
			}
		}
	}
}
