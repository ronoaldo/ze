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
	"time"

	"github.com/ronoaldo/ze/internal/agent"
	"github.com/ronoaldo/ze/internal/tools"
)

var (
	reInserts = regexp.MustCompile(`(\d+)\s+insertions?\(?\+?\)?`)
	reDeletes = regexp.MustCompile(`(\d+)\s+deletions?\(?\-?\)?`)
)

// Palette defines the colors used in the TUI.
type Palette struct {
	Reset     string
	Bold      string
	Dim       string
	Cyan      string
	Green     string
	Red       string
	Yellow    string
	Italic    string
	Underline string
}

func DefaultPalette() Palette {
	return Palette{
		Reset:     "\x1b[0m",
		Bold:      "\x1b[1m",
		Dim:       "\x1b[2m",
		Cyan:      "\x1b[36m",
		Green:     "\x1b[32m",
		Red:       "\x1b[31m",
		Yellow:    "\x1b[33m",
		Italic:    "\x1b[3m",
		Underline: "\x1b[4m",
	}
}

func NoColorPalette() Palette {
	return Palette{
		Reset:     "",
		Bold:      "",
		Dim:       "",
		Cyan:      "",
		Green:     "",
		Red:       "",
		Yellow:    "",
		Italic:    "",
		Underline: "",
	}
}

// TUI is the terminal user interface.
type TUI struct {
	w             io.Writer
	r             io.Reader
	reader        *bufio.Reader
	verbose       bool
	showThinking  bool
	palette       Palette
	rng           *rand.Rand
	isHeadless    bool
	messagePrefix string
}

func isUTF8Locale() bool {
	lang := strings.ToUpper(os.Getenv("LANG"))
	lcAll := strings.ToUpper(os.Getenv("LC_ALL"))
	lcCtype := strings.ToUpper(os.Getenv("LC_CTYPE"))
	return strings.Contains(lang, "UTF-8") || strings.Contains(lcAll, "UTF-8") ||
		strings.Contains(lcCtype, "UTF-8") ||
		strings.Contains(lang, "UTF8") || strings.Contains(lcAll, "UTF8") ||
		strings.Contains(lcCtype, "UTF8") ||
		strings.Contains(lang, "UTF8") ||
		strings.Contains(lcAll, "UTF8")
}

func (t *TUI) Run(handler func(msg string) (string, agent.AgentStats, error), isMultiline func() bool) error {
	for {
		// Print prompt
		if !t.isHeadless {
			if isMultiline != nil && isMultiline() {
				// No prompt in multiline mode as requested
			} else {
				fmt.Fprintf(t.w, "%s%s%s%s %s%s%s ", t.palette.Bold, t.palette.Cyan, "ze", t.palette.Reset, t.palette.Cyan, ">", t.palette.Reset)
			}
		}

		// Read input
		input, err := t.readLine()
		if err != nil {
			return err
		}

		// Echo input if headless
		if t.isHeadless {
			fmt.Fprintf(t.w, "prompt > %s\n", input)
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

func (t *TUI) ReportToolExecution(toolName string, args string, res tools.ToolResult, err error) {
	summary := t.summarizeArgs(toolName, args)
	header := fmt.Sprintf("%s%s%s%s('%s')",
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

// ReportStatus displays the current status and performance metrics.
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
		verbose:       verbose,
		showThinking:  showThinking,
		palette:       palette,
		rng:           rand.New(rand.NewSource(time.Now().UnixNano())),
		isHeadless:    !isTTY,
		messagePrefix: "* ",
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
		fmt.Fprintf(t.w, "%s%s%s\n", t.palette.Dim, RenderMarkdown(content, t.markdownStyle()), t.palette.Reset)
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

func (t *TUI) IsHeadless() bool {
	return t.isHeadless
}

func (t *TUI) getTerminalWidth() int {
	w, _ := getTerminalSize()
	return w
}
