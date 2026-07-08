package main

import (
	"errors"
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
}

// ParseConfig parses command line arguments and environment variables.
// It follows the priority: Flag > Environment Variable > Default.
func ParseConfig(args []string, env map[string]string) (*Config, error) {
	fs := flag.NewFlagSet("ze", flag.ContinueOnError)

	// Default values
	defaultURL := "http://localhost:8084"
	if val, ok := env["LLAMA_URL"]; ok && val != "" {
		defaultURL = val
	}

	defaultTimeout := "60s"
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
	}

	// Create TUI
	t := tui.New(cfg.Verbose)

	// Create agent with full multi-step loop and reporter
	zeAgent := agent.NewAgent(client, modelName, availableTools, cfg.Verbose)
	zeAgent.Reporter = t

	// Register commands
	commands.RegisterCommands()

	// Show model info
	fmt.Fprintf(os.Stderr, "[ Model: %s | Server: %s | Timeout: %v | Verbose: %v | API Verbose: %v ]\n", modelName, cfg.URL, cfg.Timeout, cfg.Verbose, cfg.VerboseAPICalls)

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
