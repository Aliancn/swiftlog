package websocket

import (
	"context"
	"encoding/json"
	"log"
	"sync"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// Hub maintains the set of active clients and broadcasts messages
type Hub struct {
	// Registered clients per run ID
	clients map[uuid.UUID]map[*Client]bool

	// Register requests from clients
	register chan *Client

	// Unregister requests from clients
	unregister chan *Client

	// Broadcast messages to clients
	broadcast chan *BroadcastMessage

	// Redis client for pub/sub
	redisClient *redis.Client

	// Mutex for thread-safe operations
	mu sync.RWMutex

	// Context for cancellation
	ctx context.Context
}

// BroadcastMessage represents a message to broadcast
type BroadcastMessage struct {
	RunID   uuid.UUID
	Message []byte
}

// LogMessage represents a log line message
type LogMessage struct {
	Type      string `json:"type"`
	RunID     string `json:"run_id"`
	Timestamp string `json:"timestamp"`
	Level     string `json:"level"`
	Content   string `json:"content"`
}

// RunUpdateMessage represents a run status update message
type RunUpdateMessage struct {
	Type     string  `json:"type"`
	RunID    string  `json:"run_id"`
	Status   *string `json:"status,omitempty"`
	ExitCode *int32  `json:"exit_code,omitempty"`
	AIStatus *string `json:"ai_status,omitempty"`
	AIReport *string `json:"ai_report,omitempty"`
}

// NewHub creates a new WebSocket hub
func NewHub(ctx context.Context, redisClient *redis.Client) *Hub {
	return &Hub{
		clients:     make(map[uuid.UUID]map[*Client]bool),
		register:    make(chan *Client),
		unregister:  make(chan *Client),
		broadcast:   make(chan *BroadcastMessage, 256),
		redisClient: redisClient,
		ctx:         ctx,
	}
}

// Run starts the hub
func (h *Hub) Run() {
	// Subscribe to Redis pub/sub for log events
	go h.subscribeToRedis()

	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			if h.clients[client.runID] == nil {
				h.clients[client.runID] = make(map[*Client]bool)
			}
			h.clients[client.runID][client] = true
			h.mu.Unlock()
			log.Printf("Client registered for run %s", client.runID)

		case client := <-h.unregister:
			h.mu.Lock()
			if clients, ok := h.clients[client.runID]; ok {
				if _, ok := clients[client]; ok {
					delete(clients, client)
					close(client.send)
					if len(clients) == 0 {
						delete(h.clients, client.runID)
					}
				}
			}
			h.mu.Unlock()
			log.Printf("Client unregistered for run %s", client.runID)

		case message := <-h.broadcast:
			h.mu.RLock()
			clients := h.clients[message.RunID]
			h.mu.RUnlock()

			for client := range clients {
				select {
				case client.send <- message.Message:
				default:
					// Client send buffer is full, close it
					h.mu.Lock()
					close(client.send)
					delete(h.clients[message.RunID], client)
					h.mu.Unlock()
				}
			}

		case <-h.ctx.Done():
			return
		}
	}
}

// Broadcast sends a message to all clients subscribed to a run
func (h *Hub) Broadcast(runID uuid.UUID, message []byte) {
	h.broadcast <- &BroadcastMessage{
		RunID:   runID,
		Message: message,
	}
}

// subscribeToRedis listens to Redis pub/sub for log events
func (h *Hub) subscribeToRedis() {
	pubsub := h.redisClient.Subscribe(h.ctx, "swiftlog:logs")
	defer pubsub.Close()

	ch := pubsub.Channel()

	for {
		select {
		case msg := <-ch:
			// Parse the message
			var logMsg LogMessage
			if err := json.Unmarshal([]byte(msg.Payload), &logMsg); err != nil {
				log.Printf("Failed to unmarshal log message: %v", err)
				continue
			}

			// Parse run ID
			runID, err := uuid.Parse(logMsg.RunID)
			if err != nil {
				log.Printf("Invalid run ID: %v", err)
				continue
			}

			// Broadcast to connected clients
			h.Broadcast(runID, []byte(msg.Payload))

		case <-h.ctx.Done():
			return
		}
	}
}

// PublishLog publishes a log message to Redis (called by Ingestor)
func PublishLog(ctx context.Context, redisClient *redis.Client, runID uuid.UUID, timestamp, level, content string) error {
	logMsg := LogMessage{
		Type:      "log",
		RunID:     runID.String(),
		Timestamp: timestamp,
		Level:     level,
		Content:   content,
	}

	data, err := json.Marshal(logMsg)
	if err != nil {
		return err
	}

	return redisClient.Publish(ctx, "swiftlog:logs", data).Err()
}

// PublishRunUpdate publishes a run status update to Redis
func PublishRunUpdate(ctx context.Context, redisClient *redis.Client, runID uuid.UUID, status *string, exitCode *int32, aiStatus *string, aiReport *string) error {
	updateMsg := RunUpdateMessage{
		Type:     "run_update",
		RunID:    runID.String(),
		Status:   status,
		ExitCode: exitCode,
		AIStatus: aiStatus,
		AIReport: aiReport,
	}

	data, err := json.Marshal(updateMsg)
	if err != nil {
		return err
	}

	return redisClient.Publish(ctx, "swiftlog:logs", data).Err()
}
