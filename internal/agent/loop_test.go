package agent

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/ronoaldo/ze/internal/llm"
	"github.com/ronoaldo/ze/internal/tools"
)

// MockClient simulates LLM responses for testing.
type MockClient struct {
	ChatFunc       func(req *llm.ChatRequest) (*llm.ChatResponse, error)
	ListModelsFunc func() ([]llm.ModelInfo, error)
}

func (m *MockClient) Chat(req *llm.ChatRequest) (*llm.ChatResponse, error) {
	return m.ChatFunc(req)
}

func (m *MockClient) ListModels() ([]llm.ModelInfo, error) {
	return m.ListModelsFunc()
}

// MockTool is a simple tool implementation for testing.
type MockTool struct {
	name    string
	execute func(args map[string]interface{}) (tools.ToolResult, error)
	schema  map[string]interface{}
}

func (m *MockTool) Name() string { return m.name }
func (m *MockTool) Execute(args map[string]interface{}) (tools.ToolResult, error) {
	return m.execute(args)
}
func (m *MockTool) JSONSchema() map[string]interface{} { return m.schema }

// Helper to create a ChatResponse with a single choice
func createChatResponse(msg llm.ChatMessage) *llm.ChatResponse {
	return &llm.ChatResponse{
		Choices: []struct {
			Message llm.ChatMessage `json:"message"`
			Finish  string          `json:"finish_reason"`
		}{
			{
				Message: msg,
				Finish:  "stop",
			},
		},
	}
}

func TestAgent_Run_ShellCommand(t *testing.T) {
	mockClient := &MockClient{
		ChatFunc: func(req *llm.ChatRequest) (*llm.ChatResponse, error) {
			return createChatResponse(llm.ChatMessage{Content: "I received it"}), nil
		},
	}

	a := NewAgent(mockClient, "test-model", nil, false, 5, false)

	// 1. Execute shell command
	output, _, err := a.Run("!echo hello")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if !strings.Contains(output, "hello") {
		t.Errorf("Expected output to contain 'hello', got %q", output)
	}

	// 2. Call with actual prompt to check if command was injected
	var receivedPrompt string
	mockClient.ChatFunc = func(req *llm.ChatRequest) (*llm.ChatResponse, error) {
		for _, msg := range req.Messages {
			if msg.Role == "user" && strings.Contains(msg.Content, "User executed command") {
				receivedPrompt = msg.Content
				break
			}
		}
		return createChatResponse(llm.ChatMessage{Content: "Response"}), nil
	}

	_, _, err = a.Run("What was the command?")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	expectedSnippet := "User executed command:\n$ echo hello\nhello"
	if !strings.Contains(receivedPrompt, expectedSnippet) {
		t.Errorf("Prompt did not contain expected snippet.\nGot: %q", receivedPrompt)
	}
}

func TestAgent_Run_NoCommand(t *testing.T) {
	mockClient := &MockClient{
		ChatFunc: func(req *llm.ChatRequest) (*llm.ChatResponse, error) {
			return createChatResponse(llm.ChatMessage{Content: "Response"}), nil
		},
	}

	a := NewAgent(mockClient, "test-model", nil, false, 5, false)

	_, _, err := a.Run("Just a normal message")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(a.History) != 2 {
		t.Errorf("Expected 2 history messages, got %d", len(a.History))
	}
	if a.History[0].Content != "Just a normal message" {
		t.Errorf("Expected first content 'Just a normal message', got %q", a.History[0].Content)
	}
}

func TestAgent_Run_NoToolCall_ReturnsDirectAnswer(t *testing.T) {
	mockClient := &MockClient{
		ChatFunc: func(req *llm.ChatRequest) (*llm.ChatResponse, error) {
			return createChatResponse(llm.ChatMessage{Content: "Hello there!"}), nil
		},
	}

	a := NewAgent(mockClient, "test-model", nil, false, 5, false)

	resp, _, err := a.Run("Say hi")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if resp != "Hello there!" {
		t.Errorf("Expected 'Hello there!', got %q", resp)
	}

	if len(a.History) != 2 {
		t.Errorf("Expected 2 history messages (user + assistant), got %d", len(a.History))
	}
}

func TestAgent_Run_WithToolCall_ReCallsLLM(t *testing.T) {
	toolName := "test_tool"
	toolArgs := map[string]interface{}{"val": "hello"}
	argsJSON, _ := json.Marshal(toolArgs)

	mockTool := &MockTool{
		name: toolName,
		execute: func(args map[string]interface{}) (tools.ToolResult, error) {
			return tools.ToolResult{FullResult: "tool success"}, nil
		},
		schema: map[string]interface{}{
			"description": "a test tool",
			"parameters":  map[string]interface{}{},
		},
	}

	callCount := 0
	mockClient := &MockClient{
		ChatFunc: func(req *llm.ChatRequest) (*llm.ChatResponse, error) {
			callCount++
			if callCount == 1 {
				return createChatResponse(llm.ChatMessage{
					Role: "assistant",
					ToolCalls: []llm.ToolCall{
						{
							ID:   "call_1",
							Type: "function",
							Function: llm.ToolCallFunction{
								Name:      toolName,
								Arguments: argsJSON,
							},
						},
					},
				}), nil
			}
			return createChatResponse(llm.ChatMessage{Content: "Final answer"}), nil
		},
	}

	a := NewAgent(mockClient, "test-model", []tools.Tool{mockTool}, false, 5, false)

	resp, _, err := a.Run("Use the tool")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if resp != "Final answer" {
		t.Errorf("Expected 'Final answer', got %q", resp)
	}

	if callCount != 2 {
		t.Errorf("Expected 2 LLM calls, got %d", callCount)
	}

	// History should have: User, Assistant (tool call), Tool (result), Assistant (final)
	if len(a.History) != 4 {
		t.Errorf("Expected 4 history messages, got %d", len(a.History))
	}
}

func TestAgent_Run_MaxIterations_WhenOnlyToolCalls(t *testing.T) {
	mockTool := &MockTool{
		name: "infinite_tool",
		execute: func(args map[string]interface{}) (tools.ToolResult, error) {
			return tools.ToolResult{FullResult: "still working"}, nil
		},
		schema: map[string]interface{}{
			"description": "an infinite tool",
			"parameters":  map[string]interface{}{},
		},
	}

	argsJSON, _ := json.Marshal(map[string]interface{}{})
	mockClient := &MockClient{
		ChatFunc: func(req *llm.ChatRequest) (*llm.ChatResponse, error) {
			return createChatResponse(llm.ChatMessage{
				Role: "assistant",
				ToolCalls: []llm.ToolCall{
					{
						ID:   "call_1",
						Type: "function",
						Function: llm.ToolCallFunction{
							Name:      "infinite_tool",
							Arguments: argsJSON,
						},
					},
				},
			}), nil
		},
	}

	// Limit to 2 iterations
	a := NewAgent(mockClient, "test-model", []tools.Tool{mockTool}, false, 2, false)

	_, _, err := a.Run("Keep using tool")
	if err == nil {
		t.Fatal("Expected error due to max iterations, but got nil")
	}

	if !strings.Contains(err.Error(), "reached max iterations") {
		t.Errorf("Expected max iterations error, got: %v", err)
	}
}

func TestAgent_Run_MultipleToolCallsInOneResponse(t *testing.T) {
	tool1 := &MockTool{
		name: "tool1",
		execute: func(args map[string]interface{}) (tools.ToolResult, error) {
			return tools.ToolResult{FullResult: "res1"}, nil
		},
		schema: map[string]interface{}{
			"description": "tool 1",
			"parameters":  map[string]interface{}{},
		},
	}
	tool2 := &MockTool{
		name: "tool2",
		execute: func(args map[string]interface{}) (tools.ToolResult, error) {
			return tools.ToolResult{FullResult: "res2"}, nil
		},
		schema: map[string]interface{}{
			"description": "tool 2",
			"parameters":  map[string]interface{}{},
		},
	}

	args1JSON, _ := json.Marshal(map[string]interface{}{})
	args2JSON, _ := json.Marshal(map[string]interface{}{})

	callCount := 0
	mockClient := &MockClient{
		ChatFunc: func(req *llm.ChatRequest) (*llm.ChatResponse, error) {
			callCount++
			if callCount == 1 {
				return createChatResponse(llm.ChatMessage{
					Role: "assistant",
					ToolCalls: []llm.ToolCall{
						{
							ID:   "call_1",
							Type: "function",
							Function: llm.ToolCallFunction{
								Name:      "tool1",
								Arguments: args1JSON,
							},
						},
						{
							ID:   "call_2",
							Type: "function",
							Function: llm.ToolCallFunction{
								Name:      "tool2",
								Arguments: args2JSON,
							},
						},
					},
				}), nil
			}
			return createChatResponse(llm.ChatMessage{Content: "Done with both"}), nil
		},
	}

	a := NewAgent(mockClient, "test-model", []tools.Tool{tool1, tool2}, false, 5, false)

	resp, _, err := a.Run("Use both")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if resp != "Done with both" {
		t.Errorf("Expected 'Done with both', got %q", resp)
	}

	if callCount != 2 {
		t.Errorf("Expected 2 LLM calls, got %d", callCount)
	}

	// History should have: User, Assistant (2 tool calls), Tool (res1), Tool (res2), Assistant (final)
	if len(a.History) != 5 {
		t.Errorf("Expected 5 history messages, got %d", len(a.History))
	}
}

func TestAgent_Run_ToolCallID_Is_Correctly_Set(t *testing.T) {
	toolName := "test_tool"
	tcID := "call_abc_123"
	argsJSON, _ := json.Marshal(map[string]interface{}{})

	mockTool := &MockTool{
		name: toolName,
		execute: func(args map[string]interface{}) (tools.ToolResult, error) {
			return tools.ToolResult{FullResult: "tool success"}, nil
		},
		schema: map[string]interface{}{
			"description": "a test tool",
			"parameters":  map[string]interface{}{},
		},
	}

	callCount := 0
	mockClient := &MockClient{
		ChatFunc: func(req *llm.ChatRequest) (*llm.ChatResponse, error) {
			callCount++
			if callCount == 1 {
				return createChatResponse(llm.ChatMessage{
					Role: "assistant",
					ToolCalls: []llm.ToolCall{
						{
							ID:   tcID,
							Type: "function",
							Function: llm.ToolCallFunction{
								Name:      toolName,
								Arguments: argsJSON,
							},
						},
					},
				}), nil
			}
			return createChatResponse(llm.ChatMessage{Content: "Done!"}), nil
		},
	}

	a := NewAgent(mockClient, "test-model", []tools.Tool{mockTool}, false, 5, false)

	_, _, err := a.Run("Use tool")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Check for the tool message in history
	foundCorrectID := false
	for _, msg := range a.History {
		if msg.Role == "tool" && msg.ToolCallID == tcID {
			foundCorrectID = true
			break
		}
	}

	if !foundCorrectID {
		t.Error("Did not find a tool message with the correct ToolCallID in history")
	}
}
