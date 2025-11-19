package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	// AIAnalysisQueue is the Redis key for AI analysis task queue
	AIAnalysisQueue = "swiftlog:ai:queue"
	// AIAnalysisNotify is the Redis pub/sub channel for AI analysis notifications
	AIAnalysisNotify = "swiftlog:ai:notify"
)

// AIAnalysisTask represents a task in the AI analysis queue
type AIAnalysisTask struct {
	RunID     uuid.UUID `json:"run_id"`
	UserID    uuid.UUID `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
}

// AIAnalysisResult represents the result notification for AI analysis
type AIAnalysisResult struct {
	RunID   uuid.UUID `json:"run_id"`
	Status  string    `json:"status"` // "completed" or "failed"
	Message string    `json:"message,omitempty"`
}

// Queue provides Redis-based task queue operations
type Queue struct {
	client *redis.Client
}

// NewQueue creates a new Queue instance
func NewQueue(client *redis.Client) *Queue {
	return &Queue{client: client}
}

// PublishAITask adds a new AI analysis task to the queue
func (q *Queue) PublishAITask(ctx context.Context, runID, userID uuid.UUID) error {
	task := AIAnalysisTask{
		RunID:     runID,
		UserID:    userID,
		CreatedAt: time.Now().UTC(),
	}

	data, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("failed to marshal task: %w", err)
	}

	// Use LPUSH to add task to the left of the list
	if err := q.client.LPush(ctx, AIAnalysisQueue, data).Err(); err != nil {
		return fmt.Errorf("failed to publish task: %w", err)
	}

	return nil
}

// ConsumeAITask blocks and waits for the next AI analysis task
// Returns nil task when context is cancelled
func (q *Queue) ConsumeAITask(ctx context.Context, timeout time.Duration) (*AIAnalysisTask, error) {
	// Use BRPOP to block and wait for task from the right of the list
	result, err := q.client.BRPop(ctx, timeout, AIAnalysisQueue).Result()
	if err != nil {
		if err == redis.Nil {
			// Timeout, no task available
			return nil, nil
		}
		if ctx.Err() != nil {
			// Context cancelled
			return nil, nil
		}
		return nil, fmt.Errorf("failed to consume task: %w", err)
	}

	// result[0] is the key, result[1] is the value
	if len(result) < 2 {
		return nil, fmt.Errorf("invalid result from BRPOP")
	}

	var task AIAnalysisTask
	if err := json.Unmarshal([]byte(result[1]), &task); err != nil {
		return nil, fmt.Errorf("failed to unmarshal task: %w", err)
	}

	return &task, nil
}

// NotifyAIResult publishes an AI analysis result notification
func (q *Queue) NotifyAIResult(ctx context.Context, runID uuid.UUID, status, message string) error {
	result := AIAnalysisResult{
		RunID:   runID,
		Status:  status,
		Message: message,
	}

	data, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("failed to marshal result: %w", err)
	}

	if err := q.client.Publish(ctx, AIAnalysisNotify, data).Err(); err != nil {
		return fmt.Errorf("failed to publish notification: %w", err)
	}

	return nil
}

// SubscribeAIResults subscribes to AI analysis result notifications
func (q *Queue) SubscribeAIResults(ctx context.Context) <-chan AIAnalysisResult {
	ch := make(chan AIAnalysisResult, 100)

	go func() {
		defer close(ch)

		pubsub := q.client.Subscribe(ctx, AIAnalysisNotify)
		defer pubsub.Close()

		msgCh := pubsub.Channel()

		for {
			select {
			case msg := <-msgCh:
				if msg == nil {
					return
				}
				var result AIAnalysisResult
				if err := json.Unmarshal([]byte(msg.Payload), &result); err != nil {
					continue
				}
				select {
				case ch <- result:
				default:
					// Channel full, skip
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	return ch
}

// GetQueueLength returns the current length of the AI analysis queue
func (q *Queue) GetQueueLength(ctx context.Context) (int64, error) {
	return q.client.LLen(ctx, AIAnalysisQueue).Result()
}
