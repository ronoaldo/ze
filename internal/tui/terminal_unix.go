//go:build unix

package tui

import (
	"os"
	"syscall"
	"unsafe"
)

const (
	tcgetattr = 0x5401
	tcsetattr = 0x5402
)

type platformTerminal struct {
	originalTermios *syscall.Termios
}

func (p *platformTerminal) getTerminalSize() (int, int) {
	var ws winsize
	_, _, err := syscall.Syscall(syscall.SYS_IOCTL, uintptr(os.Stdout.Fd()), syscall.TIOCGWINSZ, uintptr(unsafe.Pointer(&ws)))
	if err != 0 {
		return 80, 24
	}
	return int(ws.Col), int(ws.Row)
}

func (p *platformTerminal) enableRawMode() (any, error) {
	fd := int(os.Stdin.Fd())
	t := new(syscall.Termios)

	// Get current attributes
	if _, _, err := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), tcgetattr, uintptr(unsafe.Pointer(t))); err != 0 {
		return nil, err
	}

	// Store original to restore later
	p.originalTermios = t

	// Prepare new attributes in a copy to avoid modifying the original
	raw := *t
	raw.Lflag &^= syscall.ECHO | syscall.ICANON | syscall.ISIG | syscall.IEXTEN
	raw.Iflag &^= syscall.IXON | syscall.ICRNL

	// Set new attributes
	if _, _, err := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), tcsetattr, uintptr(unsafe.Pointer(&raw))); err != 0 {
		return nil, err
	}

	return p.originalTermios, nil
}

func (p *platformTerminal) disableRawMode(original any) error {
	if original == nil {
		return nil
	}
	orig, ok := original.(*syscall.Termios)
	if !ok {
		return syscall.Errno(0)
	}

	fd := int(os.Stdin.Fd())
	if _, _, err := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), tcsetattr, uintptr(unsafe.Pointer(orig))); err != 0 {
		return err
	}
	return nil
}

type winsize struct {
	Row    uint16
	Col    uint16
	Xpixel uint16
	Ypixel uint16
}
