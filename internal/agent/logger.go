package agent

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/ronoaldo/ze/internal/llm"
)

type LogType string

const (
	LogTypeUser    LogType = "user_input"
	LogTypeLLMReq  LogType = "llm_request"
	LogTypeLLMResp LogType = "llm_response"
	LogTypeTool    LogType = "tool_call"
	LogTypeError   LogType = "error"
)

type LogEntry struct {
	Timestamp string            `json:"timestamp"`
	Type      LogType           `json:"type"`
	Content   string            `json:"content,omitempty"`
	Metadata  map[string]string `json:"metadata,omitempty"`
	Request   interface{}       `json:"request,omitempty"`
	Response  interface{}       `json:"response,omitempty"`
}

// Logger defines an interface for logging agent activity.
type Logger interface {
	LogUserMessage(content string)
	LogLLMRequest(req *llm.ChatRequest)
	LogLLMResponse(resp *llm.ChatResponse)
	LogToolCall(toolName string, args interface{}, result interface{}, err error)
	LogError(err error, context string)
}

// FileLogger implements the Logger interface by writing to a file.
type FileLogger struct {
	file *os.File
}

// NewFileLogger creates a new FileLogger with the default directory.
func NewFileLogger(baseDir string) (*FileLogger, error) {
	logDir := filepath.Join(baseDir, "logs")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, err
	}

	logFile := filepath.Join(logDir, "agent_execution.log")
	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	return &FileLogger{file: f}, nil
}

// Close closes the log file.
func (l *FileLogger) Close() error {
	return l.file.Close()
}

func (l *FileLogger) write(entry LogEntry) {
	entry.Timestamp = time.Now().Format(time.RFC3339)
	data, err := json.Marshal(entry)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to marshal log entry: %v\n", err)
		return
	}
	if _, err := l.file.Write(append(data, '\n')); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write log entry: %v\n", err)
	}
}

func (l *FileLogger) LogUserMessage(content string) {
	l.write(LogEntry{
		Type:    LogTypeUser,
		Content: content,
	})
}

func (l *FileLogger) LogLLMRequest(req *llm.ChatRequest) {
	l.write(LogEntry{
		Type:     LogTypeLLMReq,
		Request:  req,
	})
}

func (l *FileLogger) LogLLMResponse(resp *llm.ChatResponse) {
	l.write(LogEntry{
		Type:     LogTypeLLMResp,
		Response: resp,
	})
}

func (l *FileLogger) LogToolCall(toolName string, args interface{}, result interface{}, err error) {
	metadata := make(map[string]string)
	metadata["tool_name"] = toolName
	if err != nil {
		metadata["error"] = err.Error()
	}

	l.write(LogEntry{
		Type:     LogTypeTool,
		Metadata: metadata,
		Request:  args,
		Response: result,
	})
}

func (l *FileLogger) LogError(err error, context string) {
	l.write(LogEntry{
		Type:    LogTypeError,
		Content: fmt.Sprintf("%s: %v", context, err),
	})
}
