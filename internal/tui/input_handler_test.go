package tui

import (
	"errors"
	"testing"

	"github.com/ronoaldo/ze/internal/agent"
	"github.com/ronoaldo/ze/internal/commands"
)

func TestInputHandler_NormalCommands(t *testing.T) {
	cmdExecutor := func(input string) (string, error) {
		if input == "/help" {
			return "help content", nil
		}
		if input == "/quit" {
			return "", commands.ErrQuit
		}
		return "", errors.New("unknown command")
	}
	agentExecutor := func(input string) (string, agent.AgentStats, error) {
		return "", agent.AgentStats{}, nil
	}

	h := NewInputHandler(cmdExecutor, agentExecutor)

	// Test /help
	resp, _, err := h.Process("/help")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp != "help content" {
		t.Errorf("expected 'help content', got %q", resp)
	}

	// Test /quit
	_, _, err = h.Process("/quit")
	if !errors.Is(err, commands.ErrQuit) {
		t.Errorf("expected ErrQuit, got %v", err)
	}

	// Test unknown command (starts with /)
	resp, _, err = h.Process("/unknown")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp != "Error: unknown command" {
		t.Errorf("expected 'Error: unknown command', got %q", resp)
	}
}

func TestInputHandler_AgentMessages(t *testing.T) {
	cmdExecutor := func(input string) (string, error) {
		return "", errors.New("unknown command")
	}
	agentExecutor := func(input string) (string, agent.AgentStats, error) {
		return "agent response to " + input, agent.AgentStats{}, nil
	}

	h := NewInputHandler(cmdExecutor, agentExecutor)

	// Test normal message
	resp, _, err := h.Process("hello agent")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp != "agent response to hello agent" {
		t.Errorf("expected 'agent response to hello agent', got %q", resp)
	}
}

func TestInputHandler_MultilineActivation(t *testing.T) {
	cmdExecutor := func(input string) (string, error) {
		return "", nil
	}
	agentExecutor := func(input string) (string, agent.AgentStats, error) {
		return "", agent.AgentStats{}, nil
	}

	h := NewInputHandler(cmdExecutor, agentExecutor)

	// Test activation
	resp, _, err := h.Process("/multiline")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp != "Multiline mode enabled. Type your message and end with '/send' to send." {
		t.Errorf("unexpected activation message: %q", resp)
	}
	if !h.isMultiline {
		t.Error("expected isMultiline to be true")
	}
}

func TestInputHandler_MultilineAccumulation(t *testing.T) {
	cmdExecutor := func(input string) (string, error) {
		return "", nil
	}
	agentExecutor := func(input string) (string, agent.AgentStats, error) {
		t.Error("agent executor should not be called during accumulation")
		return "", agent.AgentStats{}, nil
	}

	h := NewInputHandler(cmdExecutor, agentExecutor)
	h.isMultiline = true

	// Line 1
	resp, _, err := h.Process("line 1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp != "" {
		t.Errorf("expected empty response during accumulation, got %q", resp)
	}

	// Line 2
	resp, _, err = h.Process("line 2")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp != "" {
		t.Errorf("expected empty response during accumulation, got %q", resp)
	}

	if h.multilineBuffer.String() != "line 1\nline 2\n" {
		t.Errorf("unexpected buffer content: %q", h.multilineBuffer.String())
	}
}

func TestInputHandler_MultilineWithEmptyLines(t *testing.T) {
	cmdExecutor := func(input string) (string, error) {
		return "", nil
	}
	agentExecutor := func(input string) (string, agent.AgentStats, error) {
		if input != "line 1\n\nline 2\n" {
			t.Errorf("expected 'line 1\\n\\nline 2\\n', got %q", input)
		}
		return "final response", agent.AgentStats{}, nil
	}

	h := NewInputHandler(cmdExecutor, agentExecutor)
	h.isMultiline = true

	// Line 1
	h.Process("line 1")
	// Empty Line
	h.Process("")
	// Line 2
	h.Process("line 2")

	// Test /send
	resp, _, err := h.Process("/send")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp != "final response" {
		t.Errorf("expected 'final response', got %q", resp)
	}
	if h.multilineBuffer.Len() != 0 {
		t.Error("expected buffer to be empty after /send")
	}
}

func TestInputHandler_MultilineCompletion(t *testing.T) {
	cmdExecutor := func(input string) (string, error) {
		return "", nil
	}
	agentExecutor := func(input string) (string, agent.AgentStats, error) {
		if input != "line 1\nline 2\n" {
			t.Errorf("expected 'line 1\\nline 2\\n', got %q", input)
		}
		return "final response", agent.AgentStats{}, nil
	}

	h := NewInputHandler(cmdExecutor, agentExecutor)
	h.isMultiline = true
	h.multilineBuffer.WriteString("line 1\nline 2\n")

	// Test /send
	resp, _, err := h.Process("/send")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp != "final response" {
		t.Errorf("expected 'final response', got %q", resp)
	}
	if h.isMultiline {
		t.Error("expected isMultiline to be false after /send")
	}
	if h.multilineBuffer.Len() != 0 {
		t.Error("expected buffer to be empty after /send")
	}
}
