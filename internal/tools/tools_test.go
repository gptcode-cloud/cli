package tools

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestProjectMap(t *testing.T) {
	t.Run("basic structure", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "gptcode_test")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(tmpDir)

		os.Mkdir(filepath.Join(tmpDir, "subdir"), 0755)
		os.WriteFile(filepath.Join(tmpDir, "file1.go"), []byte("package main"), 0644)
		os.WriteFile(filepath.Join(tmpDir, "subdir", "file2.go"), []byte("package sub"), 0644)

		call := ToolCall{
			Name: "project_map",
			Arguments: map[string]interface{}{
				"max_depth": float64(3),
			},
		}

		result := ProjectMap(call, tmpDir)
		if result.Error != "" {
			t.Fatalf("ProjectMap failed: %s", result.Error)
		}

		if !strings.Contains(result.Result, "📄 file1.go") {
			t.Error("ProjectMap missing file1.go")
		}
		if !strings.Contains(result.Result, "📂 subdir/") {
			t.Error("ProjectMap missing subdir/")
		}
	})

	t.Run("filters ignored directories", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "gptcode_test_filter")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(tmpDir)

		os.Mkdir(filepath.Join(tmpDir, "node_modules"), 0755)
		os.Mkdir(filepath.Join(tmpDir, "vendor"), 0755)
		os.Mkdir(filepath.Join(tmpDir, ".git"), 0755)
		os.Mkdir(filepath.Join(tmpDir, "src"), 0755)
		os.WriteFile(filepath.Join(tmpDir, "src", "main.go"), []byte("package main"), 0644)
		os.WriteFile(filepath.Join(tmpDir, "node_modules", "lib.js"), []byte("code"), 0644)

		call := ToolCall{
			Name: "project_map",
			Arguments: map[string]interface{}{
				"max_depth": float64(3),
			},
		}

		result := ProjectMap(call, tmpDir)
		if result.Error != "" {
			t.Fatalf("ProjectMap failed: %s", result.Error)
		}

		if strings.Contains(result.Result, "node_modules") {
			t.Error("ProjectMap should filter node_modules")
		}
		if strings.Contains(result.Result, "vendor") {
			t.Error("ProjectMap should filter vendor")
		}
		if strings.Contains(result.Result, ".git") {
			t.Error("ProjectMap should filter .git")
		}
		if !strings.Contains(result.Result, "📂 src/") {
			t.Error("ProjectMap should include src/")
		}
	})

	t.Run("respects max_depth", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "gptcode_test_depth")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(tmpDir)

		os.MkdirAll(filepath.Join(tmpDir, "a", "b", "c"), 0755)
		os.WriteFile(filepath.Join(tmpDir, "a", "b", "c", "deep.go"), []byte("package deep"), 0644)

		call := ToolCall{
			Name: "project_map",
			Arguments: map[string]interface{}{
				"max_depth": float64(2),
			},
		}

		result := ProjectMap(call, tmpDir)
		if result.Error != "" {
			t.Fatalf("ProjectMap failed: %s", result.Error)
		}

		if strings.Contains(result.Result, "deep.go") {
			t.Error("ProjectMap should not include files beyond max_depth")
		}
	})
}

func TestApplyPatch(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "gptcode_patch_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	filePath := filepath.Join(tmpDir, "test.txt")
	content := "line1\nline2\nline3\n"
	os.WriteFile(filePath, []byte(content), 0644)

	t.Run("exact match", func(t *testing.T) {
		call := ToolCall{
			Name: "apply_patch",
			Arguments: map[string]interface{}{
				"path":    "test.txt",
				"search":  "line2\n",
				"replace": "line2_modified\n",
			},
		}

		result := ApplyPatch(call, tmpDir)
		if result.Error != "" {
			t.Fatalf("ApplyPatch failed: %s", result.Error)
		}

		newContent, _ := os.ReadFile(filePath)
		if string(newContent) != "line1\nline2_modified\nline3\n" {
			t.Errorf("Patch not applied correctly. Got: %s", string(newContent))
		}
	})

	os.WriteFile(filePath, []byte(content), 0644)

	t.Run("fuzzy match with whitespace", func(t *testing.T) {
		call := ToolCall{
			Name: "apply_patch",
			Arguments: map[string]interface{}{
				"path":    "test.txt",
				"search":  "  line2  \n",
				"replace": "line2_fuzzy\n",
			},
		}

		result := ApplyPatch(call, tmpDir)
		if result.Error != "" {
			t.Fatalf("Fuzzy match failed: %s", result.Error)
		}

		newContent, _ := os.ReadFile(filePath)
		if !strings.Contains(string(newContent), "line2_fuzzy") {
			t.Errorf("Fuzzy patch not applied. Got: %s", string(newContent))
		}
	})

	t.Run("search not found", func(t *testing.T) {
		call := ToolCall{
			Name: "apply_patch",
			Arguments: map[string]interface{}{
				"path":    "test.txt",
				"search":  "nonexistent\n",
				"replace": "foo\n",
			},
		}

		result := ApplyPatch(call, tmpDir)
		if result.Error == "" {
			t.Error("Expected error for nonexistent search block")
		}
	})

	t.Run("empty search", func(t *testing.T) {
		call := ToolCall{
			Name: "apply_patch",
			Arguments: map[string]interface{}{
				"path":    "test.txt",
				"search":  "",
				"replace": "foo\n",
			},
		}

		result := ApplyPatch(call, tmpDir)
		if result.Error == "" {
			t.Error("Expected error for empty search block")
		}
	})

	t.Run("missing parameters", func(t *testing.T) {
		call := ToolCall{
			Name:      "apply_patch",
			Arguments: map[string]interface{}{},
		}

		result := ApplyPatch(call, tmpDir)
		if result.Error == "" {
			t.Error("Expected error for missing parameters")
		}
	})
}
