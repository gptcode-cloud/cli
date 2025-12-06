package validation

import (
	"bufio"
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
)

type CoverageResult struct {
	Success      bool
	Coverage     float64
	Output       string
	ErrorMessage string
}

type CoverageExecutor struct {
	workDir string
}

func NewCoverageExecutor(workDir string) *CoverageExecutor {
	return &CoverageExecutor{workDir: workDir}
}

func (ce *CoverageExecutor) RunCoverage(min float64) (*CoverageResult, error) {
	res := &CoverageResult{}
	profile := filepath.Join(ce.workDir, "coverage.out")
	_ = os.Remove(profile)

	cmd := exec.Command("go", "test", "./...", "-coverprofile=coverage.out", "-covermode=atomic")
	cmd.Dir = ce.workDir
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	err := cmd.Run()
	res.Output = buf.String()
	if err != nil {
		res.ErrorMessage = err.Error()
		return res, err
	}

	if _, err := os.Stat(profile); err != nil {
		res.ErrorMessage = "coverage profile not found"
		return res, errors.New("coverage profile not found")
	}

	tool := exec.Command("go", "tool", "cover", "-func=coverage.out")
	tool.Dir = ce.workDir
	out, err := tool.Output()
	res.Output += string(out)
	if err != nil {
		res.ErrorMessage = err.Error()
		return res, err
	}

	s := bufio.NewScanner(bytes.NewReader(out))
	re := regexp.MustCompile(`total:\s+\(statements\)\s+([0-9]+\.[0-9]+)%`)
	for s.Scan() {
		m := re.FindStringSubmatch(s.Text())
		if len(m) == 2 {
			v, _ := strconv.ParseFloat(m[1], 64)
			res.Coverage = v
			break
		}
	}

	if min > 0 && res.Coverage < min {
		res.Success = false
		res.ErrorMessage = "coverage below threshold"
		return res, errors.New("coverage below threshold")
	}

	res.Success = true
	_ = os.Remove(profile)
	return res, nil
}
