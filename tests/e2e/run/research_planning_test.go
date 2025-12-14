//go:build e2e

package run_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestResearchAndPlanning(t *testing.T) {
	skipIfNoE2E(t)

	if testing.Short() {
		t.Skip("Skipping slow E2E test in short mode")
	}

	tmpDir := t.TempDir()

	// Create sample Go project with auth
	setupAuthProject(t, tmpDir)

	t.Run("Research authentication system", func(t *testing.T) {
		output := runChuResearch(t, tmpDir, "How does the authentication system work?", 5*time.Minute)

		if !strings.Contains(strings.ToLower(output), "auth") {
			t.Errorf("Expected 'auth' in output, got: %s", output)
		}

		t.Logf("✓ Researched authentication")
	})

	t.Run("Research user roles and permissions", func(t *testing.T) {
		output := runChuResearch(t, tmpDir, "Explain the user roles and permission system", 5*time.Minute)

		// Just verify it runs successfully
		if len(output) == 0 {
			t.Error("Expected some output for roles research")
		}

		t.Logf("✓ Researched roles and permissions")
	})

	t.Run("Create plan for adding JWT tokens", func(t *testing.T) {
		output := runChuPlan(t, tmpDir, "Add JWT token-based authentication while keeping existing auth", 5*time.Minute)

		// Verify plan was created
		if len(output) == 0 {
			t.Error("Expected some output for JWT plan")
		}

		t.Logf("✓ Created JWT plan")
	})

	t.Run("Plan for adding OAuth integration", func(t *testing.T) {
		output := runChuPlan(t, tmpDir, "Integrate OAuth2 for third-party login", 5*time.Minute)

		// Verify plan was created
		if len(output) == 0 {
			t.Error("Expected some output for OAuth plan")
		}

		t.Logf("✓ Created OAuth plan")
	})

	t.Logf("✓ All research and planning tests passed")
}

func setupAuthProject(t *testing.T, tmpDir string) {
	t.Helper()

	userCode := `package main

type User struct {
	ID       int
	Email    string
	Password string
	Role     string
}

func (u *User) HasPermission(action string) bool {
	if u.Role == "admin" {
		return true
	}
	return action == "read"
}
`

	authCode := `package main

type AuthService struct {
	users map[int]*User
}

func (a *AuthService) Authenticate(email, password string) (*User, error) {
	for _, user := range a.users {
		if user.Email == email && user.Password == password {
			return user, nil
		}
	}
	return nil, ErrInvalidCredentials
}
`

	goModContent := `module ecommerce

go 1.21
`

	files := map[string]string{
		"user.go": userCode,
		"auth.go": authCode,
		"go.mod":  goModContent,
	}

	for name, content := range files {
		path := filepath.Join(tmpDir, name)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create %s: %v", name, err)
		}
	}
}

func runChuResearch(t *testing.T, dir string, query string, timeout time.Duration) string {
	t.Helper()

	cmd := exec.Command("gptcode", "research", query)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GPTCODE_TELEMETRY=false")

	done := make(chan struct{})
	var output []byte
	var cmdErr error

	go func() {
		output, cmdErr = cmd.CombinedOutput()
		close(done)
	}()

	select {
	case <-done:
		if cmdErr != nil {
			t.Logf("Command output:\n%s", string(output))
			t.Fatalf("chu research failed: %v", cmdErr)
		}
	case <-time.After(timeout):
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
		t.Fatalf("chu research exceeded timeout of %s", timeout)
	}

	return string(output)
}

func runChuPlan(t *testing.T, dir string, task string, timeout time.Duration) string {
	t.Helper()

	cmd := exec.Command("gptcode", "plan", task)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GPTCODE_TELEMETRY=false")

	done := make(chan struct{})
	var output []byte
	var cmdErr error

	go func() {
		output, cmdErr = cmd.CombinedOutput()
		close(done)
	}()

	select {
	case <-done:
		if cmdErr != nil {
			t.Logf("Command output:\n%s", string(output))
			t.Fatalf("chu plan failed: %v", cmdErr)
		}
	case <-time.After(timeout):
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
		t.Fatalf("chu plan exceeded timeout of %s", timeout)
	}

	return string(output)
}
