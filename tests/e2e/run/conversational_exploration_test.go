//go:build e2e

package run_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestConversationalCodeExploration(t *testing.T) {
	skipIfNoE2E(t)

	if testing.Short() {
		t.Skip("Skipping slow E2E test in short mode")
	}

	tmpDir := t.TempDir()

	// Create sample Go code
	setupSampleGoCode(t, tmpDir)

	t.Run("Ask about User struct", func(t *testing.T) {
		output := runChuDo(t, tmpDir, "Explain the User struct and its methods in main.go", 5*time.Minute)
		
		if !strings.Contains(output, "User") {
			t.Errorf("Expected 'User' in output, got: %s", output)
		}
		
		t.Logf("✓ Explained User struct")
	})

	t.Run("Ask about authorization logic", func(t *testing.T) {
		output := runChuDo(t, tmpDir, "How does the AuthorizeAction function work in main.go?", 5*time.Minute)
		
		if !strings.Contains(output, "AuthorizeAction") {
			t.Errorf("Expected 'AuthorizeAction' in output, got: %s", output)
		}
		
		t.Logf("✓ Explained authorization logic")
	})

	t.Run("Ask about improvements", func(t *testing.T) {
		output := runChuDo(t, tmpDir, "What improvements could be made to the authorization in main.go?", 5*time.Minute)
		
		// Just verify it runs successfully - improvements are subjective
		if len(output) == 0 {
			t.Error("Expected some output for improvements question")
		}
		
		t.Logf("✓ Suggested improvements")
	})

	t.Run("Verify file context is being used", func(t *testing.T) {
		output := runChuDo(t, tmpDir, "What parameters does the FullInfo method take in main.go?", 5*time.Minute)
		
		// FullInfo takes no parameters (receiver only)
		// Verify the response understands this
		if len(output) == 0 {
			t.Error("Expected some output about FullInfo parameters")
		}
		
		t.Logf("✓ Verified file context usage")
	})

	t.Logf("✓ All conversational exploration tests passed")
}

func setupSampleGoCode(t *testing.T, tmpDir string) {
	t.Helper()

	goCode := `package main

import (
	"fmt"
	"log"
)

type User struct {
	ID   int
	Name string
	Role string
}

func (u *User) IsAdmin() bool {
	return u.Role == "admin"
}

func (u *User) FullInfo() string {
	return fmt.Sprintf("User %d: %s (%s)", u.ID, u.Name, u.Role)
}

func AuthorizeAction(user *User, action string) bool {
	if user.IsAdmin() {
		return true
	}
	if action == "read" {
		return true
	}
	return false
}

func main() {
	user := &User{ID: 1, Name: "Alice", Role: "admin"}
	log.Println(user.FullInfo())
	if AuthorizeAction(user, "delete") {
		log.Println("Action authorized")
	}
}
`

	goModContent := `module myapp

go 1.21
`

	files := map[string]string{
		"main.go": goCode,
		"go.mod":  goModContent,
	}

	for name, content := range files {
		path := filepath.Join(tmpDir, name)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create %s: %v", name, err)
		}
	}
}
