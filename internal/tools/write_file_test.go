package tools

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFileWriteTool_Execute(t *testing.T) {
	tmpDir := t.TempDir()
	tool := &FileWriteTool{BaseDir: tmpDir}

	tests := []struct {
		name    string
		path    string
		content string
	}{
		{
			name:    "write file in root",
			path:    "test.txt",
			content: "hello world",
		},
		{
			name:    "write file in sub-directory",
			path:    "subdir/test.txt",
			content: "nested content",
		},
		{
			name:    "write file in deep sub-directory",
			path:    "a/b/c/deep.txt",
			content: "very deep",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tool.Execute(map[string]interface{}{
				"path":    tt.path,
				"content": tt.content,
			})
			if err != nil {
				t.Fatalf("Execute failed: %v", err)
			}

			fullPath := filepath.Join(tmpDir, tt.path)
			content, err := os.ReadFile(fullPath)
			if err != nil {
				t.Fatalf("failed to read file: %v", err)
			}
			if string(content) != tt.content {
				t.Errorf("expected content %q, got %q", tt.content, string(content))
			}
		})
	}
}

func TestFileWriteTool_Overwrite(t *testing.T) {
	tmpDir := t.TempDir()
	tool := &FileWriteTool{BaseDir: tmpDir}

	path := "overwrite.txt"
	content1 := "original"
	content2 := "new content"

	// Write original
	_, err := tool.Execute(map[string]interface{}{
		"path":    path,
		"content": content1,
	})
	if err != nil {
		t.Fatalf("first write failed: %v", err)
	}

	// Overwrite
	_, err = tool.Execute(map[string]interface{}{
		"path":    path,
		"content": content2,
	})
	if err != nil {
		t.Fatalf("overwrite failed: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(tmpDir, path))
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	if string(content) != content2 {
		t.Errorf("expected content %q, got %q", content2, string(content))
	}
}
