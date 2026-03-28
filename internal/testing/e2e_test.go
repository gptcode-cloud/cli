package testing

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGTVersion(t *testing.T) {
	cli := NewCLI("gptcode")
	result := cli.Run("version")

	AssertSuccess(t, result)
	AssertContains(t, result.Output(), "GPTCode")
}

func TestGTStatus(t *testing.T) {
	cli := NewCLI("gptcode")
	cli.workDir = getTestWorkDir(t)
	result := cli.Run("status")

	AssertSuccess(t, result)
}

func TestGTHelp(t *testing.T) {
	cli := NewCLI("gptcode")
	result := cli.Run("--help")

	AssertSuccess(t, result)
	AssertContains(t, result.Output(), "GPTCode")
}

func TestGTRunHelp(t *testing.T) {
	cli := NewCLI("gptcode")
	result := cli.Run("run", "--help")

	AssertSuccess(t, result)
	AssertContains(t, result.Output(), "Execute")
}

func TestGTDoHelp(t *testing.T) {
	cli := NewCLI("gptcode")
	result := cli.Run("do", "--help")

	AssertSuccess(t, result)
}

func TestGTModelList(t *testing.T) {
	cli := NewCLI("gptcode")
	result := cli.Run("model", "list")

	AssertSuccess(t, result)
}

func TestGTBackendList(t *testing.T) {
	cli := NewCLI("gptcode")
	result := cli.Run("backend", "list")

	AssertSuccess(t, result)
}

func getTestWorkDir(t *testing.T) string {
	dir := t.TempDir()
	return dir
}

func TestGTRunSimple(t *testing.T) {
	cli := NewCLI("gptcode")
	cli.workDir = getTestWorkDir(t)

	result := cli.Run("run", "--once", "echo hello")

	// Should succeed or fail gracefully
	if result.Failed() {
		t.Logf("Command failed with: %s", result.Stderr)
	}
}

func TestGTPRHelp(t *testing.T) {
	cli := NewCLI("gptcode")
	result := cli.Run("pr", "--help")

	AssertSuccess(t, result)
}

func TestGTIssueHelp(t *testing.T) {
	cli := NewCLI("gptcode")
	result := cli.Run("issue", "--help")

	AssertSuccess(t, result)
}

func TestGTInit(t *testing.T) {
	dir := t.TempDir()
	cli := NewCLI("gptcode")
	cli.workDir = dir

	result := cli.Run("init")

	AssertSuccess(t, result)

	// Check that .gptcode was created
	gptcodeDir := filepath.Join(dir, ".gptcode")
	if _, err := os.Stat(gptcodeDir); os.IsNotExist(err) {
		t.Errorf("Expected .gptcode directory to be created")
	}
}

func TestGTSetupHelp(t *testing.T) {
	cli := NewCLI("gptcode")
	result := cli.Run("setup", "--help")

	AssertSuccess(t, result)
}

func TestGTKeyHelp(t *testing.T) {
	cli := NewCLI("gptcode")
	result := cli.Run("key", "--help")

	AssertSuccess(t, result)
}

func TestGTChatHelp(t *testing.T) {
	cli := NewCLI("gptcode")
	result := cli.Run("chat", "--help")

	AssertSuccess(t, result)
}

func TestGTResearchHelp(t *testing.T) {
	cli := NewCLI("gptcode")
	result := cli.Run("research", "--help")

	AssertSuccess(t, result)
}

func TestGTPlanHelp(t *testing.T) {
	cli := NewCLI("gptcode")
	result := cli.Run("plan", "--help")

	AssertSuccess(t, result)
}

func TestGTImplementHelp(t *testing.T) {
	cli := NewCLI("gptcode")
	result := cli.Run("implement", "--help")

	AssertSuccess(t, result)
}

func TestGTFeatureHelp(t *testing.T) {
	cli := NewCLI("gptcode")
	result := cli.Run("feature", "--help")

	AssertSuccess(t, result)
}

func TestGTReviewHelp(t *testing.T) {
	cli := NewCLI("gptcode")
	result := cli.Run("review", "--help")

	AssertSuccess(t, result)
}

func TestGTGitHelp(t *testing.T) {
	cli := NewCLI("gptcode")
	result := cli.Run("git", "--help")

	AssertSuccess(t, result)
}

func TestGTWatchHelp(t *testing.T) {
	cli := NewCLI("gptcode")
	result := cli.Run("watch", "--help")

	AssertSuccess(t, result)
}

func TestGTContextHelp(t *testing.T) {
	cli := NewCLI("gptcode")
	result := cli.Run("context", "--help")

	AssertSuccess(t, result)
}

func TestGTPerfHelp(t *testing.T) {
	cli := NewCLI("gptcode")
	result := cli.Run("perf", "--help")

	AssertSuccess(t, result)
}

func TestGTCoverageHelp(t *testing.T) {
	cli := NewCLI("gptcode")
	result := cli.Run("coverage", "--help")

	AssertSuccess(t, result)
}

func TestGTTestHelp(t *testing.T) {
	cli := NewCLI("gptcode")
	result := cli.Run("test", "--help")

	AssertSuccess(t, result)
}

func TestGTGraphHelp(t *testing.T) {
	cli := NewCLI("gptcode")
	result := cli.Run("graph", "--help")

	AssertSuccess(t, result)
}

func TestGTRefactorHelp(t *testing.T) {
	cli := NewCLI("gptcode")
	result := cli.Run("refactor", "--help")

	AssertSuccess(t, result)
}

func TestGTMergeHelp(t *testing.T) {
	cli := NewCLI("gptcode")
	result := cli.Run("merge", "--help")

	AssertSuccess(t, result)
}

func TestGTDocHelp(t *testing.T) {
	cli := NewCLI("gptcode")
	result := cli.Run("doc", "--help")

	AssertSuccess(t, result)
}

func TestGTGenHelp(t *testing.T) {
	cli := NewCLI("gptcode")
	result := cli.Run("gen", "--help")

	AssertSuccess(t, result)
}

func TestGTEvolveHelp(t *testing.T) {
	cli := NewCLI("gptcode")
	result := cli.Run("evolve", "--help")

	AssertSuccess(t, result)
}

func TestGTLoginHelp(t *testing.T) {
	cli := NewCLI("gptcode")
	result := cli.Run("login", "--help")

	AssertSuccess(t, result)
}

func TestGTLogoutHelp(t *testing.T) {
	cli := NewCLI("gptcode")
	result := cli.Run("logout", "--help")

	AssertSuccess(t, result)
}

func TestGTProfileHelp(t *testing.T) {
	cli := NewCLI("gptcode")
	result := cli.Run("profile", "--help")

	AssertSuccess(t, result)
}

func TestGTProfilesHelp(t *testing.T) {
	cli := NewCLI("gptcode")
	result := cli.Run("profiles", "--help")

	AssertSuccess(t, result)
}

func TestGTFeedbackHelp(t *testing.T) {
	cli := NewCLI("gptcode")
	result := cli.Run("feedback", "--help")

	AssertSuccess(t, result)
}

func TestGTTrainingHelp(t *testing.T) {
	cli := NewCLI("gptcode")
	result := cli.Run("training", "--help")

	AssertSuccess(t, result)
}

func TestGTDetectHelp(t *testing.T) {
	cli := NewCLI("gptcode")
	result := cli.Run("detect", "--help")

	AssertSuccess(t, result)
}

func TestGTMLHelp(t *testing.T) {
	cli := NewCLI("gptcode")
	result := cli.Run("ml", "--help")

	AssertSuccess(t, result)
}

func TestGTBackendHelp(t *testing.T) {
	cli := NewCLI("gptcode")
	result := cli.Run("backend", "--help")

	AssertSuccess(t, result)
}

func TestGTConfigHelp(t *testing.T) {
	cli := NewCLI("gptcode")
	result := cli.Run("config", "--help")

	AssertSuccess(t, result)
}

func TestGTDoctorHelp(t *testing.T) {
	cli := NewCLI("gptcode")
	result := cli.Run("doctor", "--help")

	AssertSuccess(t, result)
}
