package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Analyzer handles AI-powered log analysis using OpenAI API
type Analyzer struct {
	apiKey     string
	baseURL    string
	model      string
	maxTokens  int
	httpClient *http.Client
}

// Config holds analyzer configuration
type Config struct {
	APIKey    string
	BaseURL   string // Optional: custom OpenAI-compatible endpoint
	Model     string
	MaxTokens int
}

// NewAnalyzer creates a new AI analyzer
func NewAnalyzer(cfg *Config) *Analyzer {
	if cfg.Model == "" {
		cfg.Model = "gpt-4o-mini"
	}
	if cfg.MaxTokens == 0 {
		cfg.MaxTokens = 500
	}
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://api.openai.com/v1"
	}

	return &Analyzer{
		apiKey:    cfg.APIKey,
		baseURL:   cfg.BaseURL,
		model:     cfg.Model,
		maxTokens: cfg.MaxTokens,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// OpenAIRequest represents the request to OpenAI API
type OpenAIRequest struct {
	Model     string    `json:"model"`
	Messages  []Message `json:"messages"`
	MaxTokens int       `json:"max_tokens"`
}

// Message represents a chat message
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// OpenAIResponse represents the response from OpenAI API
type OpenAIResponse struct {
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
	Usage struct {
		TotalTokens int `json:"total_tokens"`
	} `json:"usage"`
}

// AnalysisResult contains the AI analysis result
type AnalysisResult struct {
	Report      string
	TokensUsed  int
	GeneratedAt time.Time
}

// AnalyzeLogs analyzes log content and generates a report
func (a *Analyzer) AnalyzeLogs(ctx context.Context, logs []string, exitCode int32, runStatus string) (*AnalysisResult, error) {
	// Prepare log content (limit to prevent token overflow)
	logContent := prepareLogs(logs, 5000) // Max 5000 lines

	// Create prompt
	prompt := buildPrompt(logContent, exitCode, runStatus)

	// Call OpenAI API
	req := OpenAIRequest{
		Model: a.model,
		Messages: []Message{
			{
				Role:    "system",
				Content: "You are an expert log analyzer. Analyze the provided script execution logs and provide a concise summary highlighting key events, errors, and outcomes.",
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		MaxTokens: a.maxTokens,
	}

	report, tokensUsed, err := a.callOpenAI(ctx, req)
	if err != nil {
		return nil, err
	}

	return &AnalysisResult{
		Report:      report,
		TokensUsed:  tokensUsed,
		GeneratedAt: time.Now(),
	}, nil
}

// callOpenAI makes a request to the OpenAI API
func (a *Analyzer) callOpenAI(ctx context.Context, req OpenAIRequest) (string, int, error) {
	// Marshal request
	body, err := json.Marshal(req)
	if err != nil {
		return "", 0, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request using configured base URL
	url := fmt.Sprintf("%s/chat/completions", a.baseURL)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return "", 0, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+a.apiKey)

	// Send request
	resp, err := a.httpClient.Do(httpReq)
	if err != nil {
		return "", 0, fmt.Errorf("failed to call OpenAI API: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", 0, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", 0, fmt.Errorf("OpenAI API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	// Parse response
	var openAIResp OpenAIResponse
	if err := json.Unmarshal(respBody, &openAIResp); err != nil {
		return "", 0, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(openAIResp.Choices) == 0 {
		return "", 0, fmt.Errorf("no choices in response")
	}

	report := openAIResp.Choices[0].Message.Content
	tokensUsed := openAIResp.Usage.TotalTokens

	return report, tokensUsed, nil
}

// prepareLogs limits log content to avoid token overflow
func prepareLogs(logs []string, maxLines int) string {
	if len(logs) <= maxLines {
		return strings.Join(logs, "\n")
	}

	// Take first 40% and last 60% of logs
	firstPart := int(float64(maxLines) * 0.4)
	lastPart := maxLines - firstPart

	var builder strings.Builder
	for i := 0; i < firstPart; i++ {
		builder.WriteString(logs[i])
		builder.WriteString("\n")
	}

	builder.WriteString(fmt.Sprintf("\n... [%d lines omitted] ...\n\n", len(logs)-maxLines))

	for i := len(logs) - lastPart; i < len(logs); i++ {
		builder.WriteString(logs[i])
		builder.WriteString("\n")
	}

	return builder.String()
}

// buildPrompt creates the analysis prompt
func buildPrompt(logContent string, exitCode int32, runStatus string) string {
	var builder strings.Builder

	builder.WriteString("Analyze the following script execution logs:\n\n")
	builder.WriteString("Execution Status: ")
	builder.WriteString(runStatus)
	builder.WriteString("\n")
	builder.WriteString(fmt.Sprintf("Exit Code: %d\n\n", exitCode))
	builder.WriteString("Logs:\n")
	builder.WriteString(logContent)
	builder.WriteString("\n\n")
	builder.WriteString("Please provide:\n")
	builder.WriteString("1. A brief summary of what the script did\n")
	builder.WriteString("2. Key events or milestones\n")
	if runStatus == "failed" {
		builder.WriteString("3. The root cause of the failure (specific line/error if possible)\n")
		builder.WriteString("4. Suggested fixes or next steps\n")
	} else {
		builder.WriteString("3. Any warnings or noteworthy observations\n")
	}

	return builder.String()
}
