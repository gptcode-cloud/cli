package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show GT CLI status and system information",
	RunE:  runStatus,
}

func init() {
	rootCmd.AddCommand(statusCmd)
	statusCmd.Flags().Bool("json", false, "Output as JSON")
}

func runStatus(cmd *cobra.Command, args []string) error {
	showJSON, _ := cmd.Flags().GetBool("json")

	status := collectStatus()

	if showJSON {
		data, _ := json.MarshalIndent(status, "", "  ")
		fmt.Println(string(data))
		return nil
	}

	printStatus(status)
	return nil
}

type SystemStatus struct {
	Version   string `json:"version"`
	GOOS      string `json:"os"`
	GOARCH    string `json:"arch"`
	CPUs      int    `json:"cpus"`
	GoVersion string `json:"go_version"`
	Uptime    string `json:"uptime"`
}

type ConfigStatus struct {
	ConfigFile  string `json:"config_file"`
	Backend     string `json:"backend"`
	Model       string `json:"model"`
	LiveEnabled bool   `json:"live_enabled"`
}

type HealthStatus struct {
	System SystemStatus  `json:"system"`
	Config ConfigStatus  `json:"config"`
	Checks []CheckResult `json:"checks"`
}

type CheckResult struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

func collectStatus() *HealthStatus {
	status := &HealthStatus{
		System: collectSystemStatus(),
		Config: collectConfigStatus(),
		Checks: []CheckResult{},
	}

	status.Checks = append(status.Checks, checkGit())
	status.Checks = append(status.Checks, checkNode())
	status.Checks = append(status.Checks, checkAPIKey())
	status.Checks = append(status.Checks, checkDiskSpace())

	return status
}

var statusStartTime = time.Now()

func collectSystemStatus() SystemStatus {
	return SystemStatus{
		Version:   "0.1.0",
		GOOS:      runtime.GOOS,
		GOARCH:    runtime.GOARCH,
		CPUs:      runtime.NumCPU(),
		GoVersion: runtime.Version(),
		Uptime:    time.Since(statusStartTime).Round(time.Second).String(),
	}
}

func collectConfigStatus() ConfigStatus {
	status := ConfigStatus{
		ConfigFile: "~/.gptcode/setup.yaml",
	}

	home, _ := os.UserHomeDir()
	setupPath := home + "/.gptcode/setup.yaml"

	if _, err := os.Stat(setupPath); err == nil {
		status.Backend = "openrouter"
		status.Model = "gemini-2.5-pro"
	}

	if os.Getenv("GPTCODE_LIVE_URL") != "" {
		status.LiveEnabled = true
	}

	return status
}

func checkGit() CheckResult {
	_, err := exec.Command("git", "version").Output()
	if err != nil {
		return CheckResult{Name: "Git", Status: "fail", Message: "Not found"}
	}
	return CheckResult{Name: "Git", Status: "ok"}
}

func checkNode() CheckResult {
	_, err := exec.Command("node", "--version").Output()
	if err != nil {
		return CheckResult{Name: "Node.js", Status: "warn", Message: "Not found"}
	}
	return CheckResult{Name: "Node.js", Status: "ok"}
}

func checkAPIKey() CheckResult {
	if os.Getenv("OPENROUTER_API_KEY") != "" || os.Getenv("ANTHROPIC_API_KEY") != "" {
		return CheckResult{Name: "API Key", Status: "ok"}
	}
	return CheckResult{Name: "API Key", Status: "warn", Message: "No API key configured"}
}

func checkDiskSpace() CheckResult {
	home, _ := os.UserHomeDir()
	testFile := home + "/.gptcode/.space_test"
	err := os.WriteFile(testFile, []byte("test"), 0644)
	if err == nil {
		os.Remove(testFile)
		return CheckResult{Name: "Disk Space", Status: "ok"}
	}
	return CheckResult{Name: "Disk Space", Status: "warn", Message: "Limited write access"}
}

func printStatus(status *HealthStatus) {
	fmt.Println("╔══════════════════════════════════════════╗")
	fmt.Println("║           GT CLI Status                 ║")
	fmt.Println("╚══════════════════════════════════════════╝")
	fmt.Println()

	fmt.Println("System:")
	fmt.Printf("  OS/Arch:    %s/%s\n", status.System.GOOS, status.System.GOARCH)
	fmt.Printf("  CPUs:       %d\n", status.System.CPUs)
	fmt.Printf("  Go Version: %s\n", status.System.GoVersion)
	fmt.Printf("  Uptime:     %s\n", status.System.Uptime)
	fmt.Println()

	fmt.Println("Configuration:")
	fmt.Printf("  Config:     %s\n", status.Config.ConfigFile)
	fmt.Printf("  Backend:    %s\n", status.Config.Backend)
	fmt.Printf("  Model:      %s\n", status.Config.Model)
	fmt.Printf("  Live:       %v\n", status.Config.LiveEnabled)
	fmt.Println()

	fmt.Println("Health Checks:")
	for _, check := range status.Checks {
		icon := "✓"
		if check.Status == "fail" {
			icon = "✗"
		} else if check.Status == "warn" {
			icon = "⚠"
		}
		msg := ""
		if check.Message != "" {
			msg = " (" + check.Message + ")"
		}
		fmt.Printf("  %s %s%s\n", icon, check.Name, msg)
	}
}
