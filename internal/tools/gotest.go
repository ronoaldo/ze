package tools

import (
	"fmt"
	"os/exec"
)

// GoTestArgs defines the arguments for the GoTestTool.
type GoTestArgs struct {
	Path string `json:"path"`
}

// GoTestTool implements running Go tests.
type GoTestTool struct{}

func (t *GoTestTool) Name() string { return "go_test" }
func (t *GoTestTool) Execute(args map[string]interface{}) (string, error) {
	var a GoTestArgs
	if err := mapToStruct(args, &a); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	path := a.Path
	if path == "" {
		path = "./..."
	}

	cmd := exec.Command("go", "test", "-count=1", path)
	output, err := cmd.CombinedOutput()

	const maxOutputSize = 51200 // 50KB limit to prevent memory explosion in agent history
	outputStr := string(output)
	if len(outputStr) > maxOutputSize {
		outputStr = outputStr[:maxOutputSize] + "... [TRUNCATED DUE TO SIZE]"
	}

	if err != nil {
		return outputStr, fmt.Errorf("go test failed: %w", err)
	}

	return outputStr, nil
}
func (t *GoTestTool) JSONSchema() map[string]interface{} {
	return map[string]interface{}{
		"name":        "go_test",
		"description": "Runs Go tests in a given path. Defaults to current project's all packages.",
		"parameters": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"path": map[string]interface{}{"type": "string"},
			},
		},
	}
}
