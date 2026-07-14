package tui

import (
	"bufio"
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/ronoaldo/ze/internal/agent"
)

func TestTUI_HeadlessPrompt(t *testing.T) {
	input := "hello\n"
	r := strings.NewReader(input)
	w := &bytes.Buffer{}

	tui := &TUI{
		w:          w,
		r:          r,
		reader:     bufio.NewReader(r),
		palette:    NoColorPalette(),
		isHeadless: true,
	}

	handler := func(msg string) (string, agent.AgentStats, error) {
		return "", agent.AgentStats{}, io.EOF
	}

	err := tui.Run(handler, nil)
	if err != nil && err != io.EOF {
		t.Fatalf("expected EOF, got %v", err)
	}

	output := w.String()
	expectedPrefix := "prompt > "
	if !strings.HasPrefix(output, expectedPrefix) {
		t.Errorf("expected output to start with %q, got %q", expectedPrefix, output)
	}
}

func TestTUI_HeadlessNoColors(t *testing.T) {
	p := NoColorPalette()
	if p.Reset != "" || p.Bold != "" || p.Dim != "" || p.Cyan != "" || p.Green != "" || p.Red != "" || p.Yellow != "" {
		t.Errorf("NoColorPalette should have empty strings for all color fields, got: %+v", p)
	}
}

func TestTUI_HeadlessEcho(t *testing.T) {
	input := "hello\n"
	r := strings.NewReader(input)
	w := &bytes.Buffer{}

	tui := &TUI{
		w:          w,
		r:          r,
		reader:     bufio.NewReader(r),
		palette:    NoColorPalette(),
		isHeadless: true,
	}

	handler := func(msg string) (string, agent.AgentStats, error) {
		return "response", agent.AgentStats{}, io.EOF
	}

	err := tui.Run(handler, nil)
	if err != nil && err != io.EOF {
		t.Fatalf("expected EOF, got %v", err)
	}

	output := w.String()
	expectedEcho := "prompt > hello"
	if !strings.Contains(output, expectedEcho) {
		t.Errorf("expected output to contain %q, got %q", expectedEcho, output)
	}
}

func TestTUI_HeadlessEOF(t *testing.T) {
	r := strings.NewReader("") // Immediate EOF
	w := &bytes.Buffer{}

	tui := &TUI{
		w:          w,
		r:          r,
		reader:     bufio.NewReader(r),
		palette:    NoColorPalette(),
		isHeadless: true,
	}

	handler := func(msg string) (string, agent.AgentStats, error) {
		return "", agent.AgentStats{}, nil
	}

	err := tui.Run(handler, nil)
	if err != io.EOF {
		t.Errorf("expected EOF, got %v", err)
	}
}
