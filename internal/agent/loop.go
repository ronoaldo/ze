package agent

import (
	"encoding/json"
	"fmt"

	"github.com/ronoaldo/ze/internal/llm"
	"github.com/ronoaldo/ze/internal/prompt"
	"github.com/ronoaldo/ze/internal/tools"
)

const maxIterations = 20

// AgentReporter defines an interface for reporting agent activity to the UI.
type AgentReporter interface {
	ReportToolCall(toolName string, args string)
	ReportToolResult(toolName string, result string, err error)
}

// Agent represents the core programming agent.
type Agent struct {
	Client   llm.Client
	Model    string // The selected model name
	Tools    map[string]tools.Tool
	ToolDefs []llm.ToolDefinition
	History  []llm.ChatMessage
	Reporter AgentReporter // Optional reporter for UI updates
}

func NewAgent(client llm.Client, model string, availableTools []tools.Tool) *Agent {
	toolMap := make(map[string]tools.Tool)
	toolDefs := make([]llm.ToolDefinition, 0, len(availableTools))
	for _, t := range availableTools {
		toolMap[t.Name()] = t
		
		schema := t.JSONSchema()
		schemaBytes, _ := json.Marshal(schema["parameters"])
		
		toolDefs = append(toolDefs, llm.ToolDefinition{
			Type: "function",
			Function: llm.FunctionDef{
				Name:        t.Name(),
				Description: schema["description"].(string),
				Parameters:  schemaBytes,
			},
		})
	}

	return &Agent{
		Client:   client,
		Model:    model,
		Tools:    toolMap,
		ToolDefs: toolDefs,
		History:  []llm.ChatMessage{},
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
		if len(assistantMsg.ToolCalls) > 0 {
			toolResults, err := a.handleToolCalls(assistantMsg.ToolCalls)
			if err != nil {
				return "", err
			}
			// Add tool results as "tool" messages to history.
			a.History = append(a.History, toolResults...)
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
		Tools:    a.ToolDefs,
	}
}

func (a *Agent) handleToolCalls(toolCalls []llm.ToolCall) ([]llm.ChatMessage, error) {
	var toolMessages []llm.ChatMessage
	for _, tc := range toolCalls {
		var args map[string]interface{}
		if err := json.Unmarshal(tc.Function.Arguments, &args); err != nil {
			return nil, fmt.Errorf("failed to unmarshal tool arguments: %w", err)
		}

		if a.Reporter != nil {
			a.Reporter.ReportToolCall(tc.Function.Name, string(tc.Function.Arguments))
		}

		tool, ok := a.Tools[tc.Function.Name]
		if !ok {
			errMsg := fmt.Sprintf("[Error: tool '%s' not found]", tc.Function.Name)
			toolMessages = append(toolMessages, llm.ChatMessage{
				Role:    "tool",
				Content: errMsg,
			})
			continue
		}

		result, err := tool.Execute(args)
		if err != nil {
			errResult := fmt.Sprintf("[Tool Error (%s): %v]", tc.Function.Name, err)
			toolMessages = append(toolMessages, llm.ChatMessage{
				Role:    "tool",
				Content: errResult,
			})
		} else {
			toolMessages = append(toolMessages, llm.ChatMessage{
				Role:    "tool",
				Content: result,
			})
		}

		if a.Reporter != nil {
			a.Reporter.ReportToolResult(tc.Function.Name, result, err)
		}
	}

	return toolMessages, nil
}
