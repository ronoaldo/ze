package tui

import (
	"testing"
)

func TestHistoryManager(t *testing.T) {
	h := NewHistoryManager(3)

	// Test Initial State
	if h.index != -1 {
		t.Errorf("expected initial index -1, got %d", h.index)
	}

	// Test Add
	h.Add("cmd1")
	h.Add("cmd2")
	h.Add("cmd2") // Duplicate consecutive
	if len(h.GetHistory()) != 2 {
		t.Errorf("expected history size 2, got %d", len(h.GetHistory()))
	}

	h.Add("cmd3")
	if len(h.GetHistory()) != 3 {
		t.Errorf("expected history size 3, got %d", len(h.GetHistory()))
	}

	// Current state: history=["cmd1", "cmd2", "cmd3"], index=-1 (new line)

	// Test Prev from new line (should go to last)
	val, ok := h.Prev()
	if !ok || val != "cmd3" {
		t.Errorf("Prev from new line: expected cmd3, ok=true, got %s, ok=%v", val, ok)
	}

	// Test Prev from cmd3 (index 2)
	val, ok = h.Prev()
	if !ok || val != "cmd2" {
		t.Errorf("Prev from cmd3: expected cmd2, ok=true, got %s, ok=%v", val, ok)
	}

	// Test Prev from cmd2 (index 1)
	val, ok = h.Prev()
	if !ok || val != "cmd1" {
		t.Errorf("Prev from cmd2: expected cmd1, ok=true, got %s, ok=%v", val, ok)
	}

	// Test Prev from cmd1 (index 0)
	val, ok = h.Prev()
	if ok {
		t.Errorf("Prev from cmd1: expected ok=false, got ok=%v", ok)
	}

	// Test Next from cmd1 (index 0)
	val, ok = h.Next()
	if !ok || val != "cmd2" {
		t.Errorf("Next from cmd1: expected cmd2, ok=true, got %s, ok=%v", val, ok)
	}

	// Test Next from cmd3 (index 2)
	h.index = 2
	val, ok = h.Next()
	if ok {
		t.Errorf("Next from cmd3: expected ok=false, got ok=%v", ok)
	} else if h.index != -1 {
		t.Errorf("Next from cmd3: expected index to be -1, got %d", h.index)
	}

	// Test Next from new line (-1)
	h.index = -1
	val, ok = h.Next()
	if ok {
		t.Errorf("Next from new line: expected ok=false, got ok=%v", ok)
	}

	// Test maxSize
	h.Add("cmd4")
	h.Add("cmd5")
	if len(h.GetHistory()) > 3 {
		t.Errorf("expected history size at most 3, got %d", len(h.GetHistory()))
	}

	// Test Clear
	h.Clear()
	if len(h.GetHistory()) != 0 || h.index != -1 {
		t.Errorf("Clear failed: history length %d, index %d", len(h.GetHistory()), h.index)
	}
}
