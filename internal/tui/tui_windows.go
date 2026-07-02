//go:build windows

package tui

import (
	"fmt"
	"os"
)

// New creates a new TUI instance (Windows).
func New() *TUI {
	t := &TUI{
		w:     os.Stdout,
		r:     os.Stdin,
		chat:  []Message{},
		sigCh: make(chan os.Signal, 1),
	}
	t.resize()
	// SIGWINCH not supported on Windows — just start the listener
	go t.listenResize()
	return t
}

// resize updates terminal dimensions (Windows fallback).
func (t *TUI) resize() {
	t.rows = 24
	t.cols = 80
}

// enableRawMode is not supported on Windows.
func (t *TUI) enableRawMode() (*os.File, error) {
	return nil, fmt.Errorf("raw terminal mode is not supported on Windows — run in WSL or Cygwin")
}

// disableRawMode is a no-op on Windows.
func (t *TUI) disableRawMode(*os.File) {}
