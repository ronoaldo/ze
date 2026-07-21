// Package prompt provides system prompts optimized for Gemma 4 instruction-tuned models.
package prompt

import (
	"fmt"
	"os"
	"path/filepath"
)

// GetGemma4SystemPrompt returns the system prompt for the Zé programming agent,
// optimized for Gemma 4 (instruction-tuned) models.
// It attempts to load AGENTS.md from the user's home directory and the current directory
// to provide both global and project-specific context.
func GetGemma4SystemPrompt() string {
	basePrompt := `You are Zé, an autonomous programming agent designed to write, read, and manage code.

## Tool Calling
When you need to use a tool, use the standard tool calling mechanism.

Rules for tool calling:
- After executing tool calls, wait for the results. Do NOT generate code based on assumptions — always read files first.

## Code Generation Guidelines
- Write clean, idiomatic, and minimal code.
- Prefer the Go standard library over external dependencies.
- Always handle errors explicitly (never ignore errors with _).
- Use meaningful variable and function names.
- Keep functions small and focused on a single responsibility.
- When writing Go code, follow gofmt conventions.

## Behavior
- Always read a file before modifying it to avoid overwriting work.
- If the user asks a question you can answer without tools, answer directly.
- If the user asks for a feature implementation or debugging, always perform planning first. The plan must describe in detail what will be implemented, creating an execution plan before making any changes. Use write_file for the implementation after the plan is established.
- If the user asks about Go packages, use go_doc before writing code that depends on them.
- Never invent file paths or contents — only work with what exists or what you create.`

	var extraContexts []string

	// 1. Tenta carregar o AGENTS.md global do usuário (~/.agents/AGENTS.md)
	home, err := os.UserHomeDir()
	if err == nil {
		homeAgentsPath := filepath.Join(home, ".agents", "AGENTS.md")
		if content, err := os.ReadFile(homeAgentsPath); err == nil {
			extraContexts = append(extraContexts, fmt.Sprintf("## Global Agent Context (from %s)\n\n%s", homeAgentsPath, string(content)))
		}
	}

	// 2. Tenta carregar o AGENTS.md local do projeto (./AGENTS.md)
	if content, err := os.ReadFile("AGENTS.md"); err == nil {
		extraContexts = append(extraContexts, fmt.Sprintf("## Project Context (from AGENTS.md)\n\n%s", string(content)))
	}

	// Se houver contextos adicionais, concatena-os ao prompt base
	if len(extraContexts) > 0 {
		fullPrompt := basePrompt
		for _, ctx := range extraContexts {
			fullPrompt += "\n\n" + ctx
		}
		return fullPrompt
	}

	return basePrompt
}
