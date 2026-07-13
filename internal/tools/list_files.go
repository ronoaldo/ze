package tools

import (
	"fmt"
	"os"
)

// ListFilesTool implements listing files in a directory.
type ListFilesTool struct {
	BaseDir string // empty = cwd
}

func (t *ListFilesTool) Name() string { return "list_files" }
func (t *ListFilesTool) Execute(args map[string]interface{}) (ToolResult, error) {
	dir := "."
	if t.BaseDir != "" {
		dir = t.BaseDir
	}
	if val, ok := args["path"]; ok {
		if d, isStr := val.(string); isStr {
			dir = d
		}
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return ToolResult{}, fmt.Errorf("failed to list files: %w", err)
	}

	var fullResult string
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() {
			name += "/"
		}
		fullResult += fmt.Sprintf("- %s\n", name)
	}

	summary := fmt.Sprintf("%d items", len(entries))
	if len(entries) == 0 {
		summary = "0 items"
	}

	return ToolResult{
		FullResult: fullResult,
		Summary:    summary,
	}, nil
}
func (t *ListFilesTool) JSONSchema() map[string]interface{} {
	return map[string]interface{}{
		"name":        "list_files",
		"description": "Lists files and directories in a given path.",
		"parameters": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"path": map[string]interface{}{"type": "string"},
			},
		},
	}
}
