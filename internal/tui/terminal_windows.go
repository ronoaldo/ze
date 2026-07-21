//go:build windows

package tui

import (
	"os"
	"syscall"
	"unsafe"
)

type coord struct {
	X int32
	Y int32
}

type consoleScreenBufferInfo struct {
	Size           coord
	CursorPosition coord
	Attributes     uint32
	Window         coord
	CursorCurrent  coord
}

var (
	kernel32           = syscall.NewLazyDLL("kernel32.dll")
	procGetConsoleSize = kernel32.NewProc("GetConsoleScreenBufferInfo")
)

func getTerminalSize() (int, int) {
	h := syscall.Handle(os.Stdout.Fd())
	var info consoleScreenBufferInfo
	ret, _, _ := procGetConsoleSize.Call(uintptr(h), uintptr(unsafe.Pointer(&info)))
	if ret == 0 {
		return 80, 24
	}
	return int(info.Size.X), int(info.Size.Y)
}
