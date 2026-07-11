package main

import (
	"errors"
	_ "embed"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/ronoaldo/ze/internal/agent"
	"github.com/ronoaldo/ze/internal/commands"
	"github.com/ronoaldo/ze/internal/llm"
	"github.com/ronoaldo/ze/internal/tools"
	"github.com/ronoaldo/ze/internal/tui"
)

//go:embed logo.txt
var logoEmbed string

// Default configuration values
const (
	DefaultURL          = "http://localhost:8084"
	DefaultTimeoutStr   = "5m"
	DefaultMaxIteration = 50
)

// Version metadata injected by GoReleaser ldflags.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

// Config holds the application configuration.
type Config struct {
	URL             string
	Timeout         time.Duration
	Version         bool
	Verbose         bool
	VerboseAPICalls bool
	MaxIteration    int
	ShowThinking    bool
}

// ParseConfig parses command line arguments and environment variables.
// It follows the priority: Flag > Environment Variable > Default.
func ParseConfig(args []string, env map[string]string) (*Config, error) {
	fs := flag.NewFlagSet("ze", flag.ContinueOnError)

	// Default values from ENV or Constants
	defaultURL := DefaultURL
	if val, ok := env["LLAMA_URL"]; ok && val != "" {
		defaultURL = val
	}

	defaultTimeout := DefaultTimeoutStr
	if val, ok := env["LLAMA_TIMEOUT"]; ok && val != "" {
		defaultTimeout = val
	}

	// Define flags
	urlFlag := fs.String("url", defaultURL, "Llama server URL")
	timeoutFlag := fs.String("timeout", defaultTimeout, "Timeout duration (e.g. 60s, 5m)")
	versionFlag := fs.Bool("version", false, "Show version")
	vShortFlag := fs.Bool("v", false, "Show version (short)")
	verboseFlag := fs.Bool("verbose", false, "Enable verbose tool output")
	verboseAPICallsFlag := fs.Bool("verbose-api-calls", false, "Log raw API requests and responses")
	maxIterFlag := fs.Int("max-iterations", DefaultMaxIteration, "Maximum number of agent iterations")
	showThinkingFlag := fs.Bool("show-thinking", false, "Show thinking process in the UI")
	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	timeout, err := time.ParseDuration(*timeoutFlag)
	if err != nil {
		return nil, fmt.Errorf("invalid timeout duration: %w", err)
	}

	return &Config{
		URL:             *urlFlag,
		Timeout:         timeout,
		Version:         *versionFlag || *vShortFlag,
		Verbose:         *verboseFlag,
		VerboseAPICalls: *verboseAPICallsFlag,
		MaxIteration:    *maxIterFlag,
		ShowThinking:    *showThinkingFlag,
	}, nil
}

func main() {
	// Use os.Args[1:] to exclude the program name
	cfg, err := ParseConfig(os.Args[1:], osEnvironAsMap())
	if err != nil {
		// If it's a help message or usage error, flag.Parse already printed it.
		if !strings.Contains(err.Error(), "flag has no usage") && !strings.Contains(err.Error(), "help") {
			fmt.Fprintf(os.Stderr, "Error parsing flags: %v\n", err)
		}
		os.Exit(1)
	}

	if cfg.Version {
		fmt.Printf("ze version %s\ncommit: %s\ndate: %s\n", version, commit, date)
		return
	}

	// Create real client
	client := llm.NewLlamaServerClient(cfg.URL, cfg.Timeout, cfg.VerboseAPICalls)

	// Discover available models from the server
	availableModels, err := client.ListModels()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not list models from %s: %v\nUsing hardware detection.\n", cfg.URL, err)
		availableModels = nil
	}

	// Select best model: loaded Gemma 4 > any loaded > hardware detection
	modelName := selectModel(availableModels)

	// Register tools
	availableTools := []tools.Tool{
		&tools.FileReadTool{},
		&tools.FileWriteTool{},
		&tools.ListFilesTool{},
		&tools.GoDocTool{},
		&tools.EditFileTool{},
		&tools.GoTestTool{},
		&tools.DiffTool{},
		&tools.RemoveFileTool{},
	}

	// Create TUI
	t := tui.New(cfg.Verbose, cfg.ShowThinking)

	// Create agent with full multi-step loop and reporter
	zeAgent := agent.NewAgent(client, modelName, availableTools, cfg.Verbose, cfg.MaxIteration, cfg.ShowThinking)
	zeAgent.Reporter = t

	// Register commands
	commands.RegisterCommands()

	// Show model info using Neofetch-style banner
	printNeofetch(modelName, cfg)

	// Run TUI — wraps the agent's Run method
	err = t.Run(func(msg string) (string, agent.AgentStats, error) {
		resp, cmdErr := commands.ExecuteCommand(msg)
		if cmdErr == nil {
			return resp, agent.AgentStats{}, nil
		}
		if errors.Is(cmdErr, commands.ErrQuit) {
			return "", agent.AgentStats{}, cmdErr
		}

		// If it's a command that failed (e.g. unknown command), return as response
		if strings.HasPrefix(msg, "/") {
			return fmt.Sprintf("Error: %v", cmdErr), agent.AgentStats{}, nil
		}

		// Otherwise, it's a user message for the agent
		res, stats, llmErr := zeAgent.Run(msg)
		if llmErr != nil {
			return fmt.Sprintf("Error: %v", llmErr), stats, nil
		}
		return res, stats, nil
	})

	if err != nil {
		if errors.Is(err, commands.ErrQuit) {
			os.Exit(0)
		}
		fmt.Fprintf(os.Stderr, "\nError: %v\n", err)
		os.Exit(1)
	}
}

// printNeofetch displays a neofetch-style banner with ASCII art from logo.txt and system info.
func printNeofetch(modelName string, cfg *Config) {
	info := []string{
		fmt.Sprintf("Model:       %s", modelName),
		fmt.Sprintf("Server:      %s", cfg.URL),
		fmt.Sprintf("Timeout:     %s", cfg.Timeout),
		fmt.Sprintf("Verbose:     %v", cfg.Verbose),
		fmt.Sprintf("API Verbose: %v", cfg.VerboseAPICalls),
	}

	fmt.Fprintln(os.Stderr, "")
	
	// Clean up logo data: split into lines. We don't use TrimSpace on the whole block
	// because it would destroy the intended indentation of the ASCII art.
	logoLines := strings.Split(logoEmbed, "\n")

	// Remove the last empty line if the file ends with a newline
	if len(logoLines) > 0 && logoLines[len(logoLines)-1] == "" {
		logoLines = logoLines[:len(logoLines)-1]
	}
	
	// We iterate based on the maximum number of elements to ensure everything is printed
	maxLines := len(logoLines)
	if len(info) > maxLines {
		maxLines = len(info)
	}

	for i := 0; i < maxLines; i++ {
		// Logo line
		if i < len(logoLines) {
			line := logoLines[i]
			// Minimal padding for alignment
			fmt.Fprint(os.Stderr, line)
			// Ensure there's enough separation between logo and info
			if len(line) < 20 {
				fmt.Fprint(os.Stderr, strings.Repeat(" ", 20-len(line)))
			} else {
				fmt.Fprint(os.Stderr, "  ")
			}
		} else {
			// Padding if logo has fewer lines than info
			fmt.Fprint(os.Stderr, "                      ")
		}

		// Info line
		if i < len(info) {
			fmt.Fprintln(os.Stderr, info[i])
		} else {
			fmt.Fprintln(os.Stderr, "")
		}
	}
	fmt.Fprintln(os.Stderr, "")
}

// selectModel picks the best model from the server or falls back to hardware detection.
// Priority: loaded Gemma 4 > any loaded model > hardware-detect best.
func selectModel(availableModels []llm.ModelInfo) string {
	// 1. Prefer a loaded Gemma 4 model
	for _, m := range availableModels {
		if m.Status == "loaded" && strings.Contains(strings.ToLower(m.ID), "gemma") {
			return m.ID
		}
	}

	// 2. Any loaded model (with a note)
	for _, m := range availableModels {
		if m.Status == "loaded" {
			fmt.Fprintf(os.Stderr, "Note: model '%s' is loaded but not a Gemma 4. Using it anyway.\n", m.ID)
			return m.ID
		}
	}

	// 3. Fall back to hardware detection (no model loaded on server)
	res := llm.DetectHardware()
	return llm.SelectBestModel(res)
}

// osEnvironAsMap converts os.Environ() to a map for testing/parsing.
func osEnvironAsMap() map[string]string {
	env := make(map[string]string)
	for _, e := range os.Environ() {
		pair := strings.SplitN(e, "=", 2)
		if len(pair) == 2 {
			env[pair[0]] = pair[1]
		}
	}
	return env
}
