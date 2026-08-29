package tui

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/ronoaldo/ze/internal/agent"
	"github.com/ronoaldo/ze/internal/tools"
)

// ErrInterrupt é retornado quando o usuário pressiona Ctrl+C.
var ErrInterrupt = errors.New("user interrupted")

// TUI is the terminal user interface.
type TUI struct {
	w             io.Writer
	r             io.Reader
	reader        *bufio.Reader
	term          terminal
	originalState any
	verbose       bool
	showThinking  bool
	palette       Palette
	rng           *rand.Rand
	isHeadless    bool
	messagePrefix string
	history       *HistoryManager
}

func isUTF8Locale() bool {
	lang := strings.ToUpper(os.Getenv("LANG"))
	lcAll := strings.ToUpper(os.Getenv("LC_ALL"))
	lcCtype := strings.ToUpper(os.Getenv("LC_CTYPE"))
	return strings.Contains(lang, "UTF-8") || strings.Contains(lcAll, "UTF-8") ||
		strings.Contains(lcCtype, "UTF-8") ||
		strings.Contains(lang, "UTF8") ||
		strings.Contains(lcAll, "UTF8") ||
		strings.Contains(lcCtype, "UTF8") ||
		strings.Contains(lang, "UTF8") ||
		strings.Contains(lcAll, "UTF8")
}

func (t *TUI) Run(ctx context.Context, handler func(ctx context.Context, msg string) (string, agent.AgentStats, error), isMultiline func() bool) error {
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "\n[Panic] Recuperando terminal e encerrando: %v\n", r)
		}
	}()

	for {
		// Check context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Print prompt
		if !t.isHeadless {
			fmt.Fprint(t.w, "\r\x1b[K")
			fmt.Fprintf(t.w, "%s%s%s%s> ", t.palette.Bold, t.palette.Cyan, "ze", t.palette.Reset)
		}

		// Read input
		input, err := t.readInput()
		if err != nil {
			return err
		}

		// Echo input if headless
		if t.isHeadless {
			fmt.Fprintf(t.w, "prompt > %s\n", input)
		}

		// Call handler (LLM)
		response, stats, err := handler(ctx, input)
		if err != nil {
			if errors.Is(err, ErrSkipLine) {
				continue
			}
			if errors.Is(err, ErrInterrupt) {
				return err
			}
			return err
		}

		// If it was a normal single-line input (not multiline and not empty), add to history
		if isMultiline == nil || !isMultiline() {
			if input != "" {
				t.history.Add(input)
			}
		}

		// Display response
		if response != "" {
			if strings.HasPrefix(response, "* Multiline input enabled") {
				fmt.Fprintf(t.w, "%s%s%s\n", t.palette.Dim, response, t.palette.Reset)
			} else {
				fmt.Fprintf(t.w, "\n%s\n\n", RenderMarkdown(response, t.markdownStyle()))
				// Only report stats if they are not empty
				if stats.TotalTokens > 0 || stats.Duration > 0 {
					t.ReportStats(stats)
				}
			}
		}
	}
}

// readInput reads a line from stdin byte by byte, handling backspace and enter.
func (t *TUI) readInput() (string, error) {
	if t.isHeadless {
		return t.readLine()
	}

	buffer := NewInputBuffer()

	for {
		key, err := ReadKey(t.reader)
		if err != nil {
			return "", err
		}

		width, _ := t.term.getTerminalSize()
		if width <= 0 {
			width = 80
		}

		switch key {
		case KeyEnter:
			fmt.Fprint(t.w, "\r\n")
			return buffer.String(), nil
		case KeyBackspace:
			buffer.Backspace()
		case KeyDelete:
			buffer.Delete()
		case KeyLeft:
			buffer.MoveLeft()
		case KeyRight:
			buffer.MoveRight()
		case KeyHome:
			buffer.MoveHome()
		case KeyEnd:
			buffer.MoveEnd()
		case KeyUp:
			if cmd, ok := t.history.Prev(); ok {
				buffer.SetContent(cmd)
			}
		case KeyDown:
			if cmd, ok := t.history.Next(); ok {
				buffer.SetContent(cmd)
			} else {
				buffer.SetContent("")
			}
		case KeyCtrlC:
			if !t.isHeadless {
				buffer.Clear()
				fmt.Fprint(t.w, "\r\x1b[K")
				return "", ErrInterrupt
			}
			buffer.Clear()
			return "", nil
		case KeyCtrlL:
			fmt.Fprintf(t.w, "\x1b[2J\x1b[H")
		default:
			if key >= Key(rune(0x10)) { // It's a rune
				buffer.Insert(rune(key))
			}
		}

		// Re-render the line
		contentLen := buffer.Len()
		totalLen := 7 + contentLen

		// Calculate current vertical position
		currentLinesOccupied := (totalLen - 1) / width

		// 1. Move up to the first line of the buffer if we are on a wrapped line
		if currentLinesOccupied > 0 {
			fmt.Fprintf(t.w, "\x1b[%dA", currentLinesOccupied)
			// Clear all lines from the first line to the last line of the buffer
			for i := 0; i <= currentLinesOccupied; i++ {
				fmt.Fprint(t.w, "\x1b[2K\r") // Clear current line
				if i < currentLinesOccupied {
					fmt.Fprint(t.w, "\x1b[B") // Move down
				}
			}
			// Move back to line 1
			fmt.Fprint(t.w, "\x1b[A")
		} else {
			// Single line
			fmt.Fprint(t.w, "\r\x1b[K")
		}

		// 2. Print prompt if not headless
		if !t.isHeadless {
			fmt.Fprintf(t.w, "%s%s%s%s> ", t.palette.Bold, t.palette.Cyan, "ze", t.palette.Reset)
		}

		// 3. Print buffer
		fmt.Fprintf(t.w, "%s", buffer.String())

		// 4. Move cursor back to cursorPos relative to the buffer start
		if buffer.CursorPosition() < buffer.Len() {
			fmt.Fprintf(t.w, "\x1b[%dD", buffer.Len()-buffer.CursorPosition())
		}
	}
}

// readLine reads a line from stdin.
func (t *TUI) readLine() (string, error) {
	line, err := t.reader.ReadString('\n')
	if err != nil {
		if err == io.EOF {
			return "", io.EOF
		}
		return "", err
	}
	return strings.TrimRight(line, "\r\n"), nil
}

// toStringSlice convert an interface slice to a string slice.
func toStringSlice(s []interface{}) []string {
	res := make([]string, len(s))
	for i, v := range s {
		res[i] = fmt.Sprintf("%v", v)
	}
	return res
}

func (t *TUI) ReportToolExecution(toolName string, summary string, res tools.ToolResult, err error) {
	header := fmt.Sprintf("%s%s%s%s(%s)",
		t.palette.Bold, t.palette.Cyan, toolName, t.palette.Reset, summary)

	if err != nil {
		fmt.Fprintf(t.w, "* %s %s[ERROR] %s%s\n", header, t.palette.Red, err.Error(), t.palette.Reset)
		return
	}

	if res.Summary != "" {
		fmt.Fprintf(t.w, "* %s %s%s%s\n", header, t.palette.Green, res.Summary, t.palette.Reset)
	} else {
		fmt.Fprintf(t.w, "* %s\n", header)
	}

	// Linha de Detalhe (Opcional/Esmaecida)
	if t.verbose || res.RequiresFullOutput {
		if res.FullResult != "" {
			fmt.Fprintf(t.w, "  %s%s%s\n", t.palette.Dim, res.FullResult, t.palette.Reset)
		}
	}
}

func (t *TUI) ReportStatus(stats agent.AgentStats) {
	status := stats.Status
	if status == "" {
		status = "OK"
	}

	line := fmt.Sprintf("Status: %s | %dt (In: %d, Out: %d) | %.0f t/s (%.0f t/s prefill)",
		status,
		stats.TotalTokens,
		stats.PromptTokens,
		stats.CompTokens,
		stats.CompPerSec,
		stats.PromptPerSec,
	)

	// Print line with dimmed color
	fmt.Fprintf(t.w, "%s%s%s%s\n", t.palette.Dim, t.messagePrefix, line, t.palette.Reset)
}

// ReportStats displays performance statistics with a visual delimiter.
func (t *TUI) ReportStats(stats agent.AgentStats) {
	t.ReportStatus(stats)
}

func (t *TUI) ReportReasoning(content string, tokens int) {
	if t.showThinking && content != "" {
		fmt.Fprintf(t.w, "\n%s%s%s\n", t.palette.Dim, content, t.palette.Reset)
	}
}

func (t *TUI) IsHeadless() bool {
	return t.isHeadless
}

// NewTestTUI creates a TUI instance for testing purposes.
func NewTestTUI(r io.Reader, w io.Writer) *TUI {
	return &TUI{
		r:             r,
		w:             w,
		reader:        bufio.NewReader(r),
		term:          newTerminal(),
		palette:       NoColorPalette(),
		rng:           rand.New(rand.NewSource(1)),
		isHeadless:    false,
		messagePrefix: "* ",
		history:       NewHistoryManager(100),
	}
}

// New creates a new TUI instance.
func New(verbose bool, showThinking bool, noColor bool) *TUI {
	EnsureUTF8Terminal()

	isTTY := false
	if stat, err := os.Stdin.Stat(); err == nil {
		if stat.Mode()&os.ModeCharDevice != 0 {
			isTTY = true
		}
	}

	palette := DefaultPalette()
	if noColor || !isTTY {
		palette = NoColorPalette()
	}

	return &TUI{
		w:             os.Stdout,
		r:             os.Stdin,
		reader:        bufio.NewReader(os.Stdin),
		term:          newTerminal(),
		verbose:       verbose,
		showThinking:  showThinking,
		palette:       palette,
		rng:           rand.New(rand.NewSource(time.Now().UnixNano())),
		isHeadless:    !isTTY,
		messagePrefix: "* ",
		history:       NewHistoryManager(100),
	}
}

func (t *TUI) markdownStyle() Style {
	return Style{
		Bold:      t.palette.Bold,
		Italic:    t.palette.Italic,
		Underline: t.palette.Underline,
		Reset:     t.palette.Reset,
	}
}

func (t *TUI) EnableRawMode() error {
	if t.isHeadless {
		return nil
	}
	state, err := t.term.enableRawMode()
	if err != nil {
		return err
	}
	t.originalState = state
	return nil
}

func (t *TUI) DisableRawMode() error {
	if t.originalState == nil {
		return nil
	}
	fmt.Fprintf(os.Stderr, "\n[DEBUG-TUI] DisableRawMode() called. originalState is NOT nil\n")
	err := t.term.disableRawMode(t.originalState)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[DEBUG-TUI] Error in term.disableRawMode: %v\n", err)
	} else {
		fmt.Fprintf(os.Stderr, "[DEBUG-TUI] term.disableRawMode() returned nil\n")
	}
	t.originalState = nil
	return err
}
