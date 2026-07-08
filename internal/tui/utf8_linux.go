//go:build linux

package tui

import (
	"os"
	"syscall"
	"unsafe"
)

// EnsureUTF8Terminal attempts to set the terminal to UTF-8 mode if the locale suggests it.
func EnsureUTF8Terminal() {
	if !isUTF8Locale() {
		return
	}

	fd := os.Stdin.Fd()
	var termios syscall.Termios

	// Get current terminal state using ioctl TCGETS
	_, _, err := syscall.Syscall6(
		syscall.SYS_IOCTL,
		fd,
		uintptr(syscall.TCGETS),
		uintptr(unsafe.Pointer(&termios)),
		0, 0, 0,
	)
	if err != 0 {
		return
	}

	// Set the IUTF8 flag in Iflag
	termios.Iflag |= syscall.IUTF8

	// Set the terminal state using ioctl TCSETS
	_, _, _ = syscall.Syscall6(
		syscall.SYS_IOCTL,
		fd,
		uintptr(syscall.TCSETS),
		uintptr(unsafe.Pointer(&termios)),
		0, 0, 0,
	)
}
