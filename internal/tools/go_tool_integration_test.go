package tools

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func setupIntegrationGoModule(t *testing.T, dir string) {
	t.Helper()

	// Cria um módulo Go básico
	err := runCommand(dir, "go", "mod", "init", "testmodule")
	if err != nil {
		t.Fatalf("failed to run go mod init: %v", err)
	}

	// --- Pacote PASS (mathutils) ---
	pkgPassDir := filepath.Join(dir, "mathutils")
	if err := os.Mkdir(pkgPassDir, 0755); err != nil {
		t.Fatalf("failed to create pkg dir: %v", err)
	}

	code := `package mathutils
func Add(a, b int) int { return a + b }
`
	if err := os.WriteFile(filepath.Join(pkgPassDir, "math.go"), []byte(code), 0644); err != nil {
		t.Fatalf("failed to write code file: %v", err)
	}

	passTest := `package mathutils
import "testing"
func TestAdd(t *testing.T) { if Add(1, 2) != 3 { t.Error("error") } }
`
	if err := os.WriteFile(filepath.Join(pkgPassDir, "math_test.go"), []byte(passTest), 0644); err != nil {
		t.Fatalf("failed to write pass test: %v", err)
	}

	// --- Pacote FAIL (mathutils_fail) ---
	pkgFailDir := filepath.Join(dir, "mathutils_fail")
	if err := os.Mkdir(pkgFailDir, 0755); err != nil {
		t.Fatalf("failed to create pkg dir: %v", err)
	}

	failTest := `package mathutils_fail
import "testing"
func TestFail(t *testing.T) { t.Error("intentional failure") }
`
	if err := os.WriteFile(filepath.Join(pkgFailDir, "fail_test.go"), []byte(failTest), 0644); err != nil {
		t.Fatalf("failed to write fail test: %v", err)
	}
}

func runCommand(dir string, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	return cmd.Run()
}

func TestGoToolIntegration(t *testing.T) {
	tmpDir := t.TempDir()
	setupIntegrationGoModule(t, tmpDir)

	goTool := &GoTool{BaseDir: tmpDir}

	t.Run("Execute Go Test - Success", func(t *testing.T) {
		args := map[string]interface{}{
			"action": "test",
			"path":   "./mathutils",
		}
		res, err := goTool.Execute(args)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res.Summary != "passed" {
			t.Errorf("expected summary 'passed', got %q", res.Summary)
		}
		if !strings.Contains(res.FullResult, "PASS") && !strings.Contains(res.FullResult, "ok") {
			t.Errorf("expected PASS or ok in output, got %q", res.FullResult)
		}
	})

	t.Run("Execute Go Test - Failure", func(t *testing.T) {
		args := map[string]interface{}{
			"action": "test",
			"path":   "./mathutils_fail",
		}
		res, err := goTool.Execute(args)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res.Summary != "failed" {
			t.Errorf("expected summary 'failed', got %q", res.Summary)
		}
		if !res.RequiresFullOutput {
			t.Error("expected RequiresFullOutput to be true for failed tests")
		}
	})

	t.Run("Execute Go Doc - All", func(t *testing.T) {
		args := map[string]interface{}{
			"action":  "doc",
			"package": "mathutils",
		}
		res, err := goTool.Execute(args)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(res.FullResult, "func Add(a, b int) int") {
			t.Errorf("expected function signature in doc, got %q", res.FullResult)
		}
	})
}
