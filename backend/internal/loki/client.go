package loki

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// Client is a Loki HTTP client
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// Config holds Loki client configuration
type Config struct {
	URL     string
	Timeout time.Duration
}

// NewClient creates a new Loki client
func NewClient(cfg *Config) *Client {
	if cfg.Timeout == 0 {
		cfg.Timeout = 10 * time.Second
	}

	return &Client{
		baseURL: cfg.URL,
		httpClient: &http.Client{
			Timeout: cfg.Timeout,
		},
	}
}

// Stream represents a Loki stream
type Stream struct {
	Stream map[string]string `json:"stream"`
	Values [][]string        `json:"values"`
}

// PushRequest represents a Loki push request
type PushRequest struct {
	Streams []Stream `json:"streams"`
}

// LogEntry represents a single log entry
type LogEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Line      string    `json:"-"`
}

// MarshalJSON implements custom JSON serialization for LogEntry
func (e LogEntry) MarshalJSON() ([]byte, error) {
	// Extract level from line (e.g., "[STDOUT]" or "[STDERR]")
	level := "STDOUT"
	content := e.Line

	if len(e.Line) > 8 && e.Line[0] == '[' {
		if len(e.Line) > 9 && e.Line[1:9] == "STDOUT] " {
			level = "STDOUT"
			content = e.Line[9:]
		} else if len(e.Line) > 9 && e.Line[1:9] == "STDERR] " {
			level = "STDERR"
			content = e.Line[9:]
		}
	}

	return json.Marshal(&struct {
		Timestamp time.Time `json:"timestamp"`
		Level     string    `json:"level"`
		Content   string    `json:"content"`
	}{
		Timestamp: e.Timestamp,
		Level:     level,
		Content:   content,
	})
}

// PushLogs pushes log lines to Loki
func (c *Client) PushLogs(ctx context.Context, runID uuid.UUID, userID uuid.UUID, projectName string, entries []LogEntry) error {
	if len(entries) == 0 {
		return nil
	}

	// Build labels (following 4-label Loki strategy from research.md)
	labels := map[string]string{
		"job":     "swiftlog",
		"user_id": userID.String(),
		"run_id":  runID.String(),
		"project": projectName,
	}

	// Convert log entries to Loki format
	values := make([][]string, len(entries))
	for i, entry := range entries {
		// Loki expects [timestamp_ns, log_line]
		timestampNs := fmt.Sprintf("%d", entry.Timestamp.UnixNano())
		values[i] = []string{timestampNs, entry.Line}
	}

	// Create push request
	pushReq := PushRequest{
		Streams: []Stream{
			{
				Stream: labels,
				Values: values,
			},
		},
	}

	// Marshal to JSON
	payload, err := json.Marshal(pushReq)
	if err != nil {
		return fmt.Errorf("failed to marshal push request: %w", err)
	}

	// Send HTTP POST to Loki
	url := fmt.Sprintf("%s/loki/api/v1/push", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to push logs to Loki: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Loki returned error: status=%d, body=%s", resp.StatusCode, string(body))
	}

	return nil
}

// QueryRequest represents a Loki query request
type QueryRequest struct {
	Query     string
	Limit     int
	Start     time.Time
	End       time.Time
	Direction string // "backward" or "forward"
}

// QueryResponse represents a Loki query response
type QueryResponse struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string `json:"resultType"`
		Result     []struct {
			Stream map[string]string `json:"stream"`
			Values [][]string        `json:"values"`
		} `json:"result"`
	} `json:"data"`
}

// QueryLogs queries logs from Loki
func (c *Client) QueryLogs(ctx context.Context, runID uuid.UUID) ([]LogEntry, error) {
	// Build LogQL query
	query := fmt.Sprintf(`{run_id="%s"}`, runID.String())

	// Build query URL
	url := fmt.Sprintf("%s/loki/api/v1/query_range?query=%s&direction=forward&limit=10000", c.baseURL, query)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create query request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to query Loki: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Loki query failed: status=%d, body=%s", resp.StatusCode, string(body))
	}

	var queryResp QueryResponse
	if err := json.NewDecoder(resp.Body).Decode(&queryResp); err != nil {
		return nil, fmt.Errorf("failed to decode query response: %w", err)
	}

	// Parse response
	var entries []LogEntry
	for _, result := range queryResp.Data.Result {
		for _, value := range result.Values {
			if len(value) != 2 {
				continue
			}
			timestampNs := value[0]
			line := value[1]

			// Parse timestamp (nanoseconds)
			var ts int64
			fmt.Sscanf(timestampNs, "%d", &ts)
			timestamp := time.Unix(0, ts)

			entries = append(entries, LogEntry{
				Timestamp: timestamp,
				Line:      line,
			})
		}
	}

	return entries, nil
}
