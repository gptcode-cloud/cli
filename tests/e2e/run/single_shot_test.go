//go:build e2e

package run_test

import (
	"os"
	"testing"
)

func TestE2EConfiguration(t *testing.T) {
	backend := os.Getenv("E2E_BACKEND")
	if backend == "" {
		t.Skip("Skipping E2E test: run via 'chu test e2e'")
	}

	profile := os.Getenv("E2E_PROFILE")
	timeout := os.Getenv("E2E_TIMEOUT")

	t.Logf("E2E Configuration:")
	t.Logf("  Backend: %s", backend)
	t.Logf("  Profile: %s", profile)
	t.Logf("  Timeout: %s", timeout)

	if profile == "" {
		t.Error("E2E_PROFILE not set")
	}

	t.Log("✓ E2E environment configured correctly")
}

func TestChuCommand(t *testing.T) {
	if os.Getenv("E2E_BACKEND") == "" {
		t.Skip("Skipping E2E test: run via 'chu test e2e'")
	}

	// Just verify chu is in PATH
	if _, err := os.Stat("/Users/jadercorrea/.local/share/mise/shims/chu"); err != nil {
		t.Skip("chu not found in expected location")
	}

	t.Log("✓ chu command available")
}
