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
func (t *EditFileTool) Execute(args map[string]interface{}) (ToolResult, error) {
	var a EditArgs
	if err := mapToStruct(args, &a); err != nil {
		return ToolResult{}, fmt.Errorf("invalid arguments: %w", err)
	}
	path := a.Path
	if t.BaseDir != "" {
		path = filepath.Join(t.BaseDir, path)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return ToolResult{}, fmt.Errorf("failed to read file: %w", err)
	}

	currentContent := string(content)
	var totalDeletions, totalAdditions int

	for i, edit := range a.Edits {
		count := strings.Count(currentContent, edit.OldString)
		if count == 0 {
			return ToolResult{}, fmt.Errorf("edit %d: oldString not found", i)
		}
		if count > 1 && !edit.ReplaceAll {
			return ToolResult{}, fmt.Errorf("edit %d: oldString found multiple times, use replaceAll: true", i)
		}

		numReplacements := count
		if !edit.ReplaceAll {
			numReplacements = 1
		}

		if edit.ReplaceAll {
			currentContent = strings.ReplaceAll(currentContent, edit.OldString, edit.NewString)
		} else {
			currentContent = strings.Replace(currentContent, edit.OldString, edit.NewString, 1)
		}

		diff := (len(edit.OldString) - len(edit.NewString)) * numReplacements
		if diff > 0 {
			totalDeletions += diff
		} else if diff < 0 {
			totalAdditions += -diff
		}
	}

	newContent := currentContent
	err = os.WriteFile(path, []byte(newContent), 0644)
	if err != nil {
		return ToolResult{}, fmt.Errorf("failed to write file: %w", err)
	}

	summary := ""
	if totalDeletions > 0 || totalAdditions > 0 {
		summary = fmt.Sprintf("[+%d, -%d]", totalAdditions, totalDeletions)
	} else {
		summary = "no changes"
	}

	return ToolResult{
		FullResult: fmt.Sprintf("Successfully applied %d edits to %s", len(a.Edits), path),
		Summary:    summary,
	}, nil
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
