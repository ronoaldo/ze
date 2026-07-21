package commands

import (
	"errors"
	"strings"
	"testing"

	"github.com/ronoaldo/ze/internal/agent"
)

func TestExecuteCommand(t *testing.T) {
	// Reset Registry for testing
	Registry = make(map[string]Command)
	RegisterCommands()

	// Dummy agent for testing
	dummyAgent := &agent.Agent{}

	tests := []struct {
		name           string
		input          string
		expectedOutput string
		expectedErr    error
	}{
		{
			name:           "quit command",
			input:          "/quit",
			expectedOutput: "",
			expectedErr:    ErrQuit,
		},
		{
			name:           "exit command",
			input:          "/exit",
			expectedOutput: "",
			expectedErr:    ErrQuit,
		},
		{
			name:           "help command",
			input:          "/help",
			expectedOutput: "Available commands:",
			expectedErr:    nil,
		},
		{
			name:           "unknown command",
			input:          "/unknown",
			expectedOutput: "",
			expectedErr:    errors.New("unknown command: /unknown"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ExecuteCommand(dummyAgent, tt.input)

			if tt.expectedErr != nil {
				if err == nil || !strings.Contains(err.Error(), tt.expectedErr.Error()) {
					t.Errorf("ExecuteCommand() error = %v, want %v", err, tt.expectedErr)
				}
			} else {
				if err != nil {
					t.Errorf("ExecuteCommand() unexpected error = %v", err)
				}
				if !strings.Contains(got, tt.expectedOutput) {
					t.Errorf("ExecuteCommand() = %q, want to contain %q", got, tt.expectedOutput)
				}
			}
		})
	}
}
