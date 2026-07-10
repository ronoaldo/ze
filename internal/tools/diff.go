package tools

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// DiffTool implements detecting project changes using 'git diff'.
type DiffTool struct {
	BaseDir string // empty = cwd
}

func (t *DiffTool) Name() string { return "diff" }
func (t *DiffTool) Execute(args map[string]interface{}) (string, error) {
	workDir := "."
	if t.BaseDir != "" {
		workDir = t.BaseDir
	}

	// First, check if it's a git repository
	if _, err := os.Stat(filepath.Join(workDir, ".git")); os.IsNotExist(err) {
		return "", fmt.Errorf("not a git repository")
	}

	var output strings.Builder

	output.WriteString("--- GIT STATUS ---\n")
	statusCmd := exec.Command("git", "status", "--short")
	statusCmd.Dir = workDir
	statusOut, err := statusCmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to run git status: %w", err)
	}
	output.Write(statusOut)

	output.WriteString("\n--- GIT DIFF (unstaged) ---\n")
	diffCmd := exec.Command("git", "diff")
	diffCmd.Dir = workDir
	diffOut, err := diffCmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to run git diff: %w", err)
	}
	output.Write(diffOut)

	output.WriteString("\n--- GIT DIFF (staged) ---\n")
	stagedDiffCmd := exec.Command("git", "diff", "--cached")
	stagedDiffCmd.Dir = workDir
	stagedDiffOut, err := stagedDiffCmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to run git diff --cached: %w", err)
	}
	output.Write(stagedDiffOut)

	return output.String(), nil
}
func (t *DiffTool) JSONSchema() map[string]interface{} {
	return map[string]interface{}{
		"name":        "diff",
		"description": "Shows all changes in the project (modified, staged, and unstaged files).",
		"parameters": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{},
		},
	}
}
