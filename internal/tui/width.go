package tui

import (
	"regexp"
)

// ansiRegex matches ANSI escape sequences used for colors and styles.
var ansiRegex = regexp.MustCompile(`\x1b\[[0-9;]*m`)

// VisibleWidth calculates the number of visible characters in a string,
// ignoring ANSI escape sequences and counting Unicode runes.
func VisibleWidth(s string) int {
	// 1. Remove ANSI escape sequences so they don't count towards width.
	plain := ansiRegex.ReplaceAllString(s, "")

	// 2. Count the number of runes.
	// This correctly handles multi-byte UTF-8 characters.
	// Note: For full support of "wide" characters (like some emojis),
	// one would need a library like 'go-runewidth', but rune count is
	// the standard next step for terminal applications.
	runes := []rune(plain)
	return len(runes)
}
