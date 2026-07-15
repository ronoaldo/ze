package tools

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestGitCommit(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "ze-git-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Save current working directory and restore it later
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current wd: %v", err)
	}
	defer os.Chdir(oldWd)

	// Change to temp dir
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to chdir to temp dir: %v", err)
	}

	// Helper to run commands in the temp dir
	runCmd := func(args ...string) error {
		cmd := exec.Command("git", args...)
		return cmd.Run()
	}

	// Helper to get output of command in the temp dir
	getCmdOutput := func(args ...string) (string, error) {
		cmd := exec.Command("git", args...)
		out, err := cmd.CombinedOutput()
		if err != nil {
			return string(out), err
		}
		return string(out), nil
	}

	// 1. Initialize git repo
	if err := runCmd("init"); err != nil {
		t.Fatalf("failed to init git: %v", err)
	}

	// Set user info for git to avoid errors in CI/local environments
	exec.Command("git", "config", "user.email", "test@example.com").Run()
	exec.Command("git", "config", "user.name", "Test User").Run()

	tool := &GitCommitTool{}

	t.Run("Success", func(t *testing.T) {
		// Create and stage a file
		filePath := filepath.Join(tmpDir, "test.txt")
		if err := os.WriteFile(filePath, []byte("hello"), 0644); err != nil {
			t.Fatalf("failed to create file: %v", err)
		}
		if err := runCmd("add", "."); err != nil {
			t.Fatalf("failed to add file: %v", err)
		}

		args := map[string]interface{}{
			"message": "feat: initial commit",
		}

		res, err := tool.Execute(args)
		if err != nil {
			t.Fatalf("expected no error, got %v: %s", err, res.FullResult)
		}

		if res.Summary != "Commit successful" {
			t.Errorf("expected summary 'Commit successful', got '%s'", res.Summary)
		}

		// Verify git log
		log, err := getCmdOutput("log", "--format=%s")
		if err != nil {
			t.Fatalf("failed to get git log: %v", err)
		}
		if strings.TrimSpace(log) != "feat: initial commit" {
			t.Errorf("expected commit message 'feat: initial commit', got '%s'", log)
		}
	})

	t.Run("NoChanges", func(t *testing.T) {
		// We already have one commit from the previous test.
		// Let's try to commit again without any new changes.
		args := map[string]interface{}{
			"message": "feat: second commit",
		}

		_, err := tool.Execute(args)
		if err == nil {
			t.Error("expected error when committing with no changes, got nil")
		}
	})

	t.Run("EmptyMessage", func(t *testing.T) {
		args := map[string]interface{}{
			"message": "",
		}

		_, err := tool.Execute(args)
		if err == nil {
			t.Error("expected error when message is empty, got nil")
		}
	})
}
