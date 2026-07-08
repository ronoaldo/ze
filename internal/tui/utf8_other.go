//go:build !linux

package tui

// EnsureUTF8Terminal is a no-op on non-Linux platforms.
func EnsureUTF8Terminal() {
}
