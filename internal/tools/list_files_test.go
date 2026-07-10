package tools

import (
	"os"
	"path/filepath"
	"testing"
)

func TestListFilesTool(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Create some files/dirs
	os.Mkdir(filepath.Join(tmpDir, "subdir"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte("content"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "file2.log"), []byte("content"), 0644)

	tool := &ListFilesTool{BaseDir: tmpDir}
	output, err := tool.Execute(nil)
	if err != nil {
		t.Fatalf("ListFilesTool execution failed: %v", err)
	}

	if !contains(output, "file1.txt") || !contains(output, "subdir/") || !contains(output, "file2.log") {
		t.Errorf("ListFilesTool output missing expected entries: %s", output)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || (len(substr) > 0 && (s[:len(substr)] == substr || contains(s[1:], substr))))
}
