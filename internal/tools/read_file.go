package tools

import (
	"fmt"
	"os"
	"path/filepath"
)

// FileReadTool implements reading a file.
type FileReadTool struct {
	BaseDir string // empty = cwd
}

func (t *FileReadTool) Name() string { return "read_file" }
func (t *FileReadTool) Execute(args map[string]interface{}) (ToolResult, error) {
	var a FileReadArgs
	if err := mapToStruct(args, &a); err != nil {
		return ToolResult{}, fmt.Errorf("invalid arguments: %w", err)
	}
	path := a.Path
	if t.BaseDir != "" {
		path = filepath.Join(t.BaseDir, path)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return ToolResult{}, fmt.Errorf("failed to read file: %w", err)
	}

	return ToolResult{
		FullResult: string(content),
		Summary:    fmt.Sprintf("%d bytes", len(content)),
	}, nil
}
func (t *FileReadTool) JSONSchema() map[string]interface{} {
	return map[string]interface{}{
		"name":        "read_file",
		"description": "Reads the content of a file.",
		"parameters": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"path": map[string]interface{}{"type": "string"},
			},
			"required": []string{"path"},
		},
	}
}
