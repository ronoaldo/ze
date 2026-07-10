package tools

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRemoveFileTool(t *testing.T) {
	// Create a temporary directory to act as our workspace
	tmpDir, err := os.MkdirTemp("", "ze-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Change working directory to tmpDir for the tests
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get wd: %v", err)
	}
	err = os.Chdir(tmpDir)
	if err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}
	defer os.Chdir(oldWd)

	tool := &RemoveFileTool{}

	// Case 1: Successfully remove an existing file
	fileName := "testfile.txt"
	filePath := filepath.Join(tmpDir, fileName)
	err = os.WriteFile(filePath, []byte("hello"), 0644)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	args := map[string]interface{}{
		"path": fileName,
	}

	result, err := tool.Execute(args)
	if err != nil {
		t.Errorf("expected success, got error: %v", err)
	}
	if result == "" {
		t.Error("expected result message, got empty string")
	}

	_, err = os.Stat(filePath)
	if err == nil {
		t.Error("expected file to be deleted, but it still exists")
	}

	// Case 2: Attempt to remove a non-existent file
	args = map[string]interface{}{
		"path": "non_existent.txt",
	}
	_, err = tool.Execute(args)
	if err == nil {
		t.Error("expected error when removing non-existent file, got nil")
	}

	// Case 3: Safety Violation - attempt to remove something outside the directory
	args = map[string]interface{}{
		"path": "../outside.txt",
	}
	_, err = tool.Execute(args)
	if err == nil {
		t.Error("expected safety violation error, got nil")
	} else if !strings.Contains(err.Error(), "safety violation") {
		t.Errorf("expected safety violation error message, got: %v", err)
	}
}
