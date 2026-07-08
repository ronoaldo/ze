package tui

import "testing"

func TestNew(t *testing.T) {
	tui := New()
	if tui == nil {
		t.Fatal("New() returned nil")
	}
}
