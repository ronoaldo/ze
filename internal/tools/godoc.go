package tools

import (
	"fmt"
	"os/exec"
)

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
