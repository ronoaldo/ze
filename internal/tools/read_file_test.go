package tools

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadFileTool(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.txt")
	content := "hello world"
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	tool := &FileReadTool{BaseDir: tmpDir}
	res, err := tool.Execute(map[string]interface{}{
		"path": filepath.Base(filePath),
	})

	if err != nil {
		t.Fatalf("ReadFileTool execution failed: %v", err)
	}
	if res.FullResult != content {
		t.Errorf("expected %q, got %q", content, res.FullResult)
	}
}

func TestReadFileTool_Errors(t *testing.T) {
	tool := &FileReadTool{} // CWD
	_, err := tool.Execute(map[string]interface{}{
		"path": "nonexistent_file_abc_123",
	})
	if err == nil {
		t.Error("expected error for non-existent file, got nil")
	}
}
