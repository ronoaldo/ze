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

func (t *DiffTool) executeGit(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to run git %s: %w", strings.Join(args, " "), err)
	}
	return string(out), nil
}

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
	statusOut, err := t.executeGit(workDir, "status", "--short")
	if err != nil {
		return "", err
	}
	output.WriteString(statusOut)

	output.WriteString("\n--- GIT STATS ---\n")
	statOut, err := t.executeGit(workDir, "diff", "--stat")
	if err != nil {
		return "", err
	}
	output.WriteString(statOut)

	output.WriteString("\n--- GIT DIFF (unstaged) ---\n")
	diffOut, err := t.executeGit(workDir, "diff")
	if err != nil {
		return "", err
	}
	output.WriteString(diffOut)

	output.WriteString("\n--- GIT DIFF (staged) ---\n")
	stagedDiffOut, err := t.executeGit(workDir, "diff", "--cached")
	if err != nil {
		return "", err
	}
	output.WriteString(stagedDiffOut)

	output.WriteString("\n--- GIT DIFF STATS (staged) ---\n")
	stagedStatOut, err := t.executeGit(workDir, "diff", "--cached", "--stat")
	if err != nil {
		return "", err
	}
	output.WriteString(stagedStatOut)

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
