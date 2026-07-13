package tools

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestListFilesTool(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Create some files/dirs
	os.Mkdir(filepath.Join(tmpDir, "subdir"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte("content"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "file2.log"), []byte("content"), 0644)

	tool := &ListFilesTool{BaseDir: tmpDir}
	res, err := tool.Execute(nil)
	if err != nil {
		t.Fatalf("ListFilesTool execution failed: %v", err)
	}

	if !strings.Contains(res.FullResult, "file1.txt") || !strings.Contains(res.FullResult, "subdir/") || !strings.Contains(res.FullResult, "file2.log") {
		t.Errorf("ListFilesTool output missing expected entries: %s", res.FullResult)
	}
	if res.Summary != "3 items" {
		t.Errorf("expected summary '3 items', got %q", res.Summary)
	}
}
