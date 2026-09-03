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
	w                 io.Writer
	r                 io.Reader
	reader            *bufio.Reader
	term              terminal
	originalState     any
	verbose           bool
	showThinking      bool
	palette           Palette
	rng               *rand.Rand
	isHeadless        bool
	messagePrefix     string
	history           *HistoryManager
	lastRenderedLines int
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

// readInput reads a line from stdin, byte by byte, handling backspace, enter, and line wrapping.
func (t *TUI) readInput() (string, error) {
	if t.isHeadless {
		return t.readLine()
	}

	buffer := NewInputBuffer()
	prompt := ""
	if !t.isHeadless {
		prompt = fmt.Sprintf("%s%s%s%s> ", t.palette.Bold, t.palette.Cyan, "ze", t.palette.Reset)
	}

	// Since Run() printed the prompt before calling readInput,
	// we start with 1 line already rendered.
	t.lastRenderedLines = 1

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
			fmt.Fprint(t.w, prompt)
			t.lastRenderedLines = 1
		default:
			if key >= Key(rune(0x10)) { // It's a rune
				buffer.Insert(rune(key))
			}
		}

		// --- RE-RENDER LOGIC (BLOCK RENDERING) ---
		content := buffer.String()
		lines := t.calculateRenderedLines(prompt, content, width)

		// 1. Clear the previous render block
		if t.lastRenderedLines > 0 {
			// Move up to the top of the block
			if t.lastRenderedLines > 1 {
				fmt.Fprintf(t.w, "\x1b[%dA", t.lastRenderedLines-1)
			}

			// Clear each line in the block
			for i := 0; i < t.lastRenderedLines; i++ {
				fmt.Fprint(t.w, "\r\x1b[K") // Clear current line and move to start
				if i < t.lastRenderedLines-1 {
					fmt.Fprint(t.w, "\x1b[B") // Move down
				}
			}

			// Return to the top line to redraw everything correctly
			if t.lastRenderedLines > 1 {
				fmt.Fprintf(t.w, "\x1b[%dA", t.lastRenderedLines-1)
			}
		}

		// 2. Print the new render
		if !t.isHeadless {
			fmt.Fprint(t.w, prompt)
		}
		fmt.Fprint(t.w, content)

		// 3. Move cursor back to the correct position within the buffer
		if buffer.CursorPosition() < buffer.Len() {
			offset := buffer.Len() - buffer.CursorPosition()
			fmt.Fprintf(t.w, "\x1b[%dD", offset)
		}

		t.lastRenderedLines = lines
	}
}

// calculateRenderedLines calculates how many lines prompt + content occupies.
func (t *TUI) calculateRenderedLines(prompt, content string, width int) int {
	if width <= 0 {
		width = 80
	}
	pWidth := VisibleWidth(prompt)
	cWidth := VisibleWidth(content)

	if pWidth >= width {
		return (pWidth + cWidth + width - 1) / width
	}

	remainingInFirstLine := width - pWidth
	if cWidth <= remainingInFirstLine {
		return 1
	}

	extraChars := cWidth - remainingInFirstLine
	return 1 + (extraChars+width-1)/width
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
	err := t.term.disableRawMode(t.originalState)
	if err != nil {
		return err
	}
	t.originalState = nil
	return nil
}
