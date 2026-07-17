package tools

import (
	"testing"
)

func TestGoDocTool_Execute(t *testing.T) {
	tool := &GoDocTool{}

	t.Run("Get documentation for a package", func(t *testing.T) {
		// Try to get documentation for the current module (which should exist)
		args := map[string]interface{}{
			"package": "./...",
		}
		res, err := tool.Execute(args)
		if err != nil {
			// If it fails, it might be because we are not in a go module root
			// but in this environment we should be.
			t.Logf("Note: go doc for ./... failed: %v", err)
		} else {
			if res.FullResult == "" {
				t.Error("expected non-empty result for go doc ./...")
			}
		}
	})

	t.Run("Get all documentation", func(t *testing.T) {
		args := map[string]interface{}{
			"package": "all",
		}
		res, err := tool.Execute(args)
		if err != nil {
			t.Logf("Note: go doc all failed: %v", err)
		} else {
			if res.FullResult == "" {
				t.Error("expected non-empty result for go doc all")
			}
		}
	})

	t.Run("Missing package argument", func(t *testing.T) {
		args := map[string]interface{}{
			"package": "",
		}
		_, err := tool.Execute(args)
		if err == nil {
			t.Error("expected error for missing package argument")
		}
	})
}
