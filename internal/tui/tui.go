package tui

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"github.com/ronoaldo/ze/internal/agent"
)

// winsize represents the terminal window size.
type winsize struct {
	Row    uint16
	Col    uint16
	Xpixel uint16
	Ypixel uint16
}

// TUI is the terminal user interface.
type TUI struct {
	w       io.Writer
	r       io.Reader
	reader  *bufio.Reader
	verbose bool
}

func isUTF8Locale() bool {
	lang := strings.ToUpper(os.Getenv("LANG"))
	lcAll := strings.ToUpper(os.Getenv("LC_ALL"))
	lcCtype := strings.ToUpper(os.Getenv("LC_CTYPE"))
	return strings.Contains(lang, "UTF-8") || strings.Contains(lcAll, "UTF-8") || strings.Contains(lcCtype, "UTF-8") ||
		strings.Contains(lang, "UTF8") || strings.Contains(lcAll, "UTF8") || strings.Contains(lcCtype, "UTF8")
}

// Run starts the TUI event loop.
func (t *TUI) Run(handler func(msg string) (string, agent.AgentStats, error)) error {
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
		response, stats, err := handler(input)
		if err != nil {
			return err
		}

		// Display response
		if response != "" {
			fmt.Fprintf(t.w, "\n%s\n\n", response)
			t.ReportStats(stats)
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

// summarizeArgs creates a human-readable summary of tool arguments.
func (t *TUI) summarizeArgs(toolName string, argsJSON string) string {
	var args map[string]interface{}
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return argsJSON
	}

	if path, ok := args["path"].(string); ok {
		return path
	}

	if toolName == "go_doc" {
		if pkg, ok := args["package"].(string); ok {
			return pkg
		}
	}

	return "{}"
}

// summarizeResult creates a human-readable summary of tool results.
func (t *TUI) summarizeResult(toolName string, result string) string {
	switch toolName {
	case "read_file", "edit_file":
		return fmt.Sprintf("%d bytes", len(result))
	case "list_files":
		lines := strings.Count(result, "\n")
		if result != "" && !strings.HasSuffix(result, "\n") {
			lines++
		}
		// The list_files implementation uses "- filename\n"
		// so count lines is a good proxy for items.
		// If result is empty, it's 0.
		if result == "" {
			return "0 items"
		}
		return fmt.Sprintf("%d items", lines)
	case "write_file":
		return "Success"
	case "go_doc":
		return "Success"
	default:
		return "Success"
	}
}

// ReportToolCall prints a formatted tool call message.
func (t *TUI) ReportToolCall(toolName string, args string) {
	summary := t.summarizeArgs(toolName, args)
	fmt.Fprintf(t.w, "* \033[1m\033[36m%s\033[0m('%s')", toolName, summary)
}

// ReportToolCallVerbose prints a full tool call message in dimmed color.
func (t *TUI) ReportToolCallVerbose(toolName string, args string) {
	fmt.Fprintf(t.w, "* \033[1m\033[36m[TOOL CALL] %s\033[0m: \033[2m%s\033[0m\n", toolName, args)
}

// ReportToolResult prints a formatted tool result message.
func (t *TUI) ReportToolResult(toolName string, result string, err error) {
	if err != nil {
		fmt.Fprintf(t.w, " | \033[1m\033[31m[ERROR] %v\033[0m\n", err)
		return
	}

	if t.verbose {
		fmt.Fprintf(t.w, " | \033[1m\033[32m[TOOL RESULT] %s\033[0m: \033[2m%s\033[0m\n", toolName, result)
	} else {
		summary := t.summarizeResult(toolName, result)
		fmt.Fprintf(t.w, " [\033[1m\033[32m%s\033[0m]\n", summary)
	}
}

// ReportStats displays performance statistics with a visual delimiter.
func (t *TUI) ReportStats(stats agent.AgentStats) {
	width := t.getTerminalWidth()
	if width < 20 {
		width = 40
	}
	
	line := fmt.Sprintf("Stats: %v | Tokens: %d | Speed: %.2f t/s", 
		stats.Duration.Round(time.Millisecond), 
		stats.TotalTokens, 
		stats.TokensPerSec,
	)
	
	// Pad the line with spaces to reach 'width'
	padding := width - len(line)
	if padding > 0 {
		line += strings.Repeat(" ", padding)
	} else if padding < 0 {
		line = line[:width]
	}

	// Print line with background color
	fmt.Fprintf(t.w, "\033[44m%s\033[0m\n", line)
}

// New creates a new TUI instance.
func New(verbose bool) *TUI {
	EnsureUTF8Terminal()
	return &TUI{
		w:       os.Stdout,
		r:       os.Stdin,
		reader:  bufio.NewReader(os.Stdin),
		verbose: verbose,
	}
}

func (t *TUI) getTerminalWidth() int {
	var ws winsize
	_, _, err := syscall.Syscall(syscall.SYS_IOCTL, uintptr(os.Stdout.Fd()), syscall.TIOCGWINSZ, uintptr(unsafe.Pointer(&ws)))
	if err != 0 {
		return 80
	}
	return int(ws.Col)
}
