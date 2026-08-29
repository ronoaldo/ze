package tui

import (
	"testing"
)

func TestInputBuffer(t *testing.T) {
	t.Run("Insert and String", func(t *testing.T) {
		b := NewInputBuffer()
		if b.String() != "" {
			t.Errorf("expected empty string, got %q", b.String())
		}
		b.Insert('a')
		b.Insert('b')
		if b.String() != "ab" {
			t.Errorf("expected 'ab', got %q", b.String())
		}
	})

	t.Run("Backspace", func(t *testing.T) {
		b := NewInputBuffer()
		b.Insert('a')
		b.Insert('b')
		b.Backspace()
		if b.String() != "a" {
			t.Errorf("expected 'a', got %q", b.String())
		}
		b.Backspace()
		if b.String() != "" {
			t.Errorf("expected empty string, got %q", b.String())
		}
	})

	t.Run("Delete", func(t *testing.T) {
		b := NewInputBuffer()
		b.Insert('a')
		b.Insert('b')
		b.Insert('c')
		b.cursorPos = 1
		b.Delete()
		if b.String() != "ac" {
			t.Errorf("expected 'ac', got %q", b.String())
		}
		b.Delete()
		if b.String() != "a" {
			t.Errorf("expected 'a', got %q", b.String())
		}
	})

	t.Run("Navigation", func(t *testing.T) {
		b := NewInputBuffer()
		b.Insert('a')
		b.Insert('b')
		b.Insert('c')

		b.MoveLeft()
		if b.CursorPosition() != 2 {
			t.Errorf("expected cursor at 2, got %d", b.CursorPosition())
		}

		b.MoveRight()
		if b.CursorPosition() != 3 {
			t.Errorf("expected cursor at 3, got %d", b.CursorPosition())
		}

		b.MoveHome()
		if b.CursorPosition() != 0 {
			t.Errorf("expected cursor at 0, got %d", b.CursorPosition())
		}

		b.MoveEnd()
		if b.CursorPosition() != 3 {
			t.Errorf("expected cursor at 3, got %d", b.CursorPosition())
		}
	})

	t.Run("Clear", func(t *testing.T) {
		b := NewInputBuffer()
		b.Insert('a')
		b.Clear()
		if b.String() != "" || b.CursorPosition() != 0 {
			t.Errorf("expected empty buffer, got %q at %d", b.String(), b.CursorPosition())
		}
	})

	t.Run("SetContent", func(t *testing.T) {
		b := NewInputBuffer()
		b.SetContent("hello")
		if b.String() != "hello" || b.CursorPosition() != 5 {
			t.Errorf("expected 'hello' at 5, got %q at %d", b.String(), b.CursorPosition())
		}
	})
}
