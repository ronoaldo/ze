package agent

import (
	"encoding/json"
	"testing"

	"github.com/ronoaldo/ze/internal/llm"
	"github.com/ronoaldo/ze/internal/tools"
)

// mockLLMClient simulates LLM responses for testing.
type mockLLMClient struct {
	responses []llm.ChatMessage
	callCount int
	infiniteToolCalls bool
}

func (m *mockLLMClient) Chat(req *llm.ChatRequest) (*llm.ChatResponse, error) {
	m.callCount++
	var msg llm.ChatMessage
	if m.infiniteToolCalls {
		args, _ := json.Marshal(map[string]interface{}{"path": "test.go", "content": "y"})
		msg = llm.ChatMessage{
			Role: "assistant",
			ToolCalls: []llm.ToolCall{
				{
					ID:   "call_1",
					Type: "function",
					Function: llm.ToolCallFunction{
						Name:      "write_file",
						Arguments: args,
					},
				},
			},
		}
	} else if m.callCount <= len(m.responses) {
		msg = m.responses[m.callCount-1]
	} else {
		msg = llm.ChatMessage{Role: "assistant", Content: "Final answer after tool use."}
	}
	return &llm.ChatResponse{
		Choices: []struct {
			Message llm.ChatMessage `json:"message"`
			Finish  string          `json:"finish_reason"`
		}{
			{Message: msg, Finish: "stop"},
		},
	}, nil
}

func (m *mockLLMClient) ListModels() ([]llm.ModelInfo, error) {
	return nil, nil
}

// newTestAgent creates an agent with a temp dir for file operations.
func newTestAgent(t *testing.T, mock *mockLLMClient, toolList []tools.Tool) *Agent {
	t.Helper()
	tmpDir := t.TempDir()

	fixedTools := make([]tools.Tool, 0, len(toolList))
	for _, tool := range toolList {
		switch tool.(type) {
		case *tools.FileReadTool:
			fixedTools = append(fixedTools, &tools.FileReadTool{BaseDir: tmpDir})
		case *tools.FileWriteTool:
			fixedTools = append(fixedTools, &tools.FileWriteTool{BaseDir: tmpDir})
		case *tools.ListFilesTool:
			fixedTools = append(fixedTools, &tools.ListFilesTool{BaseDir: tmpDir})
		default:
			fixedTools = append(fixedTools, tool)
		}
	}

	return NewAgent(mock, "gemma-4-9b", fixedTools)
}

func TestRun_NoToolCall_ReturnsDirectAnswer(t *testing.T) {
	mock := &mockLLMClient{responses: []llm.ChatMessage{
		{Role: "assistant", Content: "Hello, I can help you with code!"},
	}}
	agent := newTestAgent(t, mock, nil)

	resp, err := agent.Run("What can you do?")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp != "Hello, I can help you with code!" {
		t.Errorf("expected 'Hello, I can help you with code!', got '%s'", resp)
	}
	if mock.callCount != 1 {
		t.Errorf("expected 1 LLM call, got %d", mock.callCount)
	}
}

func TestRun_WithToolCall_ReCallsLLM(t *testing.T) {
	args, _ := json.Marshal(map[string]interface{}{"path": "test.go", "content": "package main"})
	mock := &mockLLMClient{responses: []llm.ChatMessage{
		{
			Role: "assistant",
			ToolCalls: []llm.ToolCall{
				{
					ID:   "call_1",
					Type: "function",
					Function: llm.ToolCallFunction{
						Name:      "write_file",
						Arguments: args,
					},
				},
			},
		},
	}}
	agent := newTestAgent(t, mock, []tools.Tool{&tools.FileWriteTool{}})

	resp, err := agent.Run("Write a Go file")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp != "Final answer after tool use." {
		t.Errorf("expected 'Final answer after tool use.', got '%s'", resp)
	}
	if mock.callCount != 2 {
		t.Errorf("expected 2 LLM calls (1st: tool call, 2nd: final answer), got %d", mock.callCount)
	}
}

func TestRun_MaxIterations_WhenOnlyToolCalls(t *testing.T) {
	mock := &mockLLMClient{infiniteToolCalls: true}
	agent := newTestAgent(t, mock, []tools.Tool{&tools.FileWriteTool{}})

	_, err := agent.Run("Write something")
	if err == nil {
		t.Fatal("expected max iterations error")
	}
	if mock.callCount != maxIterations {
		t.Errorf("expected %d LLM calls, got %d", maxIterations, mock.callCount)
	}
}

func TestRun_MultipleToolCallsInOneResponse(t *testing.T) {
	args1, _ := json.Marshal(map[string]interface{}{"path": "a.go", "content": "package main"})
	args2, _ := json.Marshal(map[string]interface{}{"path": "b.go", "content": "package main"})
	mock := &mockLLMClient{responses: []llm.ChatMessage{
		{
			Role: "assistant",
			ToolCalls: []llm.ToolCall{
				{
					ID:   "call_1",
					Type: "function",
					Function: llm.ToolCallFunction{
						Name:      "write_file",
						Arguments: args1,
					},
				},
				{
					ID:   "call_2",
					Type: "function",
					Function: llm.ToolCallFunction{
						Name:      "write_file",
						Arguments: args2,
					},
				},
			},
		},
	}}
	agent := newTestAgent(t, mock, []tools.Tool{&tools.FileWriteTool{}})

	resp, err := agent.Run("Write two files")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp != "Final answer after tool use." {
		t.Errorf("expected 'Final answer after tool use.', got '%s'", resp)
	}
	if mock.callCount != 2 {
		t.Errorf("expected 2 LLM calls, got %d", mock.callCount)
	}
}

func TestRun_UnknownTool_ReturnsErrorInResult(t *testing.T) {
	args, _ := json.Marshal(map[string]interface{}{"arg": "val"})
	mock := &mockLLMClient{responses: []llm.ChatMessage{
		{
			Role: "assistant",
			ToolCalls: []llm.ToolCall{
				{
					ID:   "call_1",
					Type: "function",
					Function: llm.ToolCallFunction{
						Name:      "unknown_tool",
						Arguments: args,
					},
				},
			},
		},
	}}
	agent := newTestAgent(t, mock, nil)

	resp, err := agent.Run("Use unknown tool")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp != "Final answer after tool use." {
		t.Errorf("expected 'Final answer after tool use.', got '%s'", resp)
	}
	if mock.callCount != 2 {
		t.Errorf("expected 2 LLM calls, got %d", mock.callCount)
	}
}
