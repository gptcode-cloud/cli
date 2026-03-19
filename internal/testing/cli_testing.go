package testing

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"
)

type CLI struct {
	path    string
	workDir string
}

func NewCLI(path string) *CLI {
	return &CLI{path: path}
}

func (c *CLI) WithWorkDir(dir string) *CLI {
	c.workDir = dir
	return c
}

func (c *CLI) Run(args ...string) *Result {
	return c.RunContext(context.Background(), args...)
}

func (c *CLI) RunContext(ctx context.Context, args ...string) *Result {
	cmd := exec.CommandContext(ctx, c.path, args...)
	cmd.Dir = c.workDir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	start := time.Now()
	err := cmd.Run()
	duration := time.Since(start)

	return &Result{
		Command:  c.path + " " + strings.Join(args, " "),
		Args:     args,
		ExitCode: cmd.ProcessState.ExitCode(),
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		Error:    err,
		Duration: duration,
	}
}

func (c *CLI) RunAsync(args ...string) *AsyncResult {
	cmd := exec.Command(c.path, args...)
	cmd.Dir = c.workDir

	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()

	cmd.Start()

	return &AsyncResult{
		cmd:    cmd,
		stdout: stdout,
		stderr: stderr,
	}
}

type Result struct {
	Command  string
	Args     []string
	ExitCode int
	Stdout   string
	Stderr   string
	Error    error
	Duration time.Duration
}

func (r *Result) Success() bool {
	return r.ExitCode == 0 && r.Error == nil
}

func (r *Result) Failed() bool {
	return !r.Success()
}

func (r *Result) Output() string {
	return r.Stdout + r.Stderr
}

func (r *Result) Contains(s string) bool {
	return strings.Contains(r.Stdout, s) || strings.Contains(r.Stderr, s)
}

type AsyncResult struct {
	cmd    *exec.Cmd
	stdout interface{ Read([]byte) (int, error) }
	stderr interface{ Read([]byte) (int, error) }
}

func (r *AsyncResult) Wait() *Result {
	err := r.cmd.Wait()

	var stdout, stderr bytes.Buffer
	if r.stdout != nil {
		buf := make([]byte, 1024)
		for {
			n, err := r.stdout.Read(buf)
			if n > 0 {
				stdout.Write(buf[:n])
			}
			if err != nil {
				break
			}
		}
	}

	var exitCode int
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		}
	}

	return &Result{
		Command:  r.cmd.Path,
		ExitCode: exitCode,
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		Error:    err,
	}
}

func (r *AsyncResult) Kill() error {
	return r.cmd.Process.Kill()
}

type Snapshot struct {
	name    string
	cli     *CLI
	updates bool
}

func NewSnapshot(name string) *Snapshot {
	return &Snapshot{name: name}
}

func (s *Snapshot) WithCLI(cli *CLI) *Snapshot {
	s.cli = cli
	return s
}

func (s *Snapshot) Update() *Snapshot {
	s.updates = true
	return s
}

func (s *Snapshot) Match(t *testing.T, result *Result) {
	snapshotDir := filepath.Join("testdata", "snapshots")
	snapshotFile := filepath.Join(snapshotDir, s.name+".txt")

	if s.updates {
		// Update snapshot
		os.MkdirAll(snapshotDir, 0755)
		os.WriteFile(snapshotFile, []byte(result.Output()), 0644)
		t.Logf("Updated snapshot: %s", snapshotFile)
		return
	}

	// Read expected snapshot
	expected, err := os.ReadFile(snapshotFile)
	if err != nil {
		t.Fatalf("Snapshot not found: %s (run with -update to create)", snapshotFile)
	}

	actual := result.Output()
	if string(expected) != actual {
		t.Errorf("Snapshot mismatch for %s:\nExpected:\n%s\n\nActual:\n%s", s.name, expected, actual)
	}
}

type Benchmark struct {
	name       string
	cli        *CLI
	iterations int
}

func NewBenchmark(name string) *Benchmark {
	return &Benchmark{
		name:       name,
		iterations: 10,
	}
}

func (b *Benchmark) WithCLI(cli *CLI) *Benchmark {
	b.cli = cli
	return b
}

func (b *Benchmark) Iterations(n int) *Benchmark {
	b.iterations = n
	return b
}

func (b *Benchmark) Run(args ...string) *BenchmarkResult {
	results := make([]*Result, b.iterations)

	for i := 0; i < b.iterations; i++ {
		results[i] = b.cli.Run(args...)
	}

	var totalDuration time.Duration
	var avgDuration time.Duration
	var minDuration = time.Hour
	var maxDuration time.Duration

	for _, r := range results {
		totalDuration += r.Duration
		if r.Duration < minDuration {
			minDuration = r.Duration
		}
		if r.Duration > maxDuration {
			maxDuration = r.Duration
		}
	}

	if len(results) > 0 {
		avgDuration = totalDuration / time.Duration(len(results))
	}

	return &BenchmarkResult{
		Name:       b.name,
		Iterations: b.iterations,
		Results:    results,
		TotalTime:  totalDuration,
		AvgTime:    avgDuration,
		MinTime:    minDuration,
		MaxTime:    maxDuration,
	}
}

type BenchmarkResult struct {
	Name       string
	Iterations int
	Results    []*Result
	TotalTime  time.Duration
	AvgTime    time.Duration
	MinTime    time.Duration
	MaxTime    time.Duration
}

func (r *BenchmarkResult) Report(t *testing.B) {
	t.Logf("Benchmark: %s", r.Name)
	t.Logf("Iterations: %d", r.Iterations)
	t.Logf("Total Time: %v", r.TotalTime)
	t.Logf("Avg Time: %v", r.AvgTime)
	t.Logf("Min Time: %v", r.MinTime)
	t.Logf("Max Time: %v", r.MaxTime)
}

func (r *BenchmarkResult) JSON() string {
	data, _ := json.MarshalIndent(r, "", "  ")
	return string(data)
}

type MockServer struct {
	port    int
	handler func(*Result)
}

func NewMockServer(port int) *MockServer {
	return &MockServer{port: port}
}

func (s *MockServer) WithHandler(handler func(*Result)) *MockServer {
	s.handler = handler
	return s
}

func (s *MockServer) Start() error {
	// Simplified - in real implementation, would start HTTP server
	return nil
}

func (s *MockServer) Stop() {
	// Simplified - in real implementation, would stop HTTP server
}

func (s *MockServer) URL() string {
	return fmt.Sprintf("http://localhost:%d", s.port)
}

type GoldenFile struct {
	name     string
	filePath string
}

func NewGoldenFile(name, filePath string) *GoldenFile {
	return &GoldenFile{
		name:     name,
		filePath: filePath,
	}
}

func (g *GoldenFile) Update(expected string) error {
	return os.WriteFile(g.filePath, []byte(expected), 0644)
}

func (g *GoldenFile) Read() (string, error) {
	data, err := os.ReadFile(g.filePath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (g *GoldenFile) Match(t *testing.T, actual string) {
	expected, err := g.Read()
	if err != nil {
		if os.IsNotExist(err) {
			t.Fatalf("Golden file not found: %s", g.filePath)
		}
		t.Fatalf("Failed to read golden file: %v", err)
	}

	if expected != actual {
		t.Errorf("Golden file mismatch for %s", g.name)
		t.Logf("Expected:\n%s", expected)
		t.Logf("Actual:\n%s", actual)
	}
}

func MatchOutput(t *testing.T, pattern string, output string) bool {
	matched, err := regexp.MatchString(pattern, output)
	if err != nil {
		t.Fatalf("Invalid regex pattern: %v", err)
	}
	return matched
}

func AssertContains(t *testing.T, output, substr string) {
	if !strings.Contains(output, substr) {
		t.Errorf("Expected output to contain %q, but it didn't:\n%s", substr, output)
	}
}

func AssertNotContains(t *testing.T, output, substr string) {
	if strings.Contains(output, substr) {
		t.Errorf("Expected output to NOT contain %q, but it did:\n%s", substr, output)
	}
}

func AssertExitCode(t *testing.T, result *Result, code int) {
	if result.ExitCode != code {
		t.Errorf("Expected exit code %d, got %d.\nOutput:\n%s\nStderr:\n%s",
			code, result.ExitCode, result.Stdout, result.Stderr)
	}
}

func AssertSuccess(t *testing.T, result *Result) {
	if result.Failed() {
		t.Errorf("Expected success, but command failed with exit code %d.\nOutput:\n%s\nStderr:\n%s",
			result.ExitCode, result.Stdout, result.Stderr)
	}
}

func AssertFailure(t *testing.T, result *Result) {
	if result.Success() {
		t.Errorf("Expected failure, but command succeeded")
	}
}
