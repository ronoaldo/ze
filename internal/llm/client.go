package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// ModelInfo represents information about a model known to the llama-server.
type ModelInfo struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Context string `json:"context_length,omitempty"`
	Status  string `json:"-"` // Populated from status.value, not from JSON directly
}

// modelListItem represents the raw JSON structure returned by /v1/models.
type modelListItem struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Context string `json:"context_length,omitempty"`
	Status  struct {
		Value string `json:"value"`
	} `json:"status"`
}

// modelListResponse represents the top-level /v1/models response.
type modelListResponse struct {
	Data []modelListItem `json:"data"`
}

// ChatRequest is the structure for llama-server chat completion request (OpenAI compatible).
type ChatRequest struct {
	Model       string         `json:"model"`
	Messages    []ChatMessage  `json:"messages"`
	Temperature float64        `json:"temperature,omitempty"`
	MaxTokens   int            `json:"max_tokens,omitempty"`
}

// ChatMessage represents a single message in the chat history.
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatResponse is the structure for llama-server response.
type ChatResponse struct {
	Choices []struct {
		Message ChatMessage `json:"message"`
		Finish  string      `json:"finish_reason"`
	} `json:"choices"`
}

// Client defines the interface for interacting with an LLM server.
type Client interface {
	Chat(req *ChatRequest) (*ChatResponse, error)
	ListModels() ([]ModelInfo, error)
}

// LlamaServerClient is a client implementation for llama-server API.
type LlamaServerClient struct {
	BaseURL string
	HTTP    *http.Client
}

func NewLlamaServerClient(baseURL string) *LlamaServerClient {
	return &LlamaServerClient{
		BaseURL: baseURL,
		HTTP: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// Chat sends a chat completion request to the llama-server.
func (c *LlamaServerClient) Chat(req *ChatRequest) (*ChatResponse, error) {
	data, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.HTTP.Post(c.BaseURL+"/v1/chat/completions", "application/json", bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("server returned error status %d: %s", resp.StatusCode, string(body))
	}

	var chatResp ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &chatResp, nil
}

// ListModels retrieves available models from the llama-server.
// It populates ModelInfo.Status from the server's status.value field.
func (c *LlamaServerClient) ListModels() ([]ModelInfo, error) {
	resp, err := c.HTTP.Get(c.BaseURL + "/v1/models")
	if err != nil {
		return nil, fmt.Errorf("failed to list models: %w", err)
	}
	defer resp.Body.Close()

	var result modelListResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode models list: %w", err)
	}

	models := make([]ModelInfo, 0, len(result.Data))
	for _, m := range result.Data {
		models = append(models, ModelInfo{
			ID:      m.ID,
			Name:    m.Name,
			Context: m.Context,
			Status:  m.Status.Value,
		})
	}

	return models, nil
}
