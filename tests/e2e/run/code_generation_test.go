//go:build e2e

package run_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestCodeGeneration(t *testing.T) {
	skipIfNoE2E(t)

	if testing.Short() {
		t.Skip("Skipping slow E2E test in short mode")
	}

	tmpDir := t.TempDir()

	t.Run("Generate Python script", func(t *testing.T) {
		_ = runChuDo(t, tmpDir, "create a Python script calc.py that adds two numbers", 3*time.Minute)
		
		calcFile := filepath.Join(tmpDir, "calc.py")
		content, err := os.ReadFile(calcFile)
		if err != nil {
			t.Fatalf("Failed to read calc.py: %v", err)
		}
		
		if !strings.Contains(string(content), "def") {
			t.Errorf("Expected 'def' in calc.py, got: %s", string(content))
		}
		
		t.Logf("✓ Generated Python script")
	})

	t.Run("Generate JavaScript module", func(t *testing.T) {
		_ = runChuDo(t, tmpDir, "create utils.js with array utility functions", 3*time.Minute)
		
		jsFile := filepath.Join(tmpDir, "utils.js")
		content, err := os.ReadFile(jsFile)
		if err != nil {
			t.Fatalf("Failed to read utils.js: %v", err)
		}
		
		if !strings.Contains(string(content), "function") {
			t.Errorf("Expected 'function' in utils.js, got: %s", string(content))
		}
		
		t.Logf("✓ Generated JavaScript module")
	})

	t.Run("Generate Go program", func(t *testing.T) {
		_ = runChuDo(t, tmpDir, "create main.go with a hello world HTTP server", 3*time.Minute)
		
		goFile := filepath.Join(tmpDir, "main.go")
		content, err := os.ReadFile(goFile)
		if err != nil {
			t.Fatalf("Failed to read main.go: %v", err)
		}
		
		contentStr := string(content)
		if !strings.Contains(contentStr, "package main") {
			t.Errorf("Expected 'package main' in main.go, got: %s", contentStr)
		}
		if !strings.Contains(contentStr, "http") {
			t.Errorf("Expected 'http' in main.go, got: %s", contentStr)
		}
		
		t.Logf("✓ Generated Go program")
	})

	t.Run("Generate shell script", func(t *testing.T) {
		_ = runChuDo(t, tmpDir, "create backup.sh that backs up files to a directory", 3*time.Minute)
		
		shFile := filepath.Join(tmpDir, "backup.sh")
		content, err := os.ReadFile(shFile)
		if err != nil {
			t.Fatalf("Failed to read backup.sh: %v", err)
		}
		
		if !strings.Contains(string(content), "#!/") {
			t.Errorf("Expected shebang in backup.sh, got: %s", string(content))
		}
		
		t.Logf("✓ Generated shell script")
	})

	t.Run("Generate package.json", func(t *testing.T) {
		_ = runChuDo(t, tmpDir, "initialize a Node.js project with package.json", 3*time.Minute)
		
		pkgFile := filepath.Join(tmpDir, "package.json")
		content, err := os.ReadFile(pkgFile)
		if err != nil {
			t.Fatalf("Failed to read package.json: %v", err)
		}
		
		contentStr := string(content)
		if !strings.Contains(contentStr, "name") {
			t.Errorf("Expected 'name' in package.json, got: %s", contentStr)
		}
		if !strings.Contains(contentStr, "version") {
			t.Errorf("Expected 'version' in package.json, got: %s", contentStr)
		}
		
		t.Logf("✓ Generated package.json")
	})

	t.Run("Generate Makefile", func(t *testing.T) {
		_ = runChuDo(t, tmpDir, "create a Makefile with build and test targets", 3*time.Minute)
		
		makeFile := filepath.Join(tmpDir, "Makefile")
		content, err := os.ReadFile(makeFile)
		if err != nil {
			t.Fatalf("Failed to read Makefile: %v", err)
		}
		
		contentStr := string(content)
		if !strings.Contains(contentStr, "build") {
			t.Errorf("Expected 'build' in Makefile, got: %s", contentStr)
		}
		if !strings.Contains(contentStr, "test") {
			t.Errorf("Expected 'test' in Makefile, got: %s", contentStr)
		}
		
		t.Logf("✓ Generated Makefile")
	})

	t.Logf("✓ All code generation tests passed")
}
