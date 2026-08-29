package tui

import "testing"

func TestSimple(t *testing.T) {
	b := NewInputBuffer()
	b.Insert('a')
	b.Insert('b')
	if b.String() != "ab" {
		t.Errorf("expected ab, got %s", b.String())
	}
}
