package main

import (
	"context"
	_ "embed"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
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
	err := run()
	if err != nil {
		// Se for um encerramento esperado, terminamos naturalmente para permitir que os defers de run() concluam.
		if errors.Is(err, commands.ErrQuit) || errors.Is(err, io.EOF) || errors.Is(err, context.Canceled) || errors.Is(err, tui.ErrInterrupt) {
			return
		}
		fmt.Fprintf(os.Stderr, "\nError: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := ParseConfig(os.Args[1:], osEnvironAsMap())
	if err != nil {
		if !strings.Contains(err.Error(), "flag has no usage") && !strings.Contains(err.Error(), "help") {
			fmt.Fprintf(os.Stderr, "Error parsing flags: %v\n", err)
		}
		return err
	}

	if cfg.Version {
		fmt.Printf("ze version %s\ncommit: %s\ndate: %s\n", version, commit, date)
		return nil
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
			return fmt.Errorf("Error getting user home: %w", err)
		}
		baseDir = filepath.Join(home, ".config", "ze")
	}
	logger, err := agent.NewFileLogger(baseDir)
	if err != nil {
		return fmt.Errorf("Error initializing logger: %w", err)
	}
	defer logger.Close()

	sm, err := agent.NewSessionManager()
	if err != nil {
		return fmt.Errorf("Error initializing session manager: %w", err)
	}

	sessionID, err := sm.GenerateSessionID()
	if err != nil {
		return fmt.Errorf("Error generating session ID: %w", err)
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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Register signal handler
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		t.DisableRawMode()
		fmt.Println("\nExiting...")
		cancel()
	}()

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
		func(ctx context.Context, input string) (string, agent.AgentStats, error) {
			res, stats, llmErr := zeAgent.Run(ctx, input)
			if llmErr != nil {
				if errors.Is(llmErr, tui.ErrInterrupt) {
					return "", stats, llmErr
				}
				return fmt.Sprintf("Error: %v", llmErr), stats, nil
			}
			return res, stats, nil
		},
	)

	if !t.IsHeadless() {
		if err := t.EnableRawMode(); err != nil {
			return fmt.Errorf("Error enabling raw mode: %w", err)
		}
		defer t.DisableRawMode()
	}

	// Print banner
	if !t.IsHeadless() {
		printNeofetch(os.Stdout, modelName, cfg, sessionID)
	}

	err = t.Run(ctx, func(ctx context.Context, msg string) (string, agent.AgentStats, error) {
		return inputHandler.Process(ctx, zeAgent, msg)
	}, inputHandler.IsMultiline)

	return err
}

func printNeofetch(w io.Writer, modelName string, cfg *Config, sessionID string) {
	info := []string{
		fmt.Sprintf("Model:       %s", modelName),
		fmt.Sprintf("Server:      %s", cfg.URL),
		fmt.Sprintf("Timeout:     %s", cfg.Timeout),
		fmt.Sprintf("Verbose:     %v", cfg.Verbose),
		fmt.Sprintf("API Verbose: %v", cfg.VerboseAPICalls),
		fmt.Sprintf("Session:     %s", sessionID),
	}

	fmt.Fprintln(w, "")

	logoLines := strings.Split(logoEmbed, "\n")
	if len(logoLines) > 0 && logoLines[len(logoLines)-1] == "" {
		logoLines = logoLines[:len(logoLines)-1]
	}

	maxLines := len(logoLines)
	if len(info) > maxLines {
		maxLines = len(info)
	}

	for i := 0; i < maxLines; i++ {
		// Print logo line
		if i < len(logoLines) {
			line := logoLines[i]
			fmt.Fprint(w, line)
			padding := 20 - len(line)
			if padding > 0 {
				fmt.Fprint(w, strings.Repeat(" ", padding))
			} else if padding < 0 {
				fmt.Fprint(w, "  ")
			}
		} else {
			fmt.Fprint(w, "                      ")
		}

		// Print info line
		if i < len(info) {
			fmt.Fprintln(w, info[i])
		} else {
			fmt.Fprintln(w, "")
		}
	}
	fmt.Fprintln(w, "")
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
