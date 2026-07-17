package tools

import (
	"fmt"
	"os/exec"
)

// GoDocTool implements inspecting documentation using 'go doc'.
type GoDocTool struct{}

func (t *GoDocTool) Name() string { return "go_doc" }
func (t *GoDocTool) Execute(args map[string]interface{}) (ToolResult, error) {
	var a GoDocArgs
	if err := mapToStruct(args, &a); err != nil {
		return ToolResult{}, fmt.Errorf("invalid arguments: %w", err)
	}
	if a.Package == "" {
		return ToolResult{}, fmt.Errorf("missing 'package' argument")
	}

	var cmd *exec.Cmd
	if a.Package == "all" {
		// Feature 4: Run go list and go doc -all for each package
		cmd = exec.Command("sh", "-c", "go list ./... | while read pkg ; do go doc -all $pkg ; done")
	} else {
		cmd = exec.Command("go", "doc", a.Package)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return ToolResult{}, fmt.Errorf("failed to run go doc: %w (output: %s)", err, string(output))
	}

	return ToolResult{
		FullResult: string(output),
		Summary:    "",
	}, nil
}
func (t *GoDocTool) JSONSchema() map[string]interface{} {
	return map[string]interface{}{
		"name":        "go_doc",
		"description": "Retrieves Go documentation for a package using 'go doc'. Use 'all' as package to get all local package docs.",
		"parameters": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"package": map[string]interface{}{"type": "string"},
			},
			"required": []string{"package"},
		},
	}
}
