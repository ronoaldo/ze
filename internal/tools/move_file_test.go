package tools

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMoveFileTool(t *testing.T) {
	tmpDir := t.TempDir()

	src := filepath.Join(tmpDir, "old.txt")
	dst := filepath.Join(tmpDir, "new.txt")

	if err := os.WriteFile(src, []byte("old content"), 0644); err != nil {
		t.Fatal(err)
	}

	moveTool := &MoveFileTool{}
	res, err := moveTool.Execute(map[string]interface{}{"old_path": src, "new_path": dst})
	if err != nil {
		t.Fatal(err)
	}
	if res.Summary == "Move failed" {
		t.Error("move failed unexpectedly")
	}
	if _, err := os.Stat(dst); os.IsNotExist(err) {
		t.Error("destination file does not exist")
	}

	// Verify original file is gone
	if _, err := os.Stat(src); !os.IsNotExist(err) {
		t.Error("original file still exists")
	}

	// 2. Test error: source does not exist
	errSrc := filepath.Join(tmpDir, "nonexistent.txt")
	errDst := filepath.Join(tmpDir, "target.txt")
	res, err = moveTool.Execute(map[string]interface{}{"old_path": errSrc, "new_path": errDst})
	if err == nil {
		t.Error("expected error for non-existent source, but got none")
	}
}
