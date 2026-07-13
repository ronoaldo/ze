package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// RemoveFileTool implements removing a file.
type RemoveFileTool struct {
	BaseDir string // empty = cwd
}

func (t *RemoveFileTool) Name() string { return "remove_file" }

func (t *RemoveFileTool) Execute(args map[string]interface{}) (ToolResult, error) {
	var a RemoveFileArgs
	if err := mapToStruct(args, &a); err != nil {
		return ToolResult{}, fmt.Errorf("invalid arguments: %w", err)
	}

	path := a.Path
	if t.BaseDir != "" {
		path = filepath.Join(t.BaseDir, path)
	}

	// Get absolute path of the target file
	absTarget, err := filepath.Abs(path)
	if err != nil {
		return ToolResult{}, fmt.Errorf("failed to get absolute path for target: %w", err)
	}

	// Get absolute path of the current working directory
	absWorkDir, err := filepath.Abs(".")
	if err != nil {
		return ToolResult{}, fmt.Errorf("failed to get absolute working directory: %w", err)
	}

	// Safety Check: Ensure the target is within the work directory
	// This prevents deleting files like /etc/passwd or ~/.ssh/id_rsa
	rel, err := filepath.Rel(absWorkDir, absTarget)
	if err != nil {
		return ToolResult{}, fmt.Errorf("failed to determine relative path: %w", err)
	}

	if strings.HasPrefix(rel, "..") || strings.HasPrefix(rel, "/") {
		return ToolResult{}, fmt.Errorf("safety violation: attempt to remove file outside project root: %s", path)
	}

	err = os.Remove(absTarget)
	if err != nil {
		if os.IsNotExist(err) {
			return ToolResult{}, fmt.Errorf("file does not exist: %s", path)
		}
		return ToolResult{}, fmt.Errorf("failed to remove file: %w", err)
	}

	fullMsg := fmt.Sprintf("Successfully removed %s", path)
	return ToolResult{
		FullResult: fullMsg,
		Summary:    "Removed",
	}, nil
}

func (t *RemoveFileTool) JSONSchema() map[string]interface{} {
	return map[string]interface{}{
		"name":        "remove_file",
		"description": "Removes a file from the filesystem. Only files within the current project directory can be removed.",
		"parameters": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"path": map[string]interface{}{"type": "string"},
			},
			"required": []string{"path"},
		},
	}
}
