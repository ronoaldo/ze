package tui

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/ronoaldo/ze/internal/agent"
	"github.com/ronoaldo/ze/internal/commands"
)

func TestTUI_Integration(t *testing.T) {
	// Mock dependencies
	dummyAgent := &agent.Agent{}

	cmdExecutor := func(a *agent.Agent, input string) (string, error) {
		if input == "/quit" {
			return "", commands.ErrQuit
		}
		return "command output", nil
	}
	agentExecutor := func(ctx context.Context, input string) (string, agent.AgentStats, error) {
		return "agent response to " + input, agent.AgentStats{}, nil
	}

	h := NewInputHandler(cmdExecutor, agentExecutor)

	t.Run("Basic Input/Output Flow", func(t *testing.T) {
		input := "hello\n"
		r := strings.NewReader(input)
		w := &bytes.Buffer{}

		tui := NewTestTUI(r, w)

		err := tui.Run(context.Background(), func(ctx context.Context, msg string) (string, agent.AgentStats, error) {
			return h.Process(ctx, dummyAgent, msg)
		}, nil)

		if err != nil && !errors.Is(err, io.EOF) {
			t.Fatalf("unexpected error: %v", err)
		}

		output := w.String()
		if !strings.Contains(output, "agent response to hello") {
			t.Errorf("expected output to contain 'agent response to hello', got %q", output)
		}
	})

	t.Run("Command Execution", func(t *testing.T) {
		input := "/help\n"
		r := strings.NewReader(input)
		w := &bytes.Buffer{}

		tui := NewTestTUI(r, w)

		err := tui.Run(context.Background(), func(ctx context.Context, msg string) (string, agent.AgentStats, error) {
			return h.Process(ctx, dummyAgent, msg)
		}, nil)

		if err != nil && !errors.Is(err, io.EOF) {
			t.Fatalf("unexpected error: %v", err)
		}

		output := w.String()
		if !strings.Contains(output, "command output") {
			t.Errorf("expected output to contain 'command output', got %q", output)
		}
	})
}
