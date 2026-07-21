package tools

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// GitDiffStats holds the aggregated counts of git changes.
type GitDiffStats struct {
	UnstagedAdd    int
	UnstagedDel    int
	StagedAdd      int
	StagedDel      int
	UntrackedCount int
}

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

func (t *DiffTool) Execute(args map[string]interface{}) (ToolResult, error) {
	workDir := "."
	if t.BaseDir != "" {
		workDir = t.BaseDir
	}

	// First, check if it's a git repository
	if _, err := os.Stat(filepath.Join(workDir, ".git")); os.IsNotExist(err) {
		return ToolResult{}, fmt.Errorf("not a git repository")
	}

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

func (t *DiffTool) getGitStats(dir string) GitDiffStats {
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

func (t *DiffTool) parseNumstat(output string, add *int, del *int) {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}

		// Handle binary files (represented by '-')
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

func (t *DiffTool) JSONSchema() map[string]interface{} {
	return map[string]interface{}{
		"name":        "diff",
		"description": "Shows all changes in the project (modified, staged, and unstaged files).",
		"parameters": map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
		},
	}
}
