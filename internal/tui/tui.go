// Package tui provides a minimal terminal UI that behaves like a shell.
package tui

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

// TUI is the terminal user interface.
type TUI struct {
	w      io.Writer
	r      io.Reader
	reader *bufio.Reader
}

func isUTF8Locale() bool {
	lang := strings.ToUpper(os.Getenv("LANG"))
	lcAll := strings.ToUpper(os.Getenv("LC_ALL"))
	lcCtype := strings.ToUpper(os.Getenv("LC_CTYPE"))
	return strings.Contains(lang, "UTF-8") || strings.Contains(lcAll, "UTF-8") || strings.Contains(lcCtype, "UTF-8") ||
		strings.Contains(lang, "UTF8") || strings.Contains(lcAll, "UTF8") || strings.Contains(lcCtype, "UTF8")
}

// Run starts the TUI event loop.
func (t *TUI) Run(handler func(msg string) (string, error)) error {
	for {
		// Print prompt
		fmt.Fprint(t.w, "\033[1m\033[36mze\033[0m \033[36m>\033[0m ")

		// Read input
		input, err := t.readLine()
		if err != nil {
			return err
		}

		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}

		// Call handler (LLM)
		response, err := handler(input)
		if err != nil {
			fmt.Fprintf(t.w, "Error: %v\n", err)
			continue
		}

		// Display response
		if response != "" {
			fmt.Fprintf(t.w, "%s\n", response)
		}
	}
}

// readLine reads a line from stdin.
func (t *TUI) readLine() (string, error) {
	line, err := t.reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimRight(line, "\r\n"), nil
}

// ReportToolCall prints a formatted tool call message.
func (t *TUI) ReportToolCall(toolName string, args string) {
	fmt.Fprintf(t.w, "\033[1m\033[36m[TOOL CALL] %s\033[0m: %s\n", toolName, args)
}

// ReportToolResult prints a formatted tool result message.
func (t *TUI) ReportToolResult(toolName string, result string, err error) {
	if err != nil {
		fmt.Fprintf(t.w, "\033[1m\033[31m[TOOL RESULT] %s\033[0m: ERROR: %v\n", toolName, err)
	} else {
		fmt.Fprintf(t.w, "\033[1m\033[32m[TOOL RESULT] %s\033[0m: %s\n", toolName, result)
	}
}

// New creates a new TUI instance.
func New() *TUI {
	EnsureUTF8Terminal()
	return &TUI{
		w:      os.Stdout,
		r:      os.Stdin,
		reader: bufio.NewReader(os.Stdin),
	}
}
