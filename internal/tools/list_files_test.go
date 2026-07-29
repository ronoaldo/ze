package tools

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setupListFilesEnv(t *testing.T, dir string) {
	t.Helper()

	// Create structure
	// tmpDir/
	//   file1.txt
	//   file2.go
	//   subdir/
	//     file3.go
	//     file4.go
	//   another/
	//     file5.txt

	if err := os.WriteFile(filepath.Join(dir, "file1.txt"), []byte("a"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "file2.go"), []byte("b"), 0644); err != nil {
		t.Fatal(err)
	}

	subDir := filepath.Join(dir, "subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(subDir, "file3.go"), []byte("c"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(subDir, "file4.go"), []byte("d"), 0644); err != nil {
		t.Fatal(err)
	}

	anotherDir := filepath.Join(dir, "another")
	if err := os.Mkdir(anotherDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(anotherDir, "file5.txt"), []byte("e"), 0644); err != nil {
		t.Fatal(err)
	}
}

func TestListFilesTool(t *testing.T) {
	tmpDir := t.TempDir()
	setupListFilesEnv(t, tmpDir)

	listTool := &ListFilesTool{}

	// 1. Test simple list (non-recursive)
	res, err := listTool.Execute(map[string]interface{}{"path": tmpDir})
	if err != nil {
		t.Fatalf("simple list failed: %v", err)
	}
	if !strings.Contains(res.FullResult, "file1.txt") || !strings.Contains(res.FullResult, "file2.go") || strings.Contains(res.FullResult, "file3.go") {
		t.Errorf("simple list should only show top level: %s", res.FullResult)
	}

	// 2. Test recursive list
	res, err = listTool.Execute(map[string]interface{}{"path": tmpDir, "recursive": true})
	if err != nil {
		t.Fatalf("recursive list failed: %v", err)
	}
	if !strings.Contains(res.FullResult, "file3.go") || !strings.Contains(res.FullResult, "file5.txt") {
		t.Errorf("recursive list failed to find nested files: %s", res.FullResult)
	}

	// 3. Test pattern matching (non-recursive) - only top level
	res, err = listTool.Execute(map[string]interface{}{"path": tmpDir, "pattern": "*.go"})
	if err != nil {
		t.Fatalf("pattern list failed: %v", err)
	}
	if !strings.Contains(res.FullResult, "file2.go") || strings.Contains(res.FullResult, "file3.go") {
		t.Errorf("pattern list (non-recursive) should only match top level: %s", res.FullResult)
	}

	// 4. Test pattern matching (recursive)
	res, err = listTool.Execute(map[string]interface{}{
		"path":      tmpDir,
		"pattern":   "*.go",
		"recursive": true,
	})
	if err != nil {
		t.Fatalf("pattern recursive list failed: %v", err)
	}
	if !strings.Contains(res.FullResult, "file2.go") || !strings.Contains(res.FullResult, "file3.go") || !strings.Contains(res.FullResult, "file4.go") {
		t.Errorf("pattern recursive list failed to find all .go files: %s", res.FullResult)
	}
}
