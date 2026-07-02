package tools

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// Tool defines the interface for all agent tools.
type Tool interface {
	Name() string
	Execute(args map[string]interface{}) (string, error)
}

// FileReadTool implements reading a file.
type FileReadTool struct {
	BaseDir string // empty = cwd
}

func (t *FileReadTool) Name() string { return "read_file" }
func (t *FileReadTool) Execute(args map[string]interface{}) (string, error) {
	path, ok := args["path"].(string)
	if !ok {
		return "", fmt.Errorf("missing 'path' argument")
	}
	if t.BaseDir != "" {
		path = filepath.Join(t.BaseDir, path)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	return string(content), nil
}

// FileWriteTool implements writing a file (overwrites).
type FileWriteTool struct {
	BaseDir string // empty = cwd
}

func (t *FileWriteTool) Name() string { return "write_file" }
func (t *FileWriteTool) Execute(args map[string]interface{}) (string, error) {
	path, ok := args["path"].(string)
	if !ok {
		return "", fmt.Errorf("missing 'path' argument")
	}
	if t.BaseDir != "" {
		path = filepath.Join(t.BaseDir, path)
	}

	content, ok := args["content"].(string)
	if !ok {
		return "", fmt.Errorf("missing 'content' argument")
	}

	err := os.WriteFile(path, []byte(content), 0644)
	if err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	return fmt.Sprintf("Successfully wrote to %s", path), nil
}

// ListFilesTool implements listing files in a directory.
type ListFilesTool struct {
	BaseDir string // empty = cwd
}

func (t *ListFilesTool) Name() string { return "list_files" }
func (t *ListFilesTool) Execute(args map[string]interface{}) (string, error) {
	dir := "."
	if t.BaseDir != "" {
		dir = t.BaseDir
	}
	if val, ok := args["path"]; ok {
		if d, isStr := val.(string); isStr {
			dir = d
		}
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", fmt.Errorf("failed to list files: %w", err)
	}

	var result string
	for _, entry := range entries {
		result += fmt.Sprintf("- %s\n", entry.Name())
	}
	return result, nil
}

// GoDocTool implements inspecting documentation using 'go doc'.
type GoDocTool struct{}

func (t *GoDocTool) Name() string { return "go_doc" }
func (t *GoDocTool) Execute(args map[string]interface{}) (string, error) {
	pkg, ok := args["package"].(string)
	if !ok || pkg == "" {
		return "", fmt.Errorf("missing 'package' argument")
	}

	cmd := exec.Command("go", "doc", pkg)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to run go doc: %w (output: %s)", err, string(output))
	}

	return string(output), nil
}
