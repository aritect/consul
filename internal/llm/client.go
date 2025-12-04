package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Provider string

const (
	ProviderGroq   Provider = "groq"
	ProviderOpenAI Provider = "openai"
)

type Client struct {
	provider   Provider
	apiKey     string
	model      string
	baseURL    string
	httpClient *http.Client
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
	Temperature float64       `json:"temperature,omitempty"`
}

type ChatResponse struct {
	ID      string `json:"id"`
	Choices []struct {
		Message      ChatMessage `json:"message"`
		FinishReason string      `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error,omitempty"`
}

func NewClient(provider Provider, apiKey, model string) *Client {
	baseURL := ""
	switch provider {
	case ProviderGroq:
		baseURL = "https://api.groq.com/openai/v1"
		if model == "" {
			model = "meta-llama/llama-4-scout-17b-16e-instruct"
		}
	case ProviderOpenAI:
		baseURL = "https://api.openai.com/v1"
		if model == "" {
			model = "gpt-4o-mini"
		}
	}

	return &Client{
		provider: provider,
		apiKey:   apiKey,
		model:    model,
		baseURL:  baseURL,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

func (c *Client) Chat(messages []ChatMessage) (string, error) {
	return c.ChatWithOptions(messages, 2048, 0.7)
}

func (c *Client) ChatWithOptions(messages []ChatMessage, maxTokens int, temperature float64) (string, error) {
	reqBody := ChatRequest{
		Model:       c.model,
		Messages:    messages,
		MaxTokens:   maxTokens,
		Temperature: temperature,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", c.baseURL+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	var chatResp ChatResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if chatResp.Error != nil {
		return "", fmt.Errorf("API error: %s", chatResp.Error.Message)
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("no response choices returned")
	}

	return chatResp.Choices[0].Message.Content, nil
}

func (c *Client) GetProvider() Provider {
	return c.provider
}

func (c *Client) GetModel() string {
	return c.model
}

func ParseProvider(s string) (Provider, bool) {
	switch s {
	case "groq":
		return ProviderGroq, true
	case "openai":
		return ProviderOpenAI, true
	default:
		return "", false
	}
}
