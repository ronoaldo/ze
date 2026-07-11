package tools

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGoDocToolIntegration(t *testing.T) {
	// Setup temp directory
	tmpDir := t.TempDir()

	// Create a sub-directory for the test package
	pkgDir := filepath.Join(tmpDir, "mypkg")
	if err := os.Mkdir(pkgDir, 0755); err != nil {
		t.Fatalf("failed to create pkg dir: %v", err)
	}

	// Create a file in the sub-package
	pkgFile := filepath.Join(pkgDir, "hello.go")
	pkgContent := `package mypkg

// Hello returns a greeting.
func Hello() string {
	return "hello"
}
`
	if err := os.WriteFile(pkgFile, []byte(pkgContent), 0644); err != nil {
		t.Fatalf("failed to write pkg file: %v", err)
	}

	// Create go.mod in pkgDir
	if err := os.WriteFile(filepath.Join(pkgDir, "go.mod"), []byte("module mypkg"), 0644); err != nil {
		t.Fatalf("failed to write go.mod: %v", err)
	}

	tool := &GoDocTool{}
	
	// We need to run the test from the parent directory (the one containing mypkg)
	// and we will pass the relative path to the package.
	
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	args := map[string]interface{}{
		"package": "./mypkg",
	}

	output, err := tool.Execute(args)
	if err != nil {
		t.Fatalf("GoDocTool execution failed: %v", err)
	}

	if !strings.Contains(output, "Hello") {
		t.Errorf("GoDocTool output should contain 'Hello', got: %s", output)
	}
}

func TestGoTestToolIntegration(t *testing.T) {
	// Setup temp directory
	tmpDir := t.TempDir()

	// Create a sub-directory for the test project
	projDir := filepath.Join(tmpDir, "testproject")
	if err := os.Mkdir(projDir, 0755); err != nil {
		t.Fatalf("failed to create project dir: %v", err)
	}

	// Create go.mod
	if err := os.WriteFile(filepath.Join(projDir, "go.mod"), []byte("module testproject"), 0644); err != nil {
		t.Fatalf("failed to write go.mod: %v", err)
	}

	// Create a main.go
	if err := os.WriteFile(filepath.Join(projDir, "main.go"), []byte("package main\nfunc main() {}"), 0644); err != nil {
		t.Fatalf("failed to write main.go: %v", err)
	}

	// Create a test file
	testFile := filepath.Join(projDir, "main_test.go")
	testContent := `package main
import "testing"

func TestMain(t *testing.T) {
	if false {
		t.Error("should not fail")
	}
}

func TestFail(t *testing.T) {
	t.Errorf("intentional failure")
}
`
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	tool := &GoTestTool{}

	// We need to run the tool from the project directory or change Dir in Execute.
	// Currently GoTestTool.Execute uses exec.Command("go", "test", "-count=1", "./...").
	// This works if we change CWD.
	
	oldWd, _ := os.Getwd()
	os.Chdir(projDir)
	defer os.Chdir(oldWd)

	args := map[string]interface{}{
		"path": "./...",
	}
	output, err := tool.Execute(args)

	// It should return an error because TestFail fails.
	if err == nil {
		t.Error("Expected error from failed test, got nil")
	}

	if !strings.Contains(output, "intentional failure") {
		t.Errorf("Expected failure message in output, got: %s", output)
	}
}
