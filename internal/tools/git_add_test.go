package tools

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestGitAdd(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "ze-git-add-test-*")
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

	// Helper to get git status
	getGitStatus := func() string {
		cmd := exec.Command("git", "status", "--porcelain")
		out, _ := cmd.Output()
		return string(out)
	}

	// 1. Initialize git repo
	if err := runCmd("init"); err != nil {
		t.Fatalf("failed to init git: %v", err)
	}

	// Set user info for git to avoid errors
	exec.Command("git", "config", "user.email", "test@example.com").Run()
	exec.Command("git", "config", "user.name", "Test User").Run()

	tool := &GitAddTool{}

	t.Run("AddSpecificFile", func(t *testing.T) {
		// Create a file
		filePath := filepath.Join(tmpDir, "test1.txt")
		if err := os.WriteFile(filePath, []byte("hello"), 0644); err != nil {
			t.Fatalf("failed to create file: %v", err)
		}

		args := map[string]interface{}{
			"files": []interface{}{"test1.txt"},
		}

		res, err := tool.Execute(args)
		if err != nil {
			t.Fatalf("expected no error, got %v: %s", err, res.FullResult)
		}

		if res.Summary != "Added test1.txt to staging" {
			t.Errorf("expected summary 'Added test1.txt to staging', got '%s'", res.Summary)
		}

		status := getGitStatus()
		// git status --porcelain returns "A <space> filename"
		if !strings.Contains(status, "A ") || !strings.Contains(status, filepath.Base(filePath)) {
			t.Errorf("expected file to be staged, got status: %s", status)
		}
	})

	t.Run("AddAllFiles", func(t *testing.T) {
		// Create another file
		filePath := filepath.Join(tmpDir, "test2.txt")
		if err := os.WriteFile(filePath, []byte("world"), 0644); err != nil {
			t.Fatalf("failed to create file: %v", err)
		}

		args := map[string]interface{}{
			"files": []interface{}{},
		}

		res, err := tool.Execute(args)
		if err != nil {
			t.Fatalf("expected no error, got %v: %s", err, res.FullResult)
		}

		if res.Summary != "Added all changes to staging" {
			t.Errorf("expected summary 'Added all changes to staging', got '%s'", res.Summary)
		}

		status := getGitStatus()
		if !strings.Contains(status, "A ") || !strings.Contains(status, filepath.Base(filePath)) {
			t.Errorf("expected file to be staged, got status: %s", status)
		}
	})

	t.Run("EmptyArgs", func(t *testing.T) {
		args := map[string]interface{}{
			"files": nil,
		}

		res, err := tool.Execute(args)
		if err != nil {
			t.Fatalf("expected no error, got %v: %s", err, res.FullResult)
		}

		if res.Summary != "Added all changes to staging" {
			t.Errorf("expected summary 'Added all changes to staging', got '%s'", res.Summary)
		}
	})
}
