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
func (t *FileWriteTool) Execute(args map[string]interface{}) (string, error) {
	var a FileWriteArgs
	if err := mapToStruct(args, &a); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}
	path := a.Path
	if t.BaseDir != "" {
		path = filepath.Join(t.BaseDir, path)
	}

	data := []byte(a.Content)
	err := os.WriteFile(path, data, 0644)
	if err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	return fmt.Sprintf("Successfully wrote %d bytes to %s", len(data), path), nil
}
func (t *FileWriteTool) JSONSchema() map[string]interface{} {
	return map[string]interface{}{
		"name":        "write_file",
		"description": "Writes content to a file, overwriting if it exists.",
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
