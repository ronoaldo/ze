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

func TestReportToolExecution(t *testing.T) {
	t.Run("Standard Mode Success", func(t *testing.T) {
		buf := new(bytes.Buffer)
		tui := New(false, false, false)
		tui.w = buf

		res := tools.ToolResult{
			Summary:    "[+1/-1]",
			FullResult: "some content",
		}
		// The agent should have already summarized the args before calling the reporter
		tui.ReportToolExecution("read_file", "'foo.go'", res, nil)

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
