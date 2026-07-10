package tools

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestDiffTool(t *testing.T) {
	tmpDir := t.TempDir()

	// Initialize git repo
	runGit := func(args ...string) error {
		cmd := exec.Command("git", args...)
		cmd.Dir = tmpDir
		return cmd.Run()
	}

	if err := runGit("init"); err != nil {
		t.Fatalf("failed to init git: %v", err)
	}
	if err := runGit("config", "user.email", "test@example.com"); err != nil {
		t.Fatalf("failed to config git email: %v", err)
	}
	if err := runGit("config", "user.name", "Test User"); err != nil {
		t.Fatalf("failed to config git name: %v", err)
	}

	// 1. Create and commit a file
	fileName := "test.txt"
	filePath := filepath.Join(tmpDir, fileName)
	if err := os.WriteFile(filePath, []byte("initial content\n"), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}
	if err := runGit("add", fileName); err != nil {
		t.Fatalf("failed to add file: %v", err)
	}
	if err := runGit("commit", "-m", "initial commit"); err != nil {
		t.Fatalf("failed to commit: %v", err)
	}

	// 2. Create an unstaged change
	if err := os.WriteFile(filePath, []byte("unstaged content\n"), 0644); err != nil {
		t.Fatalf("failed to modify file: %v", err)
	}

	// 3. Create a staged change
	// First, revert the unstaged change so we can stage a clean change
	if err := os.WriteFile(filePath, []byte("initial content\n"), 0644); err != nil {
		t.Fatalf("failed to revert file: %v", err)
	}
	if err := os.WriteFile(filePath, []byte("staged content\n"), 0644); err != nil {
		t.Fatalf("failed to modify file for staging: %v", err)
	}
	if err := runGit("add", fileName); err != nil {
		t.Fatalf("failed to stage file: %v", err)
	}

	// 4. Create both staged and unstaged changes
	// Current state: fileName is staged with "staged content\n"
	// Now modify it unstaged
	if err := os.WriteFile(filePath, []byte("unstaged content after staged\n"), 0644); err != nil {
		t.Fatalf("failed to modify file for dual changes: %v", err)
	}

	tool := &DiffTool{BaseDir: tmpDir}
	output, err := tool.Execute(nil)
	if err != nil {
		t.Fatalf("DiffTool execution failed: %v", err)
	}

	// Verify output contains markers
	if !strings.Contains(output, "--- GIT STATUS ---") {
		t.Errorf("expected GIT STATUS marker")
	}
	if !strings.Contains(output, "--- GIT DIFF (unstaged) ---") {
		t.Errorf("expected GIT DIFF (unstaged) marker")
	}
	if !strings.Contains(output, "--- GIT DIFF (staged) ---") {
		t.Errorf("expected GIT DIFF (staged) marker")
	}

	// Verify content is present
	if !strings.Contains(output, "staged content") {
		t.Errorf("expected staged content in output")
	}
	if !strings.Contains(output, "unstaged content after staged") {
		t.Errorf("expected unstaged content in output")
	}
}
