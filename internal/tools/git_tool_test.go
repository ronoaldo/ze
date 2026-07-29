package tools

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func setupGitRepo(t *testing.T, dir string) {
	t.Helper()
	commands := [][]string{
		{"init"},
		{"config", "user.name", "Test User"},
		{"config", "user.email", "test@example.com"},
		{"config", "commit.gpgsign", "false"},
	}

	for _, args := range commands {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if err := cmd.Run(); err != nil {
			t.Fatalf("failed to run git %s: %v", strings.Join(args, " "), err)
		}
	}
}

func TestGitTool(t *testing.T) {
	tmpDir := t.TempDir()
	setupGitRepo(t, tmpDir)

	gitTool := &GitTool{BaseDir: tmpDir}

	// 1. Test diff (no changes)
	res, err := gitTool.Execute(map[string]interface{}{"action": "diff"})
	if err != nil {
		t.Errorf("diff failed: %v", err)
	}
	if !strings.Contains(res.Summary, "no changes") {
		t.Errorf("expected 'no changes' summary, got %s", res.Summary)
	}

	// 2. Test add and diff (with changes)
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("hello"), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	// Check diff (unstaged)
	res, err = gitTool.Execute(map[string]interface{}{"action": "diff"})
	if err != nil {
		t.Errorf("diff failed: %v", err)
	}
	if !strings.Contains(res.FullResult, "1 new file") {
		t.Errorf("expected '1 new file' in output, got: %s", res.FullResult)
	}

	// Test add
	res, err = gitTool.Execute(map[string]interface{}{"action": "add", "files": []string{"test.txt"}})
	if err != nil {
		t.Errorf("add failed: %v", err)
	}
	if !strings.Contains(res.Summary, "Added test.txt") {
		t.Errorf("expected 'Added test.txt' summary, got %s", res.Summary)
	}

	// Check diff (staged)
	res, err = gitTool.Execute(map[string]interface{}{"action": "diff"})
	if err != nil {
		t.Errorf("diff failed: %v", err)
	}
	if !strings.Contains(res.Summary, "staged") {
		t.Errorf("expected 'staged' summary, got %s", res.Summary)
	}

	// 3. Test commit
	res, err = gitTool.Execute(map[string]interface{}{"action": "commit", "message": "feat: test commit"})
	if err != nil {
		t.Errorf("commit failed: %v", err)
	}
	if res.Summary != "Commit successful" {
		t.Errorf("expected 'Commit successful' summary, got %s", res.Summary)
	}

	// 4. Test error: non-git directory
	errDir := filepath.Join(tmpDir, "not-git")
	os.Mkdir(errDir, 0755)
	gitToolNoRepo := &GitTool{BaseDir: errDir}
	_, err = gitToolNoRepo.Execute(map[string]interface{}{"action": "diff"})
	if err == nil || !strings.Contains(err.Error(), "not a git repository") {
		t.Errorf("expected 'not a git repository' error, got: %v", err)
	}

	// 5. Test error: unsupported action
	_, err = gitTool.Execute(map[string]interface{}{"action": "invalid"})
	if err == nil || !strings.Contains(err.Error(), "unsupported git action") {
		t.Errorf("expected 'unsupported git action' error, got: %v", err)
	}
}
