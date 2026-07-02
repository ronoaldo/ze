// Package tui provides a minimal terminal UI using ANSI escape codes.
// No external dependencies — pure Go + ANSI.
package tui

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
)

const (
	// ANSI escape codes
	ansiClearScreen      = "\033[2J"
	ansiMoveCursorHome   = "\033[H"
	ansiHideCursor       = "\033[?25l"
	ansiShowCursor       = "\033[?25h"
	ansiClearLine        = "\033[K"
	ansiMoveCursorUp     = "\033[A"
	ansiMoveCursorDown   = "\033[B"
	ansiMoveCursorLeft   = "\033[D"
	ansiMoveCursorRight  = "\033[C"
	ansiMoveCursorTo     = "\033[%d;%dH"
	ansiEnableRawMode    = "\033[?1049h"
	ansiDisableRawMode   = "\033[?1049l"

	// Colors
	ansiReset   = "\033[0m"
	ansiBold    = "\033[1m"
	ansiGreen   = "\033[32m"
	ansiBlue    = "\033[34m"
	ansiYellow  = "\033[33m"
	ansiCyan    = "\033[36m"
	ansiDim     = "\033[2m"
)

// Message represents a single chat message.
type Message struct {
	Role    string // "user", "assistant", "system", "tool"
	Content string
}

// TUI is the terminal user interface.
type TUI struct {
	w      *os.File // output (stdout)
	r      *os.File // input (stdin)
	chat   []Message
	input  string
	cursor int    // cursor position within input
	rows   int    // current terminal rows
	cols   int    // current terminal cols
	lines  []string // rendered chat lines
	sigCh  chan os.Signal
}

// Run starts the TUI event loop.
func (t *TUI) Run(handler func(msg string) (string, error)) error {
	t.setupTerminal()
	defer t.restoreTerminal()

	t.draw()

	for {
		input, err := t.readLine()
		if err != nil {
			return err
		}

		if strings.TrimSpace(input) == "" {
			continue
		}

		// Show user message immediately — clear input line, append to chat
		t.chat = append(t.chat, Message{Role: "user", Content: input})
		t.draw()

		// Call handler (LLM) — input is "locked" during this call
		response, err := handler(input)
		if err != nil {
			response = fmt.Sprintf("Error: %v", err)
		}

		t.chat = append(t.chat, Message{Role: "assistant", Content: response})
		t.draw()
	}
}

// listenResize handles terminal resize signals.
func (t *TUI) listenResize() {
	for range t.sigCh {
		t.resize()
	}
}



// draw renders the full UI.
func (t *TUI) draw() {
	// Re-detect size on every draw (handles resize)
	t.resize()

	t.renderChatLines()

	// Clear screen and move to top
	fmt.Fprint(t.w, ansiClearScreen+ansiMoveCursorHome)

	// Header
	t.drawHeader()

	// Chat area — reserve 3 rows: header(1) + separator(1) + input(1)
	visibleRows := t.rows - 3
	start := 0
	if len(t.lines) > visibleRows {
		start = len(t.lines) - visibleRows
	}
	for i := start; i < len(t.lines); i++ {
		fmt.Fprintf(t.w, "%s\n", t.lines[i])
	}

	// Separator
	fmt.Fprint(t.w, strings.Repeat("─", t.cols)+"\n")

	// Input line (last row)
	inputRow := t.rows
	fmt.Fprintf(t.w, "\033[%d;1H", inputRow)
	fmt.Fprint(t.w, ansiClearLine)

	// Draw prompt + full input text
	prompt := "\033[1m\033[36mze\033[0m \033[36m>\033[0m "
	fmt.Fprint(t.w, prompt)
	fmt.Fprint(t.w, t.input)

	// Position cursor: visible prompt length + cursor position
	cursorCol := visibleLen(prompt) + t.cursor + 1
	if cursorCol > t.cols {
		cursorCol = t.cols
	}
	fmt.Fprintf(t.w, "\033[%d;%dH", inputRow, cursorCol)

	// Flush to ensure output reaches terminal immediately
	t.w.Sync()
}

// drawHeader renders the header.
func (t *TUI) drawHeader() {
	title := "\033[1m\033[36mZé Agent\033[0m"
	history := fmt.Sprintf("\033[2mchat: %d msgs\033[0m", len(t.chat))
	fmt.Fprintf(t.w, "%s  %s\n", title, history)
}

// renderChatLines converts chat messages to formatted text lines.
func (t *TUI) renderChatLines() {
	t.lines = []string{}
	for _, msg := range t.chat {
		switch msg.Role {
		case "user":
			t.lines = append(t.lines, fmt.Sprintf("%s\033[1mYou:\033[0m %s", ansiBlue, msg.Content))
		case "assistant":
			t.lines = append(t.lines, fmt.Sprintf("%s\033[1mZé:\033[0m %s", ansiGreen, msg.Content))
		case "system":
			t.lines = append(t.lines, fmt.Sprintf("%s\033[1mSystem:\033[0m %s", ansiYellow, msg.Content))
		case "tool":
			t.lines = append(t.lines, fmt.Sprintf("%s\033[1mTool:\033[0m %s", ansiCyan, msg.Content))
		default:
			t.lines = append(t.lines, fmt.Sprintf("%s\033[1m%s:\033[0m %s", ansiDim, msg.Role, msg.Content))
		}
	}
}

// readLine reads a line with raw mode and line editing.
func (t *TUI) readLine() (string, error) {
	t.input = ""
	t.cursor = 0

	oldState, err := t.enableRawMode()
	if err != nil {
		return "", err
	}
	defer t.disableRawMode(oldState)

	for {
		buf := make([]byte, 1)
		_, err := t.r.Read(buf)
		if err != nil {
			return "", err
		}

		switch buf[0] {
		case '\n', '\r': // Enter
			fmt.Fprint(t.w, "\n")
			return t.input, nil
		case '\x03': // Ctrl+C
			fmt.Fprint(t.w, "^C\n")
			return "", nil
		case '\x7f', '\x08': // Backspace / Ctrl+H
			if t.cursor > 0 {
				t.input = t.input[:t.cursor-1] + t.input[t.cursor:]
				t.cursor--
				t.redrawInput()
			}
		case '\x1b': // Escape sequence
			seq := make([]byte, 3)
			_, err := t.r.Read(seq)
			if err != nil {
				return "", err
			}
			if seq[0] == '[' {
				switch seq[1] {
				case 'C': // Right
					if t.cursor < len(t.input) {
						t.cursor++
						t.redrawInput()
					}
				case 'D': // Left
					if t.cursor > 0 {
						t.cursor--
						t.redrawInput()
					}
				}
			}
		default:
			if len(buf) == 1 && buf[0] >= 0x20 && buf[0] <= 0x7e {
				pos := t.cursor
				t.input = t.input[:pos] + string(buf) + t.input[pos:]
				t.cursor++
				t.redrawInput()
			}
		}
	}
}

// redrawInput re-renders only the input line.
func (t *TUI) redrawInput() {
	prompt := "\033[1m\033[36mze\033[0m \033[36m>\033[0m "
	inputRow := t.rows

	fmt.Fprintf(t.w, "\033[%d;1H", inputRow)
	fmt.Fprint(t.w, ansiClearLine)
	fmt.Fprint(t.w, prompt)
	fmt.Fprint(t.w, t.input)

	cursorCol := visibleLen(prompt) + t.cursor + 1
	if cursorCol > t.cols {
		cursorCol = t.cols
	}
	fmt.Fprintf(t.w, "\033[%d;%dH", inputRow, cursorCol)
}

// setupTerminal prepares the terminal.
func (t *TUI) setupTerminal() {
	fmt.Fprint(t.w, ansiEnableRawMode)
	t.resize()
}

// restoreTerminal restores the terminal.
func (t *TUI) restoreTerminal() {
	fmt.Fprint(t.w, ansiDisableRawMode)
	fmt.Fprint(t.w, ansiShowCursor)
	signal.Stop(t.sigCh)
}

// termSize returns terminal dimensions (platform-specific fallback).
func (t *TUI) termSize() (int, int) {
	// Fallback
	return 24, 80
}

// visibleLen returns the number of visible characters in a string,
// ignoring ANSI escape codes.
func visibleLen(s string) int {
	count := 0
	inEscape := false
	for _, r := range s {
		if r == '\033' {
			inEscape = true
			continue
		}
		if inEscape {
			if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') {
				inEscape = false
			}
			continue
		}
		count++
	}
	return count
}
