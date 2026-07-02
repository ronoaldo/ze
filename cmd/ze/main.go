package main

import (
	"fmt"
	"os"
	"strings"

	"ze/internal/agent"
	"ze/internal/llm"
	"ze/internal/tools"
	"ze/internal/tui"
)

// Version metadata injected by GoReleaser ldflags.
var (
	version  = "dev"
	commit   = "none"
	date     = "unknown"
)

func main() {
	// Handle --version flag
	for _, arg := range os.Args[1:] {
		if arg == "--version" || arg == "-v" {
			fmt.Printf("ze version %s\ncommit: %s\ndate: %s\n", version, commit, date)
			return
		}
	}

	// Determine llama-server URL: flag > env var > default
	url := flagOrEnvOr("http://localhost:8080", "-url", "LLAMA_URL")

	// Create real client
	client := llm.NewLlamaServerClient(url)

	// Discover available models from the server
	availableModels, err := client.ListModels()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not list models from %s: %v\nUsing hardware detection.\n", url, err)
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
	}

	// Create agent with full multi-step loop
	zeAgent := agent.NewAgent(client, modelName, availableTools)

	// Show model info
	fmt.Fprintf(os.Stderr, "Model: %s\nServer: %s\n", modelName, url)

	// Create TUI
	t := tui.New()

	// Run TUI — wraps the agent's Run method
	err = t.Run(func(msg string) (string, error) {
		return zeAgent.Run(msg)
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "\nError: %v\n", err)
		os.Exit(1)
	}
}

// flagOrEnvOr returns the first non-empty value among flag, env var, and default.
func flagOrEnvOr(defaultVal, flagName, envVar string) string {
	// Try flag first
	for i := 1; i < len(os.Args); i++ {
		if os.Args[i] == "-"+flagName && i+1 < len(os.Args) {
			return os.Args[i+1]
		}
		if strings.HasPrefix(os.Args[i], "-"+flagName+"=") {
			return strings.TrimPrefix(os.Args[i], "-"+flagName+"=")
		}
	}
	// Try env var
	if val := os.Getenv(envVar); val != "" {
		return val
	}
	return defaultVal
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
