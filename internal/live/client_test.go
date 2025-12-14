package live

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadWriteContextFile(t *testing.T) {
	// Create temp .gptcode directory
	tmpDir := t.TempDir()
	gptcodeDir := filepath.Join(tmpDir, ".gptcode", "context")
	if err := os.MkdirAll(gptcodeDir, 0755); err != nil {
		t.Fatalf("Failed to create context dir: %v", err)
	}

	// Change to temp directory
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	// Test writing
	content := "# Test Context\n\nThis is a test."
	if err := WriteContextFile("shared", content); err != nil {
		t.Errorf("WriteContextFile failed: %v", err)
	}

	// Test reading
	read, err := ReadContextFile("shared")
	if err != nil {
		t.Errorf("ReadContextFile failed: %v", err)
	}

	if read != content {
		t.Errorf("Content mismatch: got %q, want %q", read, content)
	}
}

func TestReadAllContext(t *testing.T) {
	// Create temp .gptcode directory
	tmpDir := t.TempDir()
	gptcodeDir := filepath.Join(tmpDir, ".gptcode", "context")
	if err := os.MkdirAll(gptcodeDir, 0755); err != nil {
		t.Fatalf("Failed to create context dir: %v", err)
	}

	// Write test files
	os.WriteFile(filepath.Join(gptcodeDir, "shared.md"), []byte("shared content"), 0644)
	os.WriteFile(filepath.Join(gptcodeDir, "next.md"), []byte("next content"), 0644)
	os.WriteFile(filepath.Join(gptcodeDir, "roadmap.md"), []byte("roadmap content"), 0644)

	// Change to temp directory
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	// Test reading all
	shared, next, roadmap, err := ReadAllContext()
	if err != nil {
		t.Errorf("ReadAllContext failed: %v", err)
	}

	if shared != "shared content" {
		t.Errorf("Shared mismatch: got %q", shared)
	}
	if next != "next content" {
		t.Errorf("Next mismatch: got %q", next)
	}
	if roadmap != "roadmap content" {
		t.Errorf("Roadmap mismatch: got %q", roadmap)
	}
}

func TestGetAgentID(t *testing.T) {
	id := GetAgentID()
	if id == "" {
		t.Error("GetAgentID returned empty string")
	}
	t.Logf("Agent ID: %s", id)
}

func TestGetDashboardURL(t *testing.T) {
	// Test default
	url := GetDashboardURL()
	if url != "https://live.gptcode.app" {
		t.Errorf("Expected default URL, got %s", url)
	}

	// Test env override
	os.Setenv("GPTCODE_LIVE_URL", "http://localhost:4000")
	url = GetDashboardURL()
	if url != "http://localhost:4000" {
		t.Errorf("Expected env URL, got %s", url)
	}
	os.Unsetenv("GPTCODE_LIVE_URL")
}

func TestNewClient(t *testing.T) {
	client := NewClient("https://live.gptcode.app", "test-agent")
	if client == nil {
		t.Fatal("NewClient returned nil")
	}
	if client.agentID != "test-agent" {
		t.Errorf("AgentID mismatch: got %s", client.agentID)
	}
	if client.url != "https://live.gptcode.app" {
		t.Errorf("URL mismatch: got %s", client.url)
	}
}

func TestFindGPTCodeDir(t *testing.T) {
	// Create temp directory with .gptcode
	tmpDir := t.TempDir()
	gptcodeDir := filepath.Join(tmpDir, ".gptcode")
	if err := os.MkdirAll(gptcodeDir, 0755); err != nil {
		t.Fatalf("Failed to create .gptcode: %v", err)
	}

	// Create nested subdirectory
	subDir := filepath.Join(tmpDir, "a", "b", "c")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdir: %v", err)
	}

	// Change to nested dir
	oldWd, _ := os.Getwd()
	os.Chdir(subDir)
	defer os.Chdir(oldWd)

	// Should find .gptcode in parent
	found, err := findGPTCodeDir()
	if err != nil {
		t.Errorf("findGPTCodeDir failed: %v", err)
	}

	// Resolve symlinks for comparison (macOS /var -> /private/var)
	foundResolved, _ := filepath.EvalSymlinks(found)
	expectedResolved, _ := filepath.EvalSymlinks(gptcodeDir)
	if foundResolved != expectedResolved {
		t.Errorf("Found wrong dir: got %s, want %s", found, gptcodeDir)
	}
}
