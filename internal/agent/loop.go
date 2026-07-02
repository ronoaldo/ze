package agent

import (
	"encoding/json"
	"fmt"
	"strings"

	"ze/internal/llm"
	"ze/internal/prompt"
	"ze/internal/tools"
)

const maxIterations = 20

// Agent represents the core programming agent.
type Agent struct {
	Client  llm.Client
	Model   string // The selected model name
	Tools   map[string]tools.Tool
	History []llm.ChatMessage
}

func NewAgent(client llm.Client, model string, availableTools []tools.Tool) *Agent {
	toolMap := make(map[string]tools.Tool)
	for _, t := range availableTools {
		toolMap[t.Name()] = t
	}

	return &Agent{
		Client:  client,
		Model:   model,
		Tools:   toolMap,
		History: []llm.ChatMessage{},
	}
}

// Run processes a user message and returns the agent's response or an error.
// It implements a multi-step loop: LLM → tool calls → LLM → ... → final answer.
func (a *Agent) Run(userInput string) (string, error) {
	// 1. Add user message to history
	a.History = append(a.History, llm.ChatMessage{Role: "user", Content: userInput})

	// 2. Multi-step loop: call LLM, execute tools, repeat until no more tool calls
	for i := 0; i < maxIterations; i++ {
		req := a.prepareRequest()

		resp, err := a.Client.Chat(req)
		if err != nil {
			return "", fmt.Errorf("LLM error: %w", err)
		}

		if len(resp.Choices) == 0 {
			return "No response from model.", nil
		}

		assistantMsg := resp.Choices[0].Message
		a.History = append(a.History, assistantMsg)

		// Check for tool calls
		if strings.Contains(assistantMsg.Content, "TOOL_CALL:") {
			toolResults, err := a.handleToolCalls(assistantMsg.Content)
			if err != nil {
				return "", err
			}
			// Add tool results as a "tool" message to history so the LLM sees them
			a.History = append(a.History, llm.ChatMessage{Role: "tool", Content: toolResults})
			continue // loop back: call LLM again with the tool results
		}

		// No more tool calls — this is the final answer
		return assistantMsg.Content, nil
	}

	return "", fmt.Errorf("reached max iterations (%d) without a final answer", maxIterations)
}

func (a *Agent) prepareRequest() *llm.ChatRequest {
	messages := []llm.ChatMessage{
		{Role: "system", Content: prompt.GetGemma4SystemPrompt()},
	}
	messages = append(messages, a.History...)

	return &llm.ChatRequest{
		Model:    a.Model,
		Messages: messages,
	}
}

func (a *Agent) handleToolCalls(content string) (string, error) {
	// Parses TOOL_CALL:tool_name{json} lines, executes each tool,
	// and returns a summary of results to be added to the LLM history.
	lines := strings.Split(content, "\n")
	var results strings.Builder

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if !strings.HasPrefix(line, "TOOL_CALL:") {
			continue
		}

		rest := line[10:] // strip "TOOL_CALL:"
		braceIdx := strings.Index(rest, "{")
		if braceIdx < 0 {
			results.WriteString(fmt.Sprintf("[Error: invalid tool call format: %s]\n", line))
			continue
		}

		toolName := strings.TrimSpace(rest[:braceIdx])
		argsJSON := rest[braceIdx:]

		var args map[string]interface{}
		if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
			results.WriteString(fmt.Sprintf("[Error: failed to parse args for '%s': %v]\n", toolName, err))
			continue
		}

		tool, ok := a.Tools[toolName]
		if !ok {
			results.WriteString(fmt.Sprintf("[Error: tool '%s' not found]\n", toolName))
			continue
		}

		result, err := tool.Execute(args)
		if err != nil {
			results.WriteString(fmt.Sprintf("[Tool Error (%s): %v]\n", toolName, err))
		} else {
			results.WriteString(fmt.Sprintf("[Tool Result (%s)]: %s\n", toolName, result))
		}
	}

	return results.String(), nil
}
