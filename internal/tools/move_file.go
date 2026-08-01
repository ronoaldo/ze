package tools

import (
	"fmt"
	"os"
	"path/filepath"
)

// MoveFileTool implements the Tool interface for renaming or moving files/directories.
type MoveFileTool struct{}

func (t *MoveFileTool) Name() string { return "move_file" }

func (t *MoveFileTool) SummarizeArgs(args map[string]interface{}) string {
	oldPath, _ := args["old_path"].(string)
	newPath, _ := args["new_path"].(string)
	return fmt.Sprintf("'%s' -> '%s'", oldPath, newPath)
}

func (t *MoveFileTool) Execute(args map[string]interface{}) (ToolResult, error) {
	var mArgs MoveFileArgs
	if err := mapToStruct(args, &mArgs); err != nil {
		return ToolResult{FullResult: fmt.Sprintf("Error unmarshaling arguments: %v", err)}, err
	}

	err := os.Rename(mArgs.OldPath, mArgs.NewPath)
	if err != nil {
		return ToolResult{
			FullResult:         err.Error(),
			Summary:            "Move failed",
			RequiresFullOutput: true,
		}, fmt.Errorf("failed to move %s to %s: %w", mArgs.OldPath, mArgs.NewPath, err)
	}

	return ToolResult{
		FullResult: fmt.Sprintf("Successfully moved/renamed %s to %s", mArgs.OldPath, mArgs.NewPath),
		Summary:    fmt.Sprintf("Moved %s to %s", filepath.Base(mArgs.OldPath), filepath.Base(mArgs.NewPath)),
	}, nil
}

func (t *MoveFileTool) JSONSchema() map[string]interface{} {
	return map[string]interface{}{
		"name":        "move_file",
		"description": "Renames or moves a file or directory.",
		"parameters": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"old_path": map[string]interface{}{"type": "string"},
				"new_path": map[string]interface{}{"type": "string"},
			},
			"required": []string{"old_path", "new_path"},
		},
	}
}
