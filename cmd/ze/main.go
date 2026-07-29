package main

import (
	_ "embed"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
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
	DefaultURL          = "http://localhost:1234"
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
	ModelName       string
	SessionID       string
	Version         bool
	Verbose         bool
	VerboseAPICalls bool
	MaxIteration    int
	ShowThinking    bool
	NoColor         bool
}

// ParseConfig parses command line arguments and environment variables.
func ParseConfig(args []string, env map[string]string) (*Config, error) {
	fs := flag.NewFlagSet("ze", flag.ContinueOnError)

	defaultURL := DefaultURL
	if val, ok := env["LLAMA_URL"]; ok && val != "" {
		defaultURL = val
	}

	defaultTimeout := DefaultTimeoutStr
	if val, ok := env["LLAMA_TIMEOUT"]; ok && val != "" {
		defaultTimeout = val
	}

	urlFlag := fs.String("url", defaultURL, "Llama server URL")
	modelFlag := fs.String("model", "", "Model name to use")
	timeoutFlag := fs.String("timeout", defaultTimeout, "Timeout duration (e.g. 60s, 5m)")
	versionFlag := fs.Bool("version", false, "Show version")
	vShortFlag := fs.Bool("v", false, "Show version (short)")
	verboseFlag := fs.Bool("verbose", false, "Enable verbose tool output")
	verboseAPICallsFlag := fs.Bool("verbose-api-calls", false, "Log raw API requests and responses")
	maxIterFlag := fs.Int("max-iterations", DefaultMaxIteration, "Maximum number of agent iterations")
	showThinkingFlag := fs.Bool("show-thinking", false, "Show thinking process in the UI")
	noColorFlag := fs.Bool("no-color", false, "Disable color output")
	sessionFlag := fs.String("session", "", "Session ID to resume a conversation")
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
		ModelName:       *modelFlag,
		SessionID:       *sessionFlag,
		Version:         *versionFlag || *vShortFlag,
		Verbose:         *verboseFlag,
		VerboseAPICalls: *verboseAPICallsFlag,
		MaxIteration:    *maxIterFlag,
		ShowThinking:    *showThinkingFlag,
		NoColor:         *noColorFlag,
	}, nil
}

func main() {
	cfg, err := ParseConfig(os.Args[1:], osEnvironAsMap())
	if err != nil {
		if !strings.Contains(err.Error(), "flag has no usage") && !strings.Contains(err.Error(), "help") {
			fmt.Fprintf(os.Stderr, "Error parsing flags: %v\n", err)
		}
		os.Exit(1)
	}

	if cfg.Version {
		fmt.Printf("ze version %s\ncommit: %s\ndate: %s\n", version, commit, date)
		return
	}

	client := llm.NewLlamaServerClient(cfg.URL, cfg.Timeout, cfg.VerboseAPICalls)

	availableModels, err := client.ListModels()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not list models from %s: %v\nUsing hardware detection.\n", cfg.URL, err)
		availableModels = nil
	}

	modelName := selectModel(availableModels, cfg.ModelName)

	// Register tools
	availableTools := []tools.Tool{
		// File system tools
		&tools.FileReadTool{},
		&tools.FileWriteTool{},
		&tools.ListFilesTool{},
		&tools.RemoveFileTool{},
		&tools.MoveFileTool{},

		// Code manipulation and inspection tools
		&tools.EditFileTool{},
		&tools.GoTool{},
		&tools.WebFetchTool{},
		&tools.GitTool{},
	}

	t := tui.New(cfg.Verbose, cfg.ShowThinking, cfg.NoColor)

	baseDir := os.Getenv("ZE_HOME")
	if baseDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting user home: %v\n", err)
			os.Exit(1)
		}
		baseDir = filepath.Join(home, ".config", "ze")
	}
	logger, err := agent.NewFileLogger(baseDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Close()

	sm, err := agent.NewSessionManager()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing session manager: %v\n", err)
		os.Exit(1)
	}

	sessionID, err := sm.GenerateSessionID()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating session ID: %v\n", err)
		os.Exit(1)
	}

	if cfg.SessionID != "" {
		sessionID = cfg.SessionID
	}

	zeAgent := agent.NewAgent(
		client,
		modelName,
		availableTools,
		agent.WithLogger(logger),
		agent.WithVerbose(cfg.Verbose),
		agent.WithMaxIteration(cfg.MaxIteration),
		agent.WithShowThinking(cfg.ShowThinking),
		agent.WithSession(sessionID, sm),
		agent.WithReporter(t),
	)

	if cfg.SessionID != "" {
		history, err := sm.LoadSession(cfg.SessionID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not load session %s: %v\n", cfg.SessionID, err)
		} else if history != nil {
			zeAgent.History = history
		}
	}

	commands.RegisterCommands()

	inputHandler := tui.NewInputHandler(
		commands.ExecuteCommand,
		func(input string) (string, agent.AgentStats, error) {
			res, stats, llmErr := zeAgent.Run(input)
			if llmErr != nil {
				return fmt.Sprintf("Error: %v", llmErr), stats, nil
			}
			return res, stats, nil
		},
	)

	if !t.IsHeadless() {
		printNeofetch(modelName, cfg, sessionID)
	}

	err = t.Run(func(msg string) (string, agent.AgentStats, error) {
		return inputHandler.Process(zeAgent, msg)
	}, inputHandler.IsMultiline)

	if err != nil {
		if errors.Is(err, commands.ErrQuit) || errors.Is(err, io.EOF) {
			os.Exit(0)
		}
		fmt.Fprintf(os.Stderr, "\nError: %v\n", err)
		os.Exit(1)
	}
}

func printNeofetch(modelName string, cfg *Config, sessionID string) {
	info := []string{
		fmt.Sprintf("Model:       %s", modelName),
		fmt.Sprintf("Server:      %s", cfg.URL),
		fmt.Sprintf("Timeout:     %s", cfg.Timeout),
		fmt.Sprintf("Verbose:     %v", cfg.Verbose),
		fmt.Sprintf("API Verbose: %v", cfg.VerboseAPICalls),
		fmt.Sprintf("Session:     %s", sessionID),
	}

	fmt.Fprintln(os.Stderr, "")

	logoLines := strings.Split(logoEmbed, "\n")

	if len(logoLines) > 0 && logoLines[len(logoLines)-1] == "" {
		logoLines = logoLines[:len(logoLines)-1]
	}

	maxLines := len(logoLines)
	if len(info) > maxLines {
		maxLines = len(info)
	}

	for i := 0; i < maxLines; i++ {
		if i < len(logoLines) {
			line := logoLines[i]
			fmt.Fprint(os.Stderr, line)
			if len(line) < 20 {
				fmt.Fprint(os.Stderr, strings.Repeat(" ", 20-len(line)))
			} else {
				fmt.Fprint(os.Stderr, "  ")
			}
		} else {
			fmt.Fprint(os.Stderr, "                      ")
		}

		if i < len(info) {
			fmt.Fprintln(os.Stderr, info[i])
		} else {
			fmt.Fprintln(os.Stderr, "")
		}
	}
	fmt.Fprintln(os.Stderr, "")
}

func selectModel(availableModels []llm.ModelInfo, userModel string) string {
	if userModel != "" {
		return userModel
	}

	for _, m := range availableModels {
		if m.Status == "loaded" && strings.Contains(strings.ToLower(m.ID), "gemma") {
			return m.ID
		}
	}

	for _, m := range availableModels {
		if m.Status == "loaded" {
			fmt.Fprintf(os.Stderr, "Note: model '%s' is loaded but not a Gemma 4. Using it anyway.\n", m.ID)
			return m.ID
		}
	}

	return ""
}

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
