package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ListFilesTool implements listing files in a directory with support for patterns and recursion.
type ListFilesTool struct {
	BaseDir string // empty = cwd
}

func (t *ListFilesTool) Name() string { return "list_files" }

func (t *ListFilesTool) Execute(args map[string]interface{}) (ToolResult, error) {
	dir := "."
	if t.BaseDir != "" {
		dir = t.BaseDir
	}
	if val, ok := args["path"]; ok {
		if d, isStr := val.(string); isStr {
			dir = d
		}
	}

	pattern := ""
	if val, ok := args["pattern"]; ok {
		if p, isStr := val.(string); isStr {
			pattern = p
		}
	}

	recursive := false
	if val, ok := args["recursive"]; ok {
		if r, isBool := val.(bool); isBool {
			recursive = r
		}
	}

	var files []string

	matchFile := func(relPath string, name string) bool {
		if pattern == "" {
			return true
		}
		// 1. Check if pattern matches the full relative path
		if matched, _ := filepath.Match(pattern, relPath); matched {
			return true
		}
		// 2. Check if pattern matches the base name (e.g., *.go)
		if matched, _ := filepath.Match(pattern, name); matched {
			return true
		}
		return false
	}

	if recursive {
		err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if path == dir {
				return nil
			}

			rel, err := filepath.Rel(dir, path)
			if err != nil {
				return err
			}

			name := d.Name()
			if d.IsDir() {
				name += "/"
			}

			if matchFile(rel, name) {
				files = append(files, rel)
			}
			return nil
		})
		if err != nil {
			return ToolResult{}, fmt.Errorf("failed to walk directory: %w", err)
		}
	} else {
		entries, err := os.ReadDir(dir)
		if err != nil {
			return ToolResult{}, fmt.Errorf("failed to list files: %w", err)
		}
		for _, entry := range entries {
			name := entry.Name()
			if entry.IsDir() {
				name += "/"
			}
			if matchFile(".", name) {
				files = append(files, name)
			}
		}
	}

	var fullResult strings.Builder
	for _, f := range files {
		fullResult.WriteString(fmt.Sprintf("- %s\n", f))
	}

	summary := fmt.Sprintf("%d items found", len(files))
	if len(files) == 0 {
		summary = "0 items found"
	}

	return ToolResult{
		FullResult: fullResult.String(),
		Summary:    summary,
	}, nil
}

func (t *ListFilesTool) JSONSchema() map[string]interface{} {
	return map[string]interface{}{
		"name":        "list_files",
		"description": "Lists files and directories in a given path with support for patterns and recursion.",
		"parameters": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"path":      map[string]interface{}{"type": "string"},
				"pattern":   map[string]interface{}{"type": "string"},
				"recursive": map[string]interface{}{"type": "boolean"},
			},
		},
	}
}
