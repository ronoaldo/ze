package tools

import (
	"fmt"
	"os/exec"
	"strings"
)

// GoTool implements consolidated Go tools: doc and test.
type GoTool struct {
	BaseDir string // empty = cwd
}

func (t *GoTool) Name() string { return "go_tool" }

func (t *GoTool) Execute(args map[string]interface{}) (ToolResult, error) {
	var ga GoToolArgs
	if err := mapToStruct(args, &ga); err != nil {
		return ToolResult{FullResult: fmt.Sprintf("Error unmarshaling arguments: %v", err)}, err
	}

	workDir := "."
	if t.BaseDir != "" {
		workDir = t.BaseDir
	}

	switch ga.Action {
	case "doc":
		return t.executeDoc(workDir, ga.Package)
	case "test":
		return t.executeTest(workDir, ga.Path)
	default:
		return ToolResult{}, fmt.Errorf("unsupported go action: %s", ga.Action)
	}
}

func (t *GoTool) executeDoc(dir string, pkg string) (ToolResult, error) {
	if pkg == "" {
		return ToolResult{}, fmt.Errorf("missing 'package' argument")
	}

	var out strings.Builder
	if pkg == "all" {
		// Feature: Run go list and go doc -all for each package
		cmdList := exec.Command("go", "list", "./...")
		cmdList.Dir = dir
		listOut, err := cmdList.Output()
		if err != nil {
			return ToolResult{}, fmt.Errorf("failed to run go list: %w", err)
		}

		packages := strings.Split(strings.TrimSpace(string(listOut)), "\n")
		for _, p := range packages {
			if p == "" {
				continue
			}
			cmdDoc := exec.Command("go", "doc", "-all", p)
			cmdDoc.Dir = dir
			docOut, err := cmdDoc.Output()
			if err != nil {
				continue
			}
			out.WriteString(string(docOut))
			out.WriteString("\n")
		}
	} else {
		cmdDoc := exec.Command("go", "doc", pkg)
		cmdDoc.Dir = dir
		docOut, err := cmdDoc.CombinedOutput()
		if err != nil {
			return ToolResult{}, fmt.Errorf("failed to run go doc: %w (output: %s)", err, string(docOut))
		}
		out.Write(docOut)
	}

	return ToolResult{
		FullResult: out.String(),
		Summary:    "",
	}, nil
}

func (t *GoTool) executeTest(dir string, path string) (ToolResult, error) {
	if path == "" {
		path = "./..."
	}

	cmd := exec.Command("go", "test", "-count=1", path)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()

	const maxOutputSize = 51200 // 50KB limit
	outputStr := string(output)
	if len(outputStr) > maxOutputSize {
		outputStr = outputStr[:maxOutputSize] + "... [TRUNCATED DUE TO SIZE]"
	}

	if err != nil {
		return ToolResult{
			FullResult:         outputStr,
			Summary:            "failed",
			RequiresFullOutput: true,
		}, nil
	}

	return ToolResult{
		FullResult:         outputStr,
		Summary:            "passed",
		RequiresFullOutput: false,
	}, nil
}

func (t *GoTool) JSONSchema() map[string]interface{} {
	return map[string]interface{}{
		"name":        "go_tool",
		"description": "Consolidated tool for Go operations: doc and test.",
		"parameters": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"action":  map[string]interface{}{"type": "string", "enum": []string{"doc", "test"}},
				"package": map[string]interface{}{"type": "string"},
				"path":    map[string]interface{}{"type": "string"},
			},
		},
	}
}
