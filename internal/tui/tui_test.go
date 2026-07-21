package tui

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/ronoaldo/ze/internal/tools"
)

func TestNew(t *testing.T) {
	tui := New(false, false, false)
	if tui == nil {
		t.Fatal("New() returned nil")
	}
	if tui.verbose {
		t.Error("Expected verbose to be false")
	}

	tuiVerbose := New(true, false, false)
	if !tuiVerbose.verbose {
		t.Error("Expected verbose to be true")
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
			expected: "test.txt",
		},
		{
			name:     "write_file path",
			toolName: "write_file",
			args:     `{"path": "test.txt", "content": "hi"}`,
			expected: "test.txt",
		},
		{
			name:     "list_files path",
			toolName: "list_files",
			args:     `{"path": "."}`,
			expected: ".",
		},
		{
			name:     "go_doc package",
			toolName: "go_doc",
			args:     `{"package": "fmt"}`,
			expected: "fmt",
		},
		{
			name:     "invalid json",
			toolName: "read_file",
			args:     `{invalid}`,
			expected: `{invalid}`,
		},
		{
			name:     "empty args",
			toolName: "read_file",
			args:     `{}`,
			expected: "{}",
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

		// Expected: * [Bold][Cyan]read_file[Reset]('foo.go') [Green][+1/-1][Reset]
		// Since we can't easily predict all ANSI codes perfectly if they change,
		// let's check for key components.
		output := buf.String()
		if !strings.Contains(output, "read_file") {
			t.Errorf("Expected tool name in output, got %q", output)
		}
		if !strings.Contains(output, "foo.go") {
			t.Errorf("Expected argument in output, got %q", output)
		}
		if !strings.Contains(output, "[+1/-1]") {
			t.Errorf("Expected summary in output, got %q", output)
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

	t.Run("Verbose Mode", func(t *testing.T) {
		buf := new(bytes.Buffer)
		tui := New(true, false, false)
		tui.w = buf

		res := tools.ToolResult{
			Summary:    "[+1/-1]",
			FullResult: "some content",
		}
		tui.ReportToolExecution("read_file", `{"path": "foo.go"}`, res, nil)

		if !strings.Contains(buf.String(), "some content") {
			t.Errorf("Expected full result in output, got %q", buf.String())
		}
	})
}
