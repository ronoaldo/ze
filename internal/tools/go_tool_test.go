package tools

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setupGoModule(t *testing.T, dir string) {
	t.Helper()

	pkgDir := filepath.Join(dir, "pkg")
	if err := os.Mkdir(pkgDir, 0755); err != nil {
		t.Fatalf("failed to create pkg dir: %v", err)
	}

	goFile := filepath.Join(pkgDir, "main.go")
	content := `package main
func Hello() string {
	return "world"
}
`
	if err := os.WriteFile(goFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write go file: %v", err)
	}

	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module test\n"), 0644); err != nil {
		t.Fatalf("failed to write go.mod: %v", err)
	}
}

func TestGoTool(t *testing.T) {
	tmpDir := t.TempDir()
	setupGoModule(t, tmpDir)

	goTool := &GoTool{BaseDir: tmpDir}

	// 1. Test doc
	res, err := goTool.Execute(map[string]interface{}{"action": "doc", "package": "./pkg.Hello"})
	if err != nil {
		t.Errorf("go doc failed: %v", err)
	}
	if err == nil && res.FullResult == "" {
		t.Errorf("expected some doc output, got empty")
	}

	// 2. Test test
	res, err = goTool.Execute(map[string]interface{}{"action": "test", "path": "./pkg"})
	if err != nil {
		t.Errorf("go test failed: %v", err)
	}
	if !strings.Contains(res.Summary, "failed") && !strings.Contains(res.Summary, "passed") {
		t.Errorf("unexpected test summary: %s", res.Summary)
	}

	// 3. Test error: invalid action
	_, err = goTool.Execute(map[string]interface{}{"action": "invalid"})
	if err == nil || !strings.Contains(err.Error(), "unsupported go action") {
		t.Errorf("expected 'unsupported go action' error, got: %v", err)
	}

	// 4. Test error: missing package for doc
	_, err = goTool.Execute(map[string]interface{}{"action": "doc"})
	if err == nil || !strings.Contains(err.Error(), "missing 'package' argument") {
		t.Errorf("expected 'missing package' error, got: %v", err)
	}
}
