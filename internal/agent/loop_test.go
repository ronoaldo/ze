package agent

import (
	"testing"

	"ze/internal/llm"
	"ze/internal/tools"
)

// mockLLMClient simulates LLM responses for testing.
type mockLLMClient struct {
	responses      []string
	callCount      int
	alwaysToolCall bool
}

func (m *mockLLMClient) Chat(req *llm.ChatRequest) (*llm.ChatResponse, error) {
	m.callCount++
	content := ""
	if m.alwaysToolCall {
		content = "TOOL_CALL:write_file{\"path\":\"x.go\",\"content\":\"y\"}"
	} else if m.callCount <= len(m.responses) {
		content = m.responses[m.callCount-1]
	} else {
		content = "Final answer after tool use."
	}
	return &llm.ChatResponse{
		Choices: []struct {
			Message llm.ChatMessage `json:"message"`
			Finish  string          `json:"finish_reason"`
		}{
			{Message: llm.ChatMessage{Role: "assistant", Content: content}},
		},
	}, nil
}

func (m *mockLLMClient) ListModels() ([]llm.ModelInfo, error) {
	return nil, nil
}

// newTestAgent creates an agent with a temp dir for file operations.
func newTestAgent(t *testing.T, mock *mockLLMClient, toolList []tools.Tool) *Agent {
	t.Helper()
	// Create temp dir for file tools
	tmpDir := t.TempDir()

	// Replace file tools with ones that use temp dir
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
	mock := &mockLLMClient{responses: []string{"Hello, I can help you with code!"}}
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
	mock := &mockLLMClient{
		responses: []string{
			"TOOL_CALL:write_file{\"path\":\"test.go\",\"content\":\"package main\"}",
		},
	}
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
	mock := &mockLLMClient{alwaysToolCall: true}
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
	mock := &mockLLMClient{
		responses: []string{
			"TOOL_CALL:write_file{\"path\":\"a.go\",\"content\":\"package main\"}\nTOOL_CALL:write_file{\"path\":\"b.go\",\"content\":\"package main\"}",
		},
	}
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
	mock := &mockLLMClient{
		responses: []string{
			"TOOL_CALL:unknown_tool{\"arg\":\"val\"}",
		},
	}
	// No tools registered — tool should be reported as not found
	agent := newTestAgent(t, mock, nil)

	resp, err := agent.Run("Use unknown tool")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should still get a response (tool error logged, LLM called again)
	if resp != "Final answer after tool use." {
		t.Errorf("expected 'Final answer after tool use.', got '%s'", resp)
	}
	if mock.callCount != 2 {
		t.Errorf("expected 2 LLM calls, got %d", mock.callCount)
	}
}

// TestFile