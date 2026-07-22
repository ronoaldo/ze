package tools

import (
	"fmt"
	"os"
	"path/filepath"
)

// FileWriteTool implements writing a file (overwrites).
type FileWriteTool struct {
	BaseDir string // empty = cwd
}

func (t *FileWriteTool) Name() string { return "write_file" }
func (t *FileWriteTool) Execute(args map[string]interface{}) (ToolResult, error) {
	var a FileWriteArgs
	if err := mapToStruct(args, &a); err != nil {
		return ToolResult{}, fmt.Errorf("invalid arguments: %w", err)
	}
	path := a.Path
	if t.BaseDir != "" {
		path = filepath.Join(t.BaseDir, path)
	}

	data := []byte(a.Content)
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return ToolResult{}, fmt.Errorf("failed to create directory: %w", err)
	}
	err := os.WriteFile(path, data, 0644)
	if err != nil {
		return ToolResult{}, fmt.Errorf("failed to write file: %w", err)
	}

	fullMsg := fmt.Sprintf("Successfully wrote %d bytes to %s", len(data), path)
	return ToolResult{
		FullResult: fullMsg,
		Summary:    fmt.Sprintf("%d bytes", len(data)),
	}, nil
}
func (t *FileWriteTool) JSONSchema() map[string]interface{} {
	return map[string]interface{}{
		"name":        "write_file",
		"description": "Writes content to a file, overwriting if it exists. Creates any missing directories in the path. IMPORTANT: You must first provide the 'path' of the file, and then the full 'content' to be written. Ensure the file path is correct before providing the content.",
		"parameters": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"path":    map[string]interface{}{"type": "string"},
				"content": map[string]interface{}{"type": "string"},
			},
			"required": []string{"path", "content"},
		},
	}
}
