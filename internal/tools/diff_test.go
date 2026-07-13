package tools

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestDiffTool(t *testing.T) {
	t.Run("CleanRepository", func(t *testing.T) {
		tmpDir := t.TempDir()
		runGit := func(args ...string) error {
			cmd := exec.Command("git", args...)
			cmd.Dir = tmpDir
			return cmd.Run()
		}

		runGit("init")
		runGit("config", "user.email", "test@example.com")
		runGit("config", "user.name", "Test User")

		tool := &DiffTool{BaseDir: tmpDir}
		res, err := tool.Execute(nil)
		if err != nil {
			t.Fatalf("DiffTool execution failed: %v", err)
		}
		if !strings.Contains(res.FullResult, "no changes") {
			t.Errorf("expected to contain \"no changes\", got %q", res.FullResult)
		}
	})

	t.Run("UnstagedAndUntracked", func(t *testing.T) {
		tmpSubDir := t.TempDir()
		runGitSub := func(args ...string) error {
			cmd := exec.Command("git", args...)
			cmd.Dir = tmpSubDir
			return cmd.Run()
		}
		runGitSub("init")
		runGitSub("config", "user.email", "test@example.com")
		runGitSub("config", "user.name", "Test User")

		fileName := "test.txt"
		filePath := filepath.Join(tmpSubDir, fileName)
		if err := os.WriteFile(filePath, []byte("initial content\n"), 0644); err != nil {
			t.Fatalf("failed to create file: %v", err)
		}
		runGitSub("add", fileName)
		runGitSub("commit", "-m", "initial commit")

		// Unstaged change
		if err := os.WriteFile(filePath, []byte("unstaged content\n"), 0644); err != nil {
			t.Fatalf("failed to modify file unstaged: %v", err)
		}

		// Untracked file
		untractedFile := filepath.Join(tmpSubDir, "new.txt")
		if err := os.WriteFile(untractedFile, []byte("new file content"), 0644); err != nil {
			t.Fatalf("failed to create new file: %v", err)
		}

		toolSub := &DiffTool{BaseDir: tmpSubDir}
		res, err := toolSub.Execute(nil)
		if err != nil {
			t.Fatalf("DiffTool execution failed: %v", err)
		}

		if !strings.Contains(res.FullResult, "+1/-1, 1 new file") {
			t.Errorf("expected summary to contain \"+1/-1, 1 new file\", got %q", res.FullResult)
		}

	})

	t.Run("StagedOnly", func(t *testing.T) {
		tmpSubDir := t.TempDir()
		runGitSub := func(args ...string) error {
			cmd := exec.Command("git", args...)
			cmd.Dir = tmpSubDir
			return cmd.Run()
		}
		runGitSub("init")
		runGitSub("config", "user.email", "test@example.com")
		runGitSub("config", "user.name", "Test User")

		fileName := "test.txt"
		filePath := filepath.Join(tmpSubDir, fileName)
		if err := os.WriteFile(filePath, []byte("initial content\n"), 0644); err != nil {
			t.Fatalf("failed to create file: %v", err)
		}
		runGitSub("add", fileName)
		runGitSub("commit", "-m", "initial commit")

		// Staged change
		if err := os.WriteFile(filePath, []byte("staged content\n"), 0644); err != nil {
			t.Fatalf("failed to modify file for staging: %v", err)
		}
		runGitSub("add", fileName)

		toolSub := &DiffTool{BaseDir: tmpSubDir}
		res, err := toolSub.Execute(nil)
		if err != nil {
			t.Fatalf("DiffTool execution failed: %v", err)
		}

		if !strings.Contains(res.FullResult, "+1/-1 staged") {
			t.Errorf("expected summary to contain \"+1/-1 staged\", got %q", res.FullResult)
		}
	})

	t.Run("MixedChanges", func(t *testing.T) {
		tmpSubDir := t.TempDir()
		runGitSub := func(args ...string) error {
			cmd := exec.Command("git", args...)
			cmd.Dir = tmpSubDir
			return cmd.Run()
		}
		runGitSub("init")
		runGitSub("config", "user.email", "test@example.com")
		runGitSub("config", "user.name", "Test User")

		fileName := "test.txt"
		filePath := filepath.Join(tmpSubDir, fileName)
		if err := os.WriteFile(filePath, []byte("initial content\n"), 0644); err != nil {
			t.Fatalf("failed to create file: %v", err)
		}
		runGitSub("add", fileName)
		runGitSub("commit", "-m", "initial commit")

		// Staged change
		if err := os.WriteFile(filePath, []byte("staged change\n"), 0644); err != nil {
			t.Fatalf("failed to modify file for staging: %v", err)
		}
		runGitSub("add", fileName)

		// Unstaged change
		if err := os.WriteFile(filePath, []byte("unstaged change\n"), 0644); err != nil {
			t.Fatalf("failed to modify file unstaged: %v", err)
		}

		// Untracked file
		untractedFile := filepath.Join(tmpSubDir, "new.txt")
		if err := os.WriteFile(untractedFile, []byte("new file content"), 0644); err != nil {
			t.Fatalf("failed to create new file: %v", err)
		}

		toolSub := &DiffTool{BaseDir: tmpSubDir}
		res, err := toolSub.Execute(nil)
		if err != nil {
			t.Fatalf("DiffTool execution failed: %v", err)
		}

		if !strings.Contains(res.FullResult, "+1/-1, +1/-1 staged, 1 new file") {
			t.Errorf("expected summary to contain \"+1/-1, +1/-1 staged, 1 new file\", got %q", res.FullResult)
		}

	})
}
