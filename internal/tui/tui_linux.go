//go:build linux

package tui

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"unsafe"
)

// New creates a new TUI instance (Linux).
func New() *TUI {
	t := &TUI{
		w:     os.Stdout,
		r:     os.Stdin,
		chat:  []Message{},
		sigCh: make(chan os.Signal, 1),
	}
	t.resize()
	signal.Notify(t.sigCh, syscall.SIGWINCH)
	go t.listenResize()
	return t
}

// resize updates terminal dimensions.
func (t *TUI) resize() {
	var size struct {
		rows uint16
		cols uint16
		_    uint16
		_    uint16
	}
	if _, _, err := syscall.Syscall6(syscall.SYS_IOCTL, t.w.Fd(), uintptr(syscall.TIOCGWINSZ), uintptr(unsafe.Pointer(&size)), 0, 0, 0); err == 0 {
		t.rows = int(size.rows)
		t.cols = int(size.cols)
		return
	}
	// Fallback
	t.rows = 24
	t.cols = 80
}

// enableRawMode puts stdin in raw mode.
func (t *TUI) enableRawMode() (*syscall.Termios, error) {
	var oldState syscall.Termios
	if _, _, err := syscall.Syscall(syscall.SYS_IOCTL, t.r.Fd(), uintptr(syscall.TCGETS), uintptr(unsafe.Pointer(&oldState))); err != 0 {
		return nil, fmt.Errorf("failed to get terminal state: %w", err)
	}

	newState := oldState
	newState.Lflag &^= syscall.ECHO | syscall.ICANON
	newState.Iflag &^= syscall.ICRNL
	newState.Cc[syscall.VMIN] = 1
	newState.Cc[syscall.VTIME] = 0

	if _, _, err := syscall.Syscall(syscall.SYS_IOCTL, t.r.Fd(), uintptr(syscall.TCSETS), uintptr(unsafe.Pointer(&newState))); err != 0 {
		return nil, fmt.Errorf("failed to set terminal state: %w", err)
	}

	return &oldState, nil
}

// disableRawMode restores the terminal.
func (t *TUI) disableRawMode(oldState *syscall.Termios) {
	syscall.Syscall(syscall.SYS_IOCTL, t.r.Fd(), uintptr(syscall.TCSETS), uintptr(unsafe.Pointer(oldState)))
}
