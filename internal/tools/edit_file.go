package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

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
	oldLen := len(currentContent)

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

	newContent := currentContent
	newLen := len(newContent)
	diffBytes := newLen - oldLen
	diffStr := ""
	if diffBytes >= 0 {
		diffStr = fmt.Sprintf(" [+%d bytes]", diffBytes)
	} else {
		diffStr = fmt.Sprintf(" [%d bytes]", diffBytes)
	}

	err = os.WriteFile(path, []byte(newContent), 0644)
	if err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	return fmt.Sprintf("Successfully applied %d edits to %s%s", len(a.Edits), path, diffStr), nil
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
