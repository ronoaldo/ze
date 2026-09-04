package tui

import (
	"bytes"
	"strings"
	"testing"
)

func TestReadInput_Headless(t *testing.T) {
	t.Run("Basic input", func(t *testing.T) {
		input := "hello\r\n"
		r := strings.NewReader(input)
		w := new(bytes.Buffer)
		tui := NewTestTUI(r, w)
		tui.isHeadless = true

		line, err := tui.readInput()
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if line != "hello" {
			t.Errorf("Expected 'hello', got %q", line)
		}
		if w.Len() > 0 {
			t.Errorf("Expected empty w in headless mode, got %q", w.String())
		}
	})
}

func TestReadInput_Interactive(t *testing.T) {
	t.Run("Basic input", func(t *testing.T) {
		input := "hello\r\n"
		r := strings.NewReader(input)
		w := new(bytes.Buffer)
		tui := NewTestTUI(r, w)
		tui.isHeadless = false

		line, err := tui.readInput()
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if line != "hello" {
			t.Errorf("Expected 'hello', got %q", line)
		}
		// We don't check the exact content of w because it's full of ANSI codes
		// and depends on how the terminal emulator (or mock) handles them.
		// But we check that it's not empty (it should at least have the prompt/echo).
		if w.Len() == 0 {
			t.Error("Expected w to contain some output (prompt/echo), but it was empty")
		}
	})

	t.Run("Backspace", func(t *testing.T) {
		// 'hel\bo\r\n' -> 'heo'
		input := "hel\bo\r\n"
		r := strings.NewReader(input)
		w := new(bytes.Buffer)
		tui := NewTestTUI(r, w)
		tui.isHeadless = false

		line, err := tui.readInput()
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if line != "heo" {
			t.Errorf("Expected 'heo', got %q", line)
		}
	})

	t.Run("DEL key", func(t *testing.T) {
		// 'hel\x7fo\r\n' -> 'heo'
		input := "hel\x7fo\r\n"
		r := strings.NewReader(input)
		w := new(bytes.Buffer)
		tui := NewTestTUI(r, w)
		tui.isHeadless = false

		line, err := tui.readInput()
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if line != "heo" {
			t.Errorf("Expected 'heo', got %q", line)
		}
	})
}
