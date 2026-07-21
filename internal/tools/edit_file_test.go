package tools

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEditFileTool(t *testing.T) {
	tmpDir := t.TempDir()
	tool := &EditFileTool{BaseDir: tmpDir}

	tests := []struct {
		name       string
		initial    string
		edits      []Edit
		expected   string
		expectErr  bool
		errMessage string
	}{
		{
			name:      "Simple replace",
			initial:   "Hello, world! This is a test.",
			edits:     []Edit{{OldString: "world", NewString: "universe"}},
			expected:  "Hello, universe! This is a test.",
			expectErr: false,
		},
		{
			name:      "Multiple replace with replaceAll",
			initial:   "test test test",
			edits:     []Edit{{OldString: "test", NewString: "success", ReplaceAll: true}},
			expected:  "success success success",
			expectErr: false,
		},
		{
			name:       "Old string not found",
			initial:    "Hello world",
			edits:      []Edit{{OldString: "missing", NewString: "here"}},
			expected:   "",
			expectErr:  true,
			errMessage: "oldString not found",
		},
		{
			name:       "Multiple occurrences without replaceAll",
			initial:    "test test",
			edits:      []Edit{{OldString: "test", NewString: "fail", ReplaceAll: false}},
			expected:   "",
			expectErr:  true,
			errMessage: "found multiple times",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath := filepath.Join(tmpDir, "test.txt")
			err := os.WriteFile(filePath, []byte(tt.initial), 0644)
			if err != nil {
				t.Fatalf("failed to create test file: %v", err)
			}

			result, err := tool.Execute(map[string]interface{}{
				"path":  filepath.Base(filePath),
				"edits": tt.edits,
			})

			if tt.expectErr {
				if err == nil {
					t.Errorf("expected error but got nil")
				} else if tt.errMessage != "" && !strings.Contains(err.Error(), tt.errMessage) {
					t.Errorf("expected error containing %q, got %q", tt.errMessage, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				content, _ := os.ReadFile(filePath)
				if string(content) != tt.expected {
					t.Errorf("expected %q, got %q", tt.expected, string(content))
				}
				if result.FullResult == "" {
					t.Errorf("expected success message, got empty string")
				}
			}
		})
	}
}
