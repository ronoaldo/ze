package tools

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Tool defines the interface for all agent tools.
type Tool interface {
	Name() string
	Execute(args map[string]interface{}) (string, error)
	JSONSchema() map[string]interface{}
}

type FileReadArgs struct {
	Path string `json:"path"`
}

type FileWriteArgs struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

type ListFilesArgs struct {
	Path string `json:"path,omitempty"`
}

type GoDocArgs struct {
	Package string `json:"package"`
}

type Edit struct {
	OldString  string `json:"oldString"`
	NewString  string `json:"newString"`
	ReplaceAll bool   `json:"replaceAll"`
}

type EditArgs struct {
	Path  string  `json:"path"`
	Edits []Edit `json:"edits"`
}

// mapToStruct is a helper to decode map into a struct using JSON tags.
func mapToStruct(m map[string]interface{}, s interface{}) error {
	data, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, s)
}

// FileReadTool implements reading a file.
type FileReadTool struct {
	BaseDir string // empty = cwd
}

func (t *FileReadTool) Name() string { return "read_file" }
func (t *FileReadTool) Execute(args map[string]interface{}) (string, error) {
	var a FileReadArgs
	if err := mapToStruct(args, &a); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}
	path := a.Path
	if t.BaseDir != "" {
		path = filepath.Join(t.BaseDir, path)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	return string(content), nil
}
func (t *FileReadTool) JSONSchema() map[string]interface{} {
	return map[string]interface{}{
		"name":        "read_file",
		"description": "Reads the content of a file.",
		"parameters": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"path": map[string]interface{}{"type": "string"},
			},
			"required": []string{"path"},
		},
	}
}

// FileWriteTool implements writing a file (overwrites).
type FileWriteTool struct {
	BaseDir string // empty = cwd
}

func (t *FileWriteTool) Name() string { return "write_file" }
func (t *FileWriteTool) Execute(args map[string]interface{}) (string, error) {
	var a FileWriteArgs
	if err := mapToStruct(args, &a); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}
	path := a.Path
	if t.BaseDir != "" {
		path = filepath.Join(t.BaseDir, path)
	}

	err := os.WriteFile(path, []byte(a.Content), 0644)
	if err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	return fmt.Sprintf("Successfully wrote to %s", path), nil
}
func (t *FileWriteTool) JSONSchema() map[string]interface{} {
	return map[string]interface{}{
		"name":        "write_file",
		"description": "Writes content to a file, overwriting if it exists.",
		"parameters": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"path":    map[string]interface{}{"type": "string"},
				"content": map[string]interface{}{"type": "string"},
			},
			"required": []string{"path", "content"},
		},
	}
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
func (t *ListFilesTool) JSONSchema() map[string]interface{} {
	return map[string]interface{}{
		"name":        "list_files",
		"description": "Lists files and directories in a given path.",
		"parameters": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"path": map[string]interface{}{"type": "string"},
			},
		},
	}
}

// GoDocTool implements inspecting documentation using 'go doc'.
type GoDocTool struct{}

func (t *GoDocTool) Name() string { return "go_doc" }
func (t *GoDocTool) Execute(args map[string]interface{}) (string, error) {
	var a GoDocArgs
	if err := mapToStruct(args, &a); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}
	if a.Package == "" {
		return "", fmt.Errorf("missing 'package' argument")
	}

	cmd := exec.Command("go", "doc", a.Package)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to run go doc: %w (output: %s)", err, string(output))
	}

	return string(output), nil
}
func (t *GoDocTool) JSONSchema() map[string]interface{} {
	return map[string]interface{}{
		"name":        "go_doc",
		"description": "Retrieves Go documentation for a package using 'go doc'.",
		"parameters": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"package": map[string]interface{}{"type": "string"},
			},
			"required": []string{"package"},
		},
	}
}

// EditFileTool implements editing a file with a list of string replacements.
type EditFileTool struct {
	BaseDir string // empty = cwd
}

func (t *EditFileTool) Name() string { return "edit_file" }
func (t *EditFileTool) Execute(args map[string]interface{}) (string, error) {
	var a EditArgs
	if err := mapToStruct(args, &a); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}
	path := a.Path
	if t.BaseDir != "" {
		path = filepath.Join(t.BaseDir, path)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	currentContent := string(content)

	for i, edit := range a.Edits {
		count := strings.Count(currentContent, edit.OldString)
		if count == 0 {
			return "", fmt.Errorf("edit %d: oldString not found", i)
		}
		if count > 1 && !edit.ReplaceAll {
			return "", fmt.Errorf("edit %d: oldString found multiple times, use replaceAll: true", i)
		}

		if edit.ReplaceAll {
			currentContent = strings.ReplaceAll(currentContent, edit.OldString, edit.NewString)
		} else {
			currentContent = strings.Replace(currentContent, edit.OldString, edit.NewString, 1)
		}
	}

	err = os.WriteFile(path, []byte(currentContent), 0644)
	if err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	return fmt.Sprintf("Successfully applied %d edits to %s", len(a.Edits), path), nil
}
func (t *EditFileTool) JSONSchema() map[string]interface{} {
	return map[string]interface{}{
		"name":        "edit_file",
		"description": "Edits a file with a list of string replacements.",
		"parameters": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"path": map[string]interface{}{"type": "string"},
				"edits": map[string]interface{}{
					"type": "array",
					"items": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"oldString":  map[string]interface{}{"type": "string"},
							"newString":  map[string]interface{}{"type": "string"},
							"replaceAll": map[string]interface{}{"type": "boolean"},
						},
						"required": []string{"oldString", "newString"},
					},
				},
			},
			"required": []string{"path", "edits"},
		},
	}
}
