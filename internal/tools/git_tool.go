package tools

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// GitTool implements consolidated git operations: diff, add, and commit.
type GitTool struct {
	BaseDir string // empty = cwd
}

func (t *GitTool) Name() string { return "git_tool" }

func (t *GitTool) executeGit(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to run git %s: %w", strings.Join(args, " "), err)
	}
	return string(out), nil
}

func (t *GitTool) Execute(args map[string]interface{}) (ToolResult, error) {
	var gitArgs GitToolArgs
	if err := mapToStruct(args, &gitArgs); err != nil {
		return ToolResult{FullResult: fmt.Sprintf("Error unmarshaling arguments: %v", err)}, err
	}

	workDir := "."
	if t.BaseDir != "" {
		workDir = t.BaseDir
	}

	// Check if it's a git repository
	if _, err := os.Stat(filepath.Join(workDir, ".git")); os.IsNotExist(err) {
		return ToolResult{}, fmt.Errorf("not a git repository")
	}

	switch gitArgs.Action {
	case "diff":
		return t.executeDiff(workDir)
	case "add":
		return t.executeAdd(workDir, gitArgs.Files)
	case "commit":
		return t.executeCommit(workDir, gitArgs.Message)
	default:
		return ToolResult{}, fmt.Errorf("unsupported git action: %s", gitArgs.Action)
	}
}

func (t *GitTool) executeDiff(workDir string) (ToolResult, error) {
	stats := t.getGitStats(workDir)

	// Build Summary
	var parts []string

	if stats.UnstagedAdd+stats.UnstagedDel > 0 {
		parts = append(parts, fmt.Sprintf("+%d/-%d", stats.UnstagedAdd, stats.UnstagedDel))
	}

	if stats.StagedAdd+stats.StagedDel > 0 {
		parts = append(parts, fmt.Sprintf("+%d/-%d staged", stats.StagedAdd, stats.StagedDel))
	}

	if stats.UntrackedCount > 0 {
		parts = append(parts, fmt.Sprintf("%d new file", stats.UntrackedCount))
	}

	summary := ""
	if len(parts) > 0 {
		summary = strings.Join(parts, ", ")
	} else {
		summary = "no changes"
	}

	// Build Full Output for the LLM
	var fullOutput strings.Builder
	fullOutput.WriteString(summary)
	fullOutput.WriteString("\n---\n")

	fullOutput.WriteString("--- GIT STATUS ---\n")
	statusOut, err := t.executeGit(workDir, "status", "--short")
	if err == nil {
		fullOutput.WriteString(statusOut)
	}

	fullOutput.WriteString("\n--- GIT DIFF (unstaged) ---\n")
	diffOut, err := t.executeGit(workDir, "diff")
	if err == nil {
		fullOutput.WriteString(diffOut)
	}

	fullOutput.WriteString("\n--- GIT DIFF (staged) ---\n")
	stagedDiffOut, err := t.executeGit(workDir, "diff", "--cached")
	if err == nil {
		fullOutput.WriteString(stagedDiffOut)
	}

	return ToolResult{
		FullResult: fullOutput.String(),
		Summary:    summary,
	}, nil
}

func (t *GitTool) executeAdd(workDir string, files []string) (ToolResult, error) {
	cmdArgs := []string{"add"}
	if len(files) > 0 {
		cmdArgs = append(cmdArgs, files...)
	} else {
		cmdArgs = append(cmdArgs, ".")
	}

	cmd := exec.Command("git", cmdArgs...)
	cmd.Dir = workDir
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
	if len(files) == 0 {
		summary = "Added all changes to staging"
	} else if len(files) == 1 {
		summary = fmt.Sprintf("Added %s to staging", files[0])
	} else {
		summary = fmt.Sprintf("Added %d files to staging", len(files))
	}

	return ToolResult{
		FullResult: resStr,
		Summary:    summary,
	}, nil
}

func (t *GitTool) executeCommit(workDir string, message string) (ToolResult, error) {
	if message == "" {
		return ToolResult{FullResult: "Error: commit message cannot be empty"}, fmt.Errorf("empty commit message")
	}

	// Execute git commit -m "message"
	cmd := exec.Command("git", "commit", "-m", message)
	cmd.Dir = workDir
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

func (t *GitTool) getGitStats(dir string) GitDiffStats {
	stats := GitDiffStats{}

	// 1. Unstaged changes: git diff --numstat
	unstagedOut, err := t.executeGit(dir, "diff", "--numstat")
	if err == nil {
		t.parseNumstat(unstagedOut, &stats.UnstagedAdd, &stats.UnstagedDel)
	}

	// 2. Staged changes: git diff --numstat --cached
	stagedOut, err := t.executeGit(dir, "diff", "--numstat", "--cached")
	if err == nil {
		t.parseNumstat(stagedOut, &stats.StagedAdd, &stats.StagedDel)
	}

	// 3. Untracked files: git status --short
	statusOut, err := t.executeGit(dir, "status", "--short")
	if err == nil {
		lines := strings.Split(strings.TrimSpace(statusOut), "\n")
		for _, line := range lines {
			if len(line) >= 3 && line[:3] == "?? " {
				stats.UntrackedCount++
			}
		}
	}

	return stats
}

type GitDiffStats struct {
	UnstagedAdd    int
	UnstagedDel    int
	StagedAdd      int
	StagedDel      int
	UntrackedCount int
}

func (t *GitTool) parseNumstat(output string, add *int, del *int) {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}

		a, errA := strconv.Atoi(fields[0])
		if errA != nil {
			a = 0
		}
		d, errD := strconv.Atoi(fields[1])
		if errD != nil {
			d = 0
		}

		*add += a
		*del += d
	}
}

func (t *GitTool) SummarizeArgs(args map[string]interface{}) string {
	action, _ := args["action"].(string)
	if action == "add" {
		if files, ok := args["files"].([]interface{}); ok && len(files) > 0 {
			var fileList []string
			for _, f := range files {
				if s, ok := f.(string); ok {
					fileList = append(fileList, s)
				}
			}
			if len(fileList) == 1 {
				return fmt.Sprintf("add(%s)", fileList[0])
			}
			return fmt.Sprintf("add(%d files)", len(fileList))
		}
		return "add(all)"
	}
	if action == "diff" || action == "" {
		return "diff"
	}
	if action == "commit" {
		if msg, ok := args["message"].(string); ok {
			return fmt.Sprintf("commit('%s')", msg)
		}
		return "commit"
	}
	return action
}

func (t *GitTool) JSONSchema() map[string]interface{} {
	return map[string]interface{}{
		"name":        "git_tool",
		"description": "Consolidated tool for git operations: diff, add, and commit.",
		"parameters": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"action":  map[string]interface{}{"type": "string", "enum": []string{"diff", "add", "commit"}},
				"files":   map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}},
				"message": map[string]interface{}{"type": "string"},
			},
		},
	}
}
