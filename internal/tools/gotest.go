package tools

import (
	"fmt"
	"os/exec"
)

// GoTestTool implements running Go tests.
type GoTestTool struct{}

func (t *GoTestTool) Name() string { return "go_test" }
func (t *GoTestTool) Execute(args map[string]interface{}) (string, error) {
	cmd := exec.Command("go", "test", "-count=1", "./...")
	output, err := cmd.CombinedOutput()

	const maxOutputSize = 51200 // 50KB limit to prevent memory explosion in agent history
	outputStr := string(output)
	if len(outputStr) > maxOutputSize {
		outputStr = outputStr[:maxOutputSize] + "... [TRUNCATED DUE TO SIZE]"
	}

	if err != nil {
		return fmt.Sprintf("%s\n\n--- DETAILED ERROR OUTPUT ---\n%s", outputStr, outputStr), fmt.Errorf("go test failed: %w", err)
	}

	return outputStr, nil
}
func (t *GoTestTool) JSONSchema() map[string]interface{} {
	return map[string]interface{}{
		"name":        "go_test",
		"description": "Runs all Go tests in the project using 'go test -count=1 ./...'.",
		"parameters": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{},
		},
	}
}
