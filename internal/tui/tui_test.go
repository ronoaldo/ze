package tui

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/ronoaldo/ze/internal/agent"
	"github.com/ronoaldo/ze/internal/tools"
)

func TestNew(t *testing.T) {
	tui := New(false, false, false)
	if tui == nil {
		t.Fatal("New() returned nil")
	}
}

func TestSummarizeArgs(t *testing.T) {
	tui := New(false, false, false)
	tests := []struct {
		name     string
		toolName string
		args     string
		expected string
	}{
		{
			name:     "read_file path",
			toolName: "read_file",
			args:     `{"path": "test.txt"}`,
			expected: "'test.txt'",
		},
		{
			name:     "write_file path",
			toolName: "write_file",
			args:     `{"path": "test.txt", "content": "hi"}`,
			expected: "'test.txt'",
		},
		{
			name:     "list_files path",
			toolName: "list_files",
			args:     `{"path": "."}`,
			expected: "'.'",
		},
		{
			name:     "go_tool doc package",
			toolName: "go_tool",
			args:     `{"action": "doc", "package": "fmt"}`,
			expected: "doc(fmt)",
		},
		{
			name:     "git_tool diff empty",
			toolName: "git_tool",
			args:     `{}`,
			expected: "diff",
		},
		{
			name:     "git_tool add all",
			toolName: "git_tool",
			args:     `{"action": "add"}`,
			expected: "add(all)",
		},
		{
			name:     "invalid json",
			toolName: "read_file",
			args:     `{invalid}`,
			expected: `{invalid}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tui.summarizeArgs(tt.toolName, tt.args)
			if got != tt.expected {
				t.Errorf("summarizeArgs() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestReportToolExecution(t *testing.T) {
	t.Run("Standard Mode Success", func(t *testing.T) {
		buf := new(bytes.Buffer)
		tui := New(false, false, false)
		tui.w = buf

		res := tools.ToolResult{
			Summary:    "[+1/-1]",
			FullResult: "some content",
		}
		tui.ReportToolExecution("read_file", `{"path": "foo.go"}`, res, nil)

		output := buf.String()
		if !strings.Contains(output, "read_file") {
			t.Errorf("Expected tool name in output, got %q", output)
		}
		if !strings.Contains(output, "'foo.go'") {
			t.Errorf("Expected argument in output, got %q", output)
		}
	})

	t.Run("Standard Mode Error", func(t *testing.T) {
		buf := new(bytes.Buffer)
		tui := New(false, false, false)
		tui.w = buf

		tui.ReportToolExecution("read_file", `{"path": "foo.go"}`, tools.ToolResult{}, fmt.Errorf("file not found"))
		if !strings.Contains(buf.String(), "[ERROR] file not found") {
			t.Errorf("Expected error message in output, got %q", buf.String())
		}
	})
}

func TestReportStatus(t *testing.T) {
	buf := new(bytes.Buffer)
	tui := New(false, false, false)
	tui.w = buf

	stats := agent.AgentStats{
		Status:       "Running",
		TotalTokens:  100,
		PromptTokens: 50,
		CompTokens:   50,
		CompPerSec:   10,
		PromptPerSec: 20,
	}
	tui.ReportStatus(stats)
	output := buf.String()
	if !strings.Contains(output, "Status: Running") {
		t.Errorf("Expected status in output, got %q", output)
	}
}
