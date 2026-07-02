package tui

import (
	"strings"
	"testing"
)

func TestNew_TUINotNil(t *testing.T) {
	tui := New()
	if tui == nil {
		t.Fatal("New() returned nil")
	}
	if tui.rows <= 0 {
		t.Errorf("expected rows > 0, got %d", tui.rows)
	}
	if tui.cols <= 0 {
		t.Errorf("expected cols > 0, got %d", tui.cols)
	}
}

func TestRenderChatLines_UserMessage(t *testing.T) {
	tui := New()
	tui.chat = []Message{{Role: "user", Content: "Hello"}}
	tui.renderChatLines()

	if len(tui.lines) != 1 {
		t.Fatalf("expected 1 line, got %d", len(tui.lines))
	}
	if !strings.Contains(tui.lines[0], "You:") {
		t.Errorf("expected 'You:' in line, got: %s", tui.lines[0])
	}
	if !strings.Contains(tui.lines[0], "Hello") {
		t.Errorf("expected 'Hello' in line, got: %s", tui.lines[0])
	}
}

func TestRenderChatLines_AssistantMessage(t *testing.T) {
	tui := New()
	tui.chat = []Message{{Role: "assistant", Content: "I can help you."}}
	tui.renderChatLines()

	if len(tui.lines) != 1 {
		t.Fatalf("expected 1 line, got %d", len(tui.lines))
	}
	if !strings.Contains(tui.lines[0], "Zé:") {
		t.Errorf("expected 'Zé:' in line, got: %s", tui.lines[0])
	}
}

func TestRenderChatLines_ToolMessage(t *testing.T) {
	tui := New()
	tui.chat = []Message{{Role: "tool", Content: "[Tool Result]: ok"}}
	tui.renderChatLines()

	if len(tui.lines) != 1 {
		t.Fatalf("expected 1 line, got %d", len(tui.lines))
	}
	if !strings.Contains(tui.lines[0], "Tool:") {
		t.Errorf("expected 'Tool:' in line, got: %s", tui.lines[0])
	}
}

func TestRenderChatLines_MultipleMessages(t *testing.T) {
	tui := New()
	tui.chat = []Message{
		{Role: "user", Content: "Write code"},
		{Role: "assistant", Content: "TOOL_CALL:write_file{...}"},
		{Role: "tool", Content: "[Tool Result]: ok"},
		{Role: "assistant", Content: "Done!"},
	}
	tui.renderChatLines()

	if len(tui.lines) != 4 {
		t.Fatalf("expected 4 lines, got %d", len(tui.lines))
	}
	if !strings.Contains(tui.lines[0], "You:") {
		t.Error("first line should be 'You:'")
	}
	if !strings.Contains(tui.lines[1], "Zé:") {
		t.Error("second line should be 'Zé:'")
	}
	if !strings.Contains(tui.lines[2], "Tool:") {
		t.Error("third line should be 'Tool:'")
	}
	if !strings.Contains(tui.lines[3], "Zé:") {
		t.Error("fourth line should be 'Zé:'")
	}
}

func TestRenderChatLines_EmptyChat(t *testing.T) {
	tui := New()
	tui.renderChatLines()
	if len(tui.lines) != 0 {
		t.Errorf("expected 0 lines for empty chat, got %d", len(tui.lines))
	}
}

func TestVisibleLen_PureText(t *testing.T) {
	if got := visibleLen("Hello World"); got != 11 {
		t.Errorf("visibleLen('Hello World') = %d, want 11", got)
	}
}

func TestVisibleLen_WithANSI(t *testing.T) {
	prompt := "\033[1m\033[36mze\033[0m \033[36m>\033[0m "
	if got := visibleLen(prompt); got != 5 {
		t.Errorf("visibleLen('ze > ') with ANSI = %d, want 5", got)
	}
}

func TestVisibleLen_Empty(t *testing.T) {
	if got := visibleLen(""); got != 0 {
		t.Errorf("visibleLen('') = %d, want 0", got)
	}
}

func TestVisibleLen_OnlyANSI(t *testing.T) {
	onlyANSI := "\033[1m\033[36m\033[0m"
	if got := visibleLen(onlyANSI); got != 0 {
		t.Errorf("visibleLen(only ANSI) = %d, want 0", got)
	}
}
