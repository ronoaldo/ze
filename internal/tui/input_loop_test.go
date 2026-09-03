package tui

import (
	"bytes"
	"strings"
	"testing"
)

func TestReadInput(t *testing.T) {
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

		// Check echo
		expectedEcho := "\r\x1b[Kze> h\r\x1b[Kze> he\r\x1b[Kze> hel\r\x1b[Kze> hell\r\x1b[Kze> hello\r\n"
		if w.String() != expectedEcho {
			t.Errorf("Expected echo %q, got %q", expectedEcho, w.String())
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
		// Check echo
		expectedEcho := "\r\x1b[Kze> h\r\x1b[Kze> he\r\x1b[Kze> hel\r\x1b[Kze> he\r\x1b[Kze> heo\r\n"
		if w.String() != expectedEcho {
			t.Errorf("Expected echo %q, got %q", expectedEcho, w.String())
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
		// Check echo
		expectedEcho := "\r\x1b[Kze> h\r\x1b[Kze> he\r\x1b[Kze> hel\r\x1b[Kze> he\r\x1b[Kze> heo\r\n"
		if w.String() != expectedEcho {
			t.Errorf("Expected echo %q, got %q", expectedEcho, w.String())
		}
	})
}
