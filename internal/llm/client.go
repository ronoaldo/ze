package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
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

// ToolCallFunction represents a tool call function.
type ToolCallFunction struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

func (tf *ToolCallFunction) UnmarshalJSON(data []byte) error {
	type Alias ToolCallFunction
	var aux Alias
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	*tf = ToolCallFunction(aux)

	// Handle double-encoded string in Arguments.
	// If the arguments are a JSON string (e.g., "\"{\\\"a\\\": 1}\""),
	// unmarshal it into a string, then unmarshal that string back into the RawMessage.
	var s string
	if err := json.Unmarshal(tf.Arguments, &s); err == nil {
		return json.Unmarshal([]byte(s), &tf.Arguments)
	}

	return nil
}

// ToolCall represents a tool call from the LLM.
type ToolCall struct {
	ID       string           `json:"id"`
	Type     string           `json:"type"`
	Function ToolCallFunction `json:"function"`
}

// ToolDefinition defines a tool available for the LLM.
type ToolDefinition struct {
	Type     string      `json:"type"`
	Function FunctionDef `json:"function"`
}

// FunctionDef defines the function properties for a tool.
type FunctionDef struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Parameters  json.RawMessage `json:"parameters"`
}

// ChatMessage represents a single message in the chat history.
type ChatMessage struct {
	Role            string     `json:"role"`
	Content         string     `json:"content,omitempty"`
	ReasoningContent string    `json:"reasoning_content,omitempty"`
	ToolCalls       []ToolCall `json:"tool_calls,omitempty"`
	ToolCallID      string     `json:"tool_call_id,omitempty"`
}

// Usage represents token usage in a chat completion.
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ChatRequest is the structure for llama-server chat completion request (OpenAI compatible).
type ChatRequest struct {
	Model       string           `json:"model"`
	Messages    []ChatMessage    `json:"messages"`
	Temperature float64          `json:"temperature,omitempty"`
	MaxTokens   int              `json:"max_tokens,omitempty"`
	Tools       []ToolDefinition `json:"tools,omitempty"`
}

// ChatResponse is the structure for llama-server response.
type ChatResponse struct {
	Choices []struct {
		Message ChatMessage `json:"message"`
		Finish  string      `json:"finish_reason"`
	} `json:"choices"`
	Usage Usage `json:"usage"`
}

// Client defines the interface for interacting with an LLM server.
type Client interface {
	Chat(req *ChatRequest) (*ChatResponse, error)
	ListModels() ([]ModelInfo, error)
}

// LlamaServerClient is a client implementation for llama-server API.
type LlamaServerClient struct {
	BaseURL         string
	HTTP            *http.Client
	VerboseAPICalls bool
}

// NewLlamaServerClient creates a new client for llama-server API with a specified timeout.
func NewLlamaServerClient(baseURL string, timeout time.Duration, verboseAPICalls bool) *LlamaServerClient {
	return &LlamaServerClient{
		BaseURL:         baseURL,
		HTTP:            &http.Client{Timeout: timeout},
		VerboseAPICalls: verboseAPICalls,
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

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if c.VerboseAPICalls {
		fmt.Fprintf(os.Stderr, "\n--- [API REQUEST] ---\n%s\n", string(data))
		fmt.Fprintf(os.Stderr, "--- [API RESPONSE] ---\n%s\n--------------------\n", string(body))
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned error status %d: %s", resp.StatusCode, string(body))
	}

	var chatResp ChatResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w. Body: %s", err, string(body))
	}

	return &chatResp, nil
}

// ListModels retrieves available models from the llama-server.
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
