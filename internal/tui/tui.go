package tui

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"github.com/ronoaldo/ze/internal/agent"
)

var (
	reInserts = regexp.MustCompile(`(\d+)\s+insertions?\(?\+?\)?`)
	reDeletes = regexp.MustCompile(`(\d+)\s+deletions?\(?\-?\)?`)
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
	w               io.Writer
	r               io.Reader
	reader          *bufio.Reader
	verbose         bool
	showThinking    bool
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
		fmt.Fprint(t.w, "\x1b[1m\x1b[36mze\x1b[0m \x1b[36m>\x1b[0m ")

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
	
	// Since the Agent now ensures Arguments are cleaned during unmarshaling,
	// a single unmarshal attempt should be sufficient.
	err := json.Unmarshal([]byte(argsJSON), &args)
	if err != nil {
		return argsJSON
	}

	// Extract path for most file-related tools
	if path, ok := args["path"].(string); ok {
		return path
	}

	// Specific handling for other tools
	switch toolName {
	case "go_doc":
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
		if result == "" {
			return "0 items"
		}
		return fmt.Sprintf("%d items", lines)
	case "write_file", "go_doc":
		return "Success"
	case "diff":
		startMarker := "--- GIT STATS ---"
		endMarker := "--- GIT DIFF"
		
		startIdx := strings.Index(result, startMarker)
		if startIdx == -1 {
			return "Success (No stats)"
		}
		
		endIdx := strings.Index(result[startIdx:], endMarker)
		var statsPart string
		if endIdx == -1 {
			statsPart = result[startIdx:]
		} else {
			statsPart = result[startIdx : startIdx+endIdx]
		}

		insMatch := reInserts.FindStringSubmatch(statsPart)
		delMatch := reDeletes.FindStringSubmatch(statsPart)

		ins := "0"
		if len(insMatch) > 1 {
			ins = insMatch[1]
		}
		del := "0"
		if len(delMatch) > 1 {
			del = delMatch[1]
		}

		return fmt.Sprintf("+%s, -%s", ins, del)
	default:
		return "Success"
	}
}

// ReportToolCall prints a formatted tool call message.
func (t *TUI) ReportToolCall(toolName string, args string) {
	summary := t.summarizeArgs(toolName, args)
	fmt.Fprintf(t.w, "* \x1b[1m\x1b[36m%s\x1b[0m('%s')", toolName, summary)
}

// ReportToolCallVerbose prints a full tool call message in dimmed color.
func (t *TUI) ReportToolCallVerbose(toolName string, args string) {
	fmt.Fprintf(t.w, "* \x1b[1m\x1b[36m[TOOL CALL] %s\x1b[0m: \x1b[2m%s\x1b[0m\n", toolName, args)
}

// ReportToolResult prints a formatted tool result message.
func (t *TUI) ReportToolResult(toolName string, result string, err error) {
	if err != nil {
		fmt.Fprintf(t.w, " | \x1b[1m\x1b[31m[ERROR] %v\x1b[0m\n", err)
		return
	}

	if t.verbose {
		fmt.Fprintf(t.w, " | \x1b[1m\x1b[32m[TOOL RESULT] %s\x1b[0m: \x1b[2m%s\x1b[0m\n", toolName, result)
	} else {
		summary := t.summarizeResult(toolName, result)
		fmt.Fprintf(t.w, " [\x1b[1m\x1b[32m%s\x1b[0m]\n", summary)
	}
}

// ReportStats displays performance statistics with a visual delimiter.
func (t *TUI) ReportStats(stats agent.AgentStats) {
	width := t.getTerminalWidth()
	if width < 20 {
		width = 40
	}
	
	line := fmt.Sprintf("Stats: %v | P: %d C: %d | Speed: %.2f t/s", 
		stats.Duration.Round(time.Millisecond), 
		stats.PromptTokens,
		stats.CompTokens,
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
	fmt.Fprintf(t.w, "\x1b[44m%s\x1b[0m\n", line)
}

// New creates a new TUI instance.
func New(verbose bool, showThinking bool) *TUI {
	EnsureUTF8Terminal()
	return &TUI{
		w:               os.Stdout,
		r:               os.Stdin,
		reader:          bufio.NewReader(os.Stdin),
		verbose:         verbose,
		showThinking:    showThinking,
	}
}

// ReportReasoning prints a summary of the reasoning process and the full content if requested.
func (t *TUI) ReportReasoning(content string) {
	fmt.Fprintf(t.w, "\n\x1b[33m🧠 Pensou %d bytes de pura genialidade...\x1b[0m\n", len(content))
	if t.showThinking {
		fmt.Fprintf(t.w, "\x1b[2m%s\x1b[0m\n", content)
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
