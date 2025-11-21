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

// Language constants
const (
	LanguageEnglish = "en"
	LanguageChinese = "zh"
)

// Predefined system prompts
const (
	SystemPromptEnglish = "You are an expert log analyzer. Analyze the provided script execution logs and provide a concise summary highlighting key events, errors, and outcomes. Format your response using markdown for better readability."
	SystemPromptChinese = "你是一位专业的日志分析专家。请分析提供的脚本执行日志，并提供一个简洁的摘要，突出关键事件、错误和结果。请使用 markdown 格式化你的回复以提高可读性。"
)

// Analyzer handles AI-powered log analysis using OpenAI API
type Analyzer struct {
	apiKey         string
	baseURL        string
	model          string
	maxTokens      int
	systemPrompt   string
	promptLanguage string
	httpClient     *http.Client
}

// Config holds analyzer configuration
type Config struct {
	APIKey         string
	BaseURL        string // Optional: custom OpenAI-compatible endpoint
	Model          string
	MaxTokens      int
	PromptLanguage string // "en" or "zh"
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

	// Default to English if not specified
	if cfg.PromptLanguage == "" {
		cfg.PromptLanguage = LanguageEnglish
	}

	// Select system prompt based on language
	var systemPrompt string
	if cfg.PromptLanguage == LanguageChinese {
		systemPrompt = SystemPromptChinese
	} else {
		systemPrompt = SystemPromptEnglish
	}

	return &Analyzer{
		apiKey:         cfg.APIKey,
		baseURL:        cfg.BaseURL,
		model:          cfg.Model,
		maxTokens:      cfg.MaxTokens,
		systemPrompt:   systemPrompt,
		promptLanguage: cfg.PromptLanguage,
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
func (a *Analyzer) AnalyzeLogs(ctx context.Context, logs []string, exitCode int32, runStatus string, maxLogLines int, truncateStrategy string) (*AnalysisResult, error) {
	// Prepare log content based on user's truncation strategy
	logContent := prepareLogs(logs, maxLogLines, truncateStrategy)

	// Create prompt with language support
	prompt := buildPrompt(logContent, exitCode, runStatus, a.promptLanguage)

	// Call OpenAI API
	req := OpenAIRequest{
		Model: a.model,
		Messages: []Message{
			{
				Role:    "system",
				Content: a.systemPrompt,
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

// prepareLogs limits log content based on truncation strategy
func prepareLogs(logs []string, maxLines int, strategy string) string {
	if len(logs) <= maxLines {
		return strings.Join(logs, "\n")
	}

	var builder strings.Builder

	switch strategy {
	case "head":
		// Keep first N lines
		for i := 0; i < maxLines && i < len(logs); i++ {
			builder.WriteString(logs[i])
			builder.WriteString("\n")
		}
		builder.WriteString(fmt.Sprintf("\n... [%d lines omitted] ...\n", len(logs)-maxLines))

	case "tail":
		// Keep last N lines
		builder.WriteString(fmt.Sprintf("... [%d lines omitted] ...\n\n", len(logs)-maxLines))
		for i := len(logs) - maxLines; i < len(logs); i++ {
			builder.WriteString(logs[i])
			builder.WriteString("\n")
		}

	case "smart":
		// Keep first 40% and last 60% with summary
		firstPart := int(float64(maxLines) * 0.4)
		lastPart := maxLines - firstPart

		for i := 0; i < firstPart; i++ {
			builder.WriteString(logs[i])
			builder.WriteString("\n")
		}

		builder.WriteString(fmt.Sprintf("\n... [%d lines omitted] ...\n\n", len(logs)-maxLines))

		for i := len(logs) - lastPart; i < len(logs); i++ {
			builder.WriteString(logs[i])
			builder.WriteString("\n")
		}

	default:
		// Default to tail strategy
		return prepareLogs(logs, maxLines, "tail")
	}

	return builder.String()
}

// buildPrompt creates the analysis prompt
func buildPrompt(logContent string, exitCode int32, runStatus string, language string) string {
	var builder strings.Builder

	if language == LanguageChinese {
		// Chinese prompt
		builder.WriteString("分析以下脚本执行日志：\n\n")
		builder.WriteString("执行状态：")
		builder.WriteString(runStatus)
		builder.WriteString("\n")
		builder.WriteString(fmt.Sprintf("退出代码：%d\n\n", exitCode))
		builder.WriteString("日志内容：\n```\n")
		builder.WriteString(logContent)
		builder.WriteString("\n```\n\n")
		builder.WriteString("请使用Markdown格式提供分析报告，包含以下内容：\n\n")
		builder.WriteString("## 1. 执行摘要\n")
		builder.WriteString("简要说明脚本做了什么\n\n")
		builder.WriteString("## 2. 关键事件\n")
		builder.WriteString("列出重要的执行步骤和里程碑\n\n")
		if runStatus == "failed" {
			builder.WriteString("## 3. 失败原因\n")
			builder.WriteString("分析根本原因，指出具体的错误行或错误信息\n\n")
			builder.WriteString("## 4. 修复建议\n")
			builder.WriteString("提供具体的修复方法或下一步操作建议\n")
		} else {
			builder.WriteString("## 3. 注意事项\n")
			builder.WriteString("列出任何警告或值得注意的观察结果\n")
		}
		builder.WriteString("\n重要：请直接输出markdown格式的文本，不要将整个响应包裹在代码块中。")
	} else {
		// English prompt
		builder.WriteString("Analyze the following script execution logs:\n\n")
		builder.WriteString("Execution Status: ")
		builder.WriteString(runStatus)
		builder.WriteString("\n")
		builder.WriteString(fmt.Sprintf("Exit Code: %d\n\n", exitCode))
		builder.WriteString("Logs:\n```\n")
		builder.WriteString(logContent)
		builder.WriteString("\n```\n\n")
		builder.WriteString("Please provide an analysis report in Markdown format with the following sections:\n\n")
		builder.WriteString("## 1. Execution Summary\n")
		builder.WriteString("Brief overview of what the script did\n\n")
		builder.WriteString("## 2. Key Events\n")
		builder.WriteString("List important execution steps and milestones\n\n")
		if runStatus == "failed" {
			builder.WriteString("## 3. Failure Analysis\n")
			builder.WriteString("Root cause analysis with specific error lines or messages\n\n")
			builder.WriteString("## 4. Recommended Fixes\n")
			builder.WriteString("Specific suggestions for fixing the issue or next steps\n")
		} else {
			builder.WriteString("## 3. Observations\n")
			builder.WriteString("Any warnings or noteworthy findings\n")
		}
		builder.WriteString("\nImportant: Output the markdown text directly, do not wrap the entire response in a code block.")
	}

	return builder.String()
}
