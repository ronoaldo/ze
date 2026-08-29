package tui

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// Key represents a detected key press from the terminal.
// Values 1-9 are reserved for special keys. Values > 9 are runes.
type Key int32

const (
	KeyNone Key = iota
	KeyEnter
	KeyBackspace
	KeyDelete
	KeyLeft
	KeyRight
	KeyHome
	KeyEnd
	KeyUp
	KeyDown
	KeyCtrlC
	KeyCtrlL
)

// ParseEscapeSequence reads from the provided reader and attempts to parse an ANSI escape sequence.
// It expects the reader to be positioned right after the ESC (\x1b) byte.
// It returns the identified Key or KeyNone if the sequence is unknown or incomplete.
func ParseEscapeSequence(r *bufio.Reader) (Key, error) {
	// The next byte should be '[' for CSI (Control Sequence Introducer)
	next, err := r.ReadByte()
	if err != nil {
		return KeyNone, err
	}

	if next != '[' {
		return KeyNone, fmt.Errorf("not a CSI sequence")
	}

	// Read the rest of the sequence until we hit a character that is not a number or ';'
	cmd := ""
	for {
		next, err := r.ReadByte()
		if err != nil {
			if err == io.EOF {
				break
			}
			return KeyNone, err
		}
		cmd += string(next)
		if (next < '0' || next > '9') && next != ';' {
			break
		}
	}

	switch strings.ToUpper(cmd) {
	case "A":
		return KeyUp, nil
	case "B":
		return KeyDown, nil
	case "C":
		return KeyRight, nil
	case "D":
		return KeyLeft, nil
	case "H":
		return KeyHome, nil
	case "1~":
		return KeyHome, nil
	case "F":
		return KeyHome, nil
	case "G":
		return KeyEnd, nil
	case "4~":
		return KeyEnd, nil
	case "L":
		return KeyEnd, nil
	case "3~":
		return KeyDelete, nil
	}

	return KeyNone, nil
}

// ReadKey reads a single key from the provided bufio.Reader, handling escape sequences and UTF-8.
func ReadKey(br *bufio.Reader) (Key, error) {
	// Peek at the first byte to see if it's an escape sequence
	first, err := br.Peek(1)
	if err != nil {
		return KeyNone, err
	}

	b := first[0]

	if b == 0x1b {
		// It's an escape sequence. Consume ESC and parse.
		_, _ = br.ReadByte()
		key, err := ParseEscapeSequence(br)
		if err != nil {
			return KeyNone, nil
		}
		return key, nil
	}

	// For everything else, try to read a full rune
	r, _, err := br.ReadRune()
	if err != nil {
		return KeyNone, err
	}

	switch r {
	case '\r', '\n':
		return KeyEnter, nil
	case '\b', 0x7f:
		return KeyBackspace, nil
	case 0x03:
		return KeyCtrlC, nil
	case 0x0c:
		return KeyCtrlL, nil
	default:
		if r >= 32 {
			return Key(int32(r)), nil
		}
		return KeyNone, nil
	}
}
