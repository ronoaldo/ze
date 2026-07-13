package tui

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
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

// Palette defines the colors used in the TUI.
type Palette struct {
	Reset  string
	Bold   string
	Dim    string
	Cyan   string
	Green  string
	Red    string
	Yellow string
}

func DefaultPalette() Palette {
	return Palette{
		Reset:  "\x1b[0m",
		Bold:   "\x1b[1m",
		Dim:    "\x1b[2m",
		Cyan:   "\x1b[36m",
		Green:  "\x1b[32m",
		Red:    "\x1b[31m",
		Yellow: "\x1b[33m",
	}
}

func NoColorPalette() Palette {
	return Palette{
		Reset:  "",
		Bold:   "",
		Dim:    "",
		Cyan:   "",
		Green:  "",
		Red:    "",
		Yellow: "",
	}
}

// TUI is the terminal user interface.
type TUI struct {
	w               io.Writer
	r               io.Reader
	reader          *bufio.Reader
	verbose         bool
	showThinking    bool
	palette         Palette
	rng             *rand.Rand
}

func isUTF8Locale() bool {
	lang := strings.ToUpper(os.Getenv("LANG"))
	lcAll := strings.ToUpper(os.Getenv("LC_ALL"))
	lcCtype := strings.ToUpper(os.Getenv("LC_CTYPE"))
	return strings.Contains(lang, "UTF-8") || strings.Contains(lcAll, "UTF-8") || strings.Contains(lcCtype, "UTF-8") ||
		strings.Contains(lang, "UTF8") || strings.Contains(lcAll, "UTF8") || strings.Contains(lcCtype, "UTF8") ||
		strings.Contains(lang, "UTF8") || strings.Contains(lcAll, "UTF8")
}

// Run starts the TUI event loop.
func (t *TUI) Run(handler func(msg string) (string, agent.AgentStats, error), isMultiline func() bool) error {
	for {
		// Print prompt
		if isMultiline != nil && isMultiline() {
			fmt.Fprintf(t.w, "   %s%s%s ", t.palette.Cyan, ">", t.palette.Reset)
		} else {
			fmt.Fprintf(t.w, "%s%s%s%s %s%s%s ", t.palette.Bold, t.palette.Cyan, "ze", t.palette.Reset, t.palette.Cyan, ">", t.palette.Reset)
		}

		// Read input
		input, err := t.readLine()
		if err != nil {
			return err
		}

		// Call handler (LLM)
		response, stats, err := handler(input)
		if err != nil {
			if errors.Is(err, ErrSkipLine) {
				continue
			}
			return err
		}

		// Display response
		if response != "" {
			fmt.Fprintf(t.w, "\n%s\n\n", response)
			// Only report stats if they are not empty
			if stats.TotalTokens > 0 || stats.Duration > 0 {
				t.ReportStats(stats)
			}
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
	
	err := json.Unmarshal([]byte(argsJSON), &args)
	if err != nil {
		return argsJSON
	}

	if path, ok := args["path"].(string); ok {
		return path
	}

	switch toolName {
	case "go_doc":
		if pkg, ok := args["package"].(string); ok {
			return pkg
		}
	case "diff":
		return "."
	}

	return "{}"
}

// summarizeResult creates a human-readable summary of tool results.
func (t *TUI) summarizeResult(toolName string, result string) string {
	switch toolName {
	case "read_file", "edit_file":
		if toolName == "edit_file" {
			// Try to extract the summary from the end of the message, e.g., "[+10, -5]"
			start := strings.LastIndex(result, "[")
			end := strings.LastIndex(result, "]")
			if start != -1 && end != -1 && end > start {
				return result[start+1 : end]
			}
		}
		return fmt.Sprintf("%d bytes", len(result))
	case "write_file":
		re := regexp.MustCompile(`wrote (\d+) bytes`)
		match := re.FindStringSubmatch(result)
		if len(match) > 1 {
			return fmt.Sprintf("%s bytes", match[1])
		}
		return "Success"
	case "list_files":
		lines := strings.Count(result, "\n")
		if result != "" && !strings.HasSuffix(result, "\n") {
			lines++
		}
		if result == "" {
			return "0 items"
		}
		return fmt.Sprintf("%d items", lines)
	case "go_doc":
		return "Success"
	case "diff":
		if strings.HasPrefix(result, "SUMMARY: ") {
			lines := strings.SplitN(result, "\n", 2)
			summary := strings.TrimPrefix(lines[0], "SUMMARY: ")
			summary = strings.TrimSpace(summary)
			if summary == "" {
				return "no changes"
			}
			// Remove redundant prefix if it's in the summary line
			summary = strings.TrimPrefix(summary, "git_diff('.')")
			summary = strings.TrimPrefix(summary, "git_diff('.'),")
			return strings.TrimSpace(summary)
		}
		return "Success (No summary)"
	default:
		return "Success"
	}
}

// ReportToolCall prints a formatted tool call message.
func (t *TUI) ReportToolCall(toolName string, args string) {
	summary := t.summarizeArgs(toolName, args)
	fmt.Fprintf(t.w, "* %s%s%s%s('%s')", t.palette.Bold, t.palette.Cyan, toolName, t.palette.Reset, summary)
}

// ReportToolCallVerbose prints a full tool call message in dimmed color.
func (t *TUI) ReportToolCallVerbose(toolName string, args string) {
	fmt.Fprintf(t.w, "* %s%s[TOOL CALL] %s%s: %s%s%s\n", t.palette.Bold, t.palette.Cyan, toolName, t.palette.Reset, t.palette.Dim, args, t.palette.Reset)
}

// ReportToolResult prints a formatted tool result message.
func (t *TUI) ReportToolResult(toolName string, result string, err error) {
	if err != nil {
		fmt.Fprintf(t.w, " | %s%s[ERROR] %v%s\n", t.palette.Bold, t.palette.Red, err, t.palette.Reset)
		if result != "" {
			fmt.Fprintf(t.w, " | %s%s%s\n", t.palette.Dim, result, t.palette.Reset)
		}
		return
	}

	if t.verbose {
		fmt.Fprintf(t.w, " | %s%s[TOOL RESULT] %s%s: %s%s%s\n", t.palette.Bold, t.palette.Green, toolName, t.palette.Reset, t.palette.Dim, result, t.palette.Reset)
	} else {
		summary := t.summarizeResult(toolName, result)
		fmt.Fprintf(t.w, " [%s%s%s%s]\n", t.palette.Bold, t.palette.Green, summary, t.palette.Reset)
	}
}

// ReportStats displays performance statistics with a visual delimiter.
func (t *TUI) ReportStats(stats agent.AgentStats) {
	line := fmt.Sprintf("Stats: %v | P: %d C: %d | Speed: %.2f t/s", 
		stats.Duration.Round(time.Millisecond), 
		stats.PromptTokens,
		stats.CompTokens,
		stats.TokensPerSec,
	)
	
	// Print line with dimmed color (grayish/faded style)
	fmt.Fprintf(t.w, "%s%s%s\n", t.palette.Dim, line, t.palette.Reset)
}

// New creates a new TUI instance.
func New(verbose bool, showThinking bool, noColor bool) *TUI {
	EnsureUTF8Terminal()
	palette := DefaultPalette()
	if noColor {
		palette = NoColorPalette()
	}
	return &TUI{
		w:               os.Stdout,
		r:               os.Stdin,
		reader:          bufio.NewReader(os.Stdin),
		verbose:         verbose,
		showThinking:    showThinking,
		palette:         palette,
		rng:             rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// ReportReasoning prints a summary of the reasoning process and the full content if requested.
func (t *TUI) ReportReasoning(content string, tokens int) {
	terms := []string{
		"pura alucinação",
		"um surto de lógica",
		"um delírio de silício",
		"um sonho de robô",
		"uma epifania de bits",
		"um erro de sistema elegante",
		"uma conexão de Wi-Fi espiritual",
		"um sussurro de processador",
		"um caos de algoritmos",
		"uma sinapse de eletricidade",
		"um salto no escuro digital",
		"um insight de calculadora",
		"uma magia de código mal escrito",
		"um fluxo de dados caótico",
		"um enigma de bytes",
		"uma genialidade de baixo nível",
	}

	term := terms[t.rng.Intn(len(terms))]
	fmt.Fprintf(t.w, "* %sPensou %d tokens de %s...%s\n", t.palette.Yellow, tokens, term, t.palette.Reset)
	if t.showThinking {
		fmt.Fprintf(t.w, "%s%s%s\n", t.palette.Dim, content, t.palette.Reset)
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
