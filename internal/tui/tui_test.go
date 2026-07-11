package tui

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	tui := New(false, false)
	if tui == nil {
		t.Fatal("New() returned nil")
	}
	if tui.verbose {
		t.Error("Expected verbose to be false")
	}

	tuiVerbose := New(true, false)
	if !tuiVerbose.verbose {
		t.Error("Expected verbose to be true")
	}
}

func TestSummarizeArgs(t *testing.T) {
	tui := New(false, false)
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

func TestSummarizeResult(t *testing.T) {
	tui := New(false, false)
	tests := []struct {
		name     string
		toolName string
		result   string
		expected string
	}{
		{
			name:     "read_file bytes",
			toolName: "read_file",
			result:   "some content",
			expected: "12 bytes",
		},
		{
			name:     "list_files items",
			toolName: "list_files",
			result:   "- file1.txt\n- file2.txt\n",
			expected: "2 items",
		},
		{
			name:     "list_files empty",
			toolName: "list_files",
			result:   "",
			expected: "0 items",
		},
		{
			name:     "write_file success",
			toolName: "write_file",
			result:   "Successfully wrote to file.txt",
			expected: "Success",
		},
		{
			name:     "go_doc success",
			toolName: "go_doc",
			result:   "package fmt...",
			expected: "Success",
		},
		{
			name:     "default success",
			toolName: "unknown",
			result:   "anything",
			expected: "Success",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tui.summarizeResult(tt.toolName, tt.result)
			if got != tt.expected {
				t.Errorf("summarizeResult() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestReportToolCall(t *testing.T) {
	t.Run("Standard Mode", func(t *testing.T) {
		buf := new(bytes.Buffer)
		tui := New(false, false)
		tui.w = buf

		tui.ReportToolCall("read_file", `{"path": "foo.go"}`)
		expected := "* \033[1m\033[36mread_file\033[0m('foo.go')"
		if buf.String() != expected {
			t.Errorf("Expected %q, got %q", expected, buf.String())
		}
	})

	t.Run("Verbose Mode", func(t *testing.T) {
		buf := new(bytes.Buffer)
		tui := New(true, false)
		tui.w = buf

		args := `{"path": "foo.go"}`
		tui.ReportToolCallVerbose("read_file", args)
		// Note the ANSI escape codes. Using strings.Contains to be safer.
		if !strings.Contains(buf.String(), "[TOOL CALL] read_file") {
			t.Errorf("Expected tool call in output, got %q", buf.String())
		}
		if !strings.Contains(buf.String(), args) {
			t.Errorf("Expected full args in output, got %q", buf.String())
		}
	})
}

func TestReportToolResult(t *testing.T) {
	t.Run("Standard Mode Success", func(t *testing.T) {
		buf := new(bytes.Buffer)
		tui := New(false, false)
		tui.w = buf

		tui.ReportToolResult("read_file", "hello world", nil)
		expected := " [\033[1m\033[32m11 bytes\033[0m]"
		if !strings.Contains(buf.String(), expected) {
			t.Errorf("Expected summary %q in output, got %q", expected, buf.String())
		}
	})

	t.Run("Standard Mode Error", func(t *testing.T) {
		buf := new(bytes.Buffer)
		tui := New(false, false)
		tui.w = buf

		tui.ReportToolResult("read_file", "", fmt.Errorf("file not found"))
		if !strings.Contains(buf.String(), "[ERROR] file not found") {
			t.Errorf("Expected error message in output, got %q", buf.String())
		}
	})

	t.Run("Verbose Mode Success", func(t *testing.T) {
		buf := new(bytes.Buffer)
		tui := New(true, false)
		tui.w = buf

		result := "full content"
		tui.ReportToolResult("read_file", result, nil)
		if !strings.Contains(buf.String(), result) {
			t.Errorf("Expected full result in output, got %q", buf.String())
		}
	})
}
