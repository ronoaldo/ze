package agent

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/ronoaldo/ze/internal/llm"
	"github.com/ronoaldo/ze/internal/prompt"
	"github.com/ronoaldo/ze/internal/tools"
)

// AgentStats contains performance metrics for an agent run.
type AgentStats struct {
	Duration      time.Duration
	PromptTokens  int
	CompTokens    int
	TotalTokens   int
	TokensPerSec  float64
}

// AgentReporter defines an interface for reporting agent activity to the UI.
type AgentReporter interface {
	ReportToolCall(toolName string, args string)
	ReportToolCallVerbose(toolName string, args string)
	ReportToolResult(toolName string, result string, err error)
	ReportReasoning(content string)
}


// Agent represents the core programming agent.
type Agent struct {
	Client       llm.Client
	Model        string // The selected model name
	Tools        map[string]tools.Tool
	ToolDefs     []llm.ToolDefinition
	History      []llm.ChatMessage
	Reporter     AgentReporter // Optional reporter for UI updates
	Verbose      bool
	MaxIteration int
	ShowThinking bool
}

func NewAgent(client llm.Client, model string, availableTools []tools.Tool, verbose bool, maxIter int, showThinking bool) *Agent {
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
	
	if maxIter <= 0 {
		maxIter = 20
	}

	return &Agent{
		Client:       client,
		Model:        model,
		Tools:        toolMap,
		ToolDefs:     toolDefs,
		History:      []llm.ChatMessage{},
		Verbose:      verbose,
		MaxIteration: maxIter,
		ShowThinking: showThinking,
	}
}

// Run processes a user message and returns the agent's response, stats, or an error.
func (a *Agent) Run(userInput string) (string, AgentStats, error) {
	startTime := time.Now()
	var lastPromptTokens int
	var lastCompTokens int
	var lastChatDuration time.Duration

	// 1. Add user message to history
	a.History = append(a.History, llm.ChatMessage{Role: "user", Content: userInput})

	// 2. Multi-step loop
	for i := 0; i < a.MaxIteration; i++ {
		req := a.prepareRequest()

		chatStartTime := time.Now()
		resp, err := a.Client.Chat(req)
		lastChatDuration = time.Since(chatStartTime)

		if err != nil {
			return "", AgentStats{}, fmt.Errorf("LLM error: %w", err)
		}

		if len(resp.Choices) == 0 {
			return "No response from model.", AgentStats{}, nil
		}

		// Track tokens for the last call
		lastPromptTokens = resp.Usage.PromptTokens
		lastCompTokens = resp.Usage.CompletionTokens

		assistantMsg := resp.Choices[0].Message
		a.History = append(a.History, assistantMsg)

		// Report reasoning if present
		if assistantMsg.ReasoningContent != "" {
			if a.Reporter != nil {
				a.Reporter.ReportReasoning(assistantMsg.ReasoningContent)
			}
		}

		// Check for tool calls
		if len(assistantMsg.ToolCalls) > 0 {

			toolResults, err := a.handleToolCalls(assistantMsg.ToolCalls)
			if err != nil {
				return "", AgentStats{}, err
			}
			a.History = append(a.History, toolResults...)
			continue
		}

		// No more tool calls — this is the final answer
		duration := time.Since(startTime)
		var tokensPerSec float64
		if lastChatDuration > 0 && lastCompTokens > 0 {
			tokensPerSec = float64(lastCompTokens) / lastChatDuration.Seconds()
		}

		stats := AgentStats{
			Duration:     duration,
			PromptTokens: lastPromptTokens,
			CompTokens:   lastCompTokens,
			TotalTokens:  lastPromptTokens + lastCompTokens,
			TokensPerSec: tokensPerSec,
		}
		return assistantMsg.Content, stats, nil
	}

	return "", AgentStats{}, fmt.Errorf("reached max iterations (%d) without a final answer", a.MaxIteration)
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
			return nil, fmt.Errorf("failed to unmarshal tool arguments: %w. Arguments: %s", err, string(tc.Function.Arguments))
		}

		if a.Reporter != nil {
			a.Reporter.ReportToolCall(tc.Function.Name, string(tc.Function.Arguments))
		}

		tool, ok := a.Tools[tc.Function.Name]
		if !ok {
			errMsg := fmt.Sprintf("[Error: tool '%s' not found]", tc.Function.Name)
			toolMessages = append(toolMessages, llm.ChatMessage{
				Role:       "tool",
				ToolCallID: tc.ID,
				Content:    errMsg,
			})
			continue
		}

		result, err := tool.Execute(args)
		if err != nil {
			content := fmt.Sprintf("[Tool Error (%s): %v]\n\n%s", tc.Function.Name, err, result)
			toolMessages = append(toolMessages, llm.ChatMessage{
				Role:       "tool",
				ToolCallID: tc.ID,
				Content:    content,
			})
		} else {
			toolMessages = append(toolMessages, llm.ChatMessage{
				Role:       "tool",
				ToolCallID: tc.ID,
				Content:    result,
			})
		}

		if a.Reporter != nil {
			a.Reporter.ReportToolResult(tc.Function.Name, result, err)
		}
	}

	return toolMessages, nil
}
