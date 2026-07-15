package tools

import (
	"fmt"
	"os/exec"
)

// GitCommitArgs defines the arguments for the git_commit tool.
type GitCommitArgs struct {
	Message string `json:"message"`
}

// GitCommitTool implements the Tool interface for performing git commits.
type GitCommitTool struct{}

// Name returns the name of the tool.
func (t *GitCommitTool) Name() string {
	return "git_commit"
}

// JSONSchema returns the JSON schema for the tool's arguments and description.
func (t *GitCommitTool) JSONSchema() map[string]interface{} {
	return map[string]interface{}{
		"description": "Commits staged changes to the git repository. CRITICAL: This tool must ONLY be called after the user has explicitly approved the commit and the commit message. Do not use this tool just to generate a message; use it to actually perform the commit.",
		"parameters": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"message": map[string]interface{}{
					"type":        "string",
					"description": "The commit message.",
				},
			},
			"required": []string{"message"},
		},
	}
}

// Execute performs the git commit command.
func (t *GitCommitTool) Execute(args map[string]interface{}) (ToolResult, error) {
	var gitArgs GitCommitArgs
	if err := mapToStruct(args, &gitArgs); err != nil {
		return ToolResult{FullResult: fmt.Sprintf("Error unmarshaling arguments: %v", err)}, err
	}

	if gitArgs.Message == "" {
		return ToolResult{FullResult: "Error: commit message cannot be empty"}, fmt.Errorf("empty commit message")
	}

	// Execute git commit -m "message"
	cmd := exec.Command("git", "commit", "-m", gitArgs.Message)
	output, err := cmd.CombinedOutput()
	resStr := string(output)

	if err != nil {
		return ToolResult{
			FullResult:         resStr,
			Summary:            "Commit failed",
			RequiresFullOutput: true,
		}, fmt.Errorf("git commit failed: %w", err)
	}

	return ToolResult{
		FullResult: resStr,
		Summary:    "Commit successful",
	}, nil
}
