package tools

import (
	"strings"
	"testing"
)

func TestGoDocTool(t *testing.T) {
	tool := &GoDocTool{}
	
	// Test with a known package
	args := map[string]interface{}{
		"package": "fmt",
	}
	output, err := tool.Execute(args)
	if err != nil {
		t.Fatalf("GoDocTool execution failed: %v", err)
	}

	if !strings.Contains(output, "fmt") {
		t.Errorf("GoDocTool output should contain 'fmt', got: %s", output)
	}

	// Test with invalid package
	args["package"] = "nonexistent_package_xyz_123"
	_, err = tool.Execute(args)
	if err == nil {
		t.Error("Expected error for nonexistent package, got nil")
	}
}

func TestGoTestTool(t *testing.T) {
	// Removed because it triggers an infinite loop during 'go test ./...'
}
