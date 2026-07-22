package tools

import (
	"fmt"
	"os/exec"
)

// GitAddArgs defines the arguments for the git_add tool.
type GitAddArgs struct {
	Files []string `json:"files"`
}

// GitAddTool implements the Tool interface for performing git add commands.
type GitAddTool struct{}

// Name returns the name of the tool.
func (t *GitAddTool) Name() string {
	return "git_add"
}

// JSONSchema returns the JSON schema for the tool's arguments and description.
func (t *GitAddTool) JSONSchema() map[string]interface{} {
	return map[string]interface{}{
		"description": "Adds files to the git staging area. If files is empty, it adds all changes (git add .).",
		"parameters": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"files": map[string]interface{}{
					"type":        "array",
					"items":       map[string]interface{}{"type": "string"},
					"description": "List of files to add. If empty, all changes are added.",
				},
			},
		},
	}
}

// Execute performs the git add command.
func (t *GitAddTool) Execute(args map[string]interface{}) (ToolResult, error) {
	var gitAddArgs GitAddArgs
	if err := mapToStruct(args, &gitAddArgs); err != nil {
		return ToolResult{FullResult: fmt.Sprintf("Error unmarshaling arguments: %v", err)}, err
	}

	cmdArgs := []string{"add"}
	if len(gitAddArgs.Files) > 0 {
		cmdArgs = append(cmdArgs, gitAddArgs.Files...)
	} else {
		cmdArgs = append(cmdArgs, ".")
	}

	cmd := exec.Command("git", cmdArgs...)
	output, err := cmd.CombinedOutput()
	resStr := string(output)

	if err != nil {
		return ToolResult{
			FullResult:         resStr,
			Summary:            "Git add failed",
			RequiresFullOutput: true,
		}, fmt.Errorf("git add failed: %w", err)
	}

	summary := "Added files to staging"
	if len(gitAddArgs.Files) == 0 {
		summary = "Added all changes to staging"
	} else if len(gitAddArgs.Files) == 1 {
		summary = fmt.Sprintf("Added %s to staging", gitAddArgs.Files[0])
	} else {
		summary = fmt.Sprintf("Added %d files to staging", len(gitAddArgs.Files))
	}

	return ToolResult{
		FullResult: resStr,
		Summary:    summary,
	}, nil
}
