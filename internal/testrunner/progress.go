package testrunner

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"time"
)

type ProgressTracker struct {
	startTime      time.Time
	timeout        int
	category       string
	notify         bool
	totalTests     int
	passedTests    int
	failedTests    int
	skippedTests   int
	currentTest    string
	mu             sync.Mutex
	done           chan bool
	testPattern    *regexp.Regexp
	statusPattern  *regexp.Regexp
	summaryPattern *regexp.Regexp
}

func NewProgressTracker(category string, timeout int, notify bool) *ProgressTracker {
	return &ProgressTracker{
		startTime:      time.Now(),
		timeout:        timeout,
		category:       category,
		notify:         notify,
		done:           make(chan bool),
		testPattern:    regexp.MustCompile(`^=== RUN\s+(.+)$`),
		statusPattern:  regexp.MustCompile(`^---\s+(PASS|FAIL|SKIP):\s+(.+)\s+\([\d.]+s\)$`),
		summaryPattern: regexp.MustCompile(`^(PASS|FAIL)$`),
	}
}

func (pt *ProgressTracker) Start() {
	go pt.updateProgress()
}

func (pt *ProgressTracker) Stop() {
	close(pt.done)
}

func (pt *ProgressTracker) updateProgress() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-pt.done:
			return
		case <-ticker.C:
			pt.printProgress()
		}
	}
}

func (pt *ProgressTracker) printProgress() {
	pt.mu.Lock()
	defer pt.mu.Unlock()

	elapsed := time.Since(pt.startTime)
	remaining := time.Duration(pt.timeout)*time.Second - elapsed

	if remaining < 0 {
		remaining = 0
	}

	test := pt.currentTest
	if test == "" {
		test = "waiting..."
	} else if len(test) > 50 {
		test = test[:47] + "..."
	}

	fmt.Printf("\r %s | [OK] %d passed | [ERROR] %d failed |  %s remaining | [RETRY] %s",
		formatDuration(elapsed),
		pt.passedTests,
		pt.failedTests,
		formatDuration(remaining),
		test)
}

func (pt *ProgressTracker) ParseOutput(r io.Reader) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		fmt.Println(line)

		pt.mu.Lock()
		if matches := pt.testPattern.FindStringSubmatch(line); matches != nil {
			pt.currentTest = matches[1]
			pt.totalTests++
		} else if matches := pt.statusPattern.FindStringSubmatch(line); matches != nil {
			status := matches[1]
			switch status {
			case "PASS":
				pt.passedTests++
			case "FAIL":
				pt.failedTests++
			case "SKIP":
				pt.skippedTests++
			}
		}
		pt.mu.Unlock()
	}
}

func (pt *ProgressTracker) Finish(success bool) {
	pt.Stop()

	elapsed := time.Since(pt.startTime)

	fmt.Print("\r")
	fmt.Print(strings.Repeat(" ", 120))
	fmt.Print("\r")
	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	if pt.totalTests == 0 {
		if success {
			fmt.Println("[OK] All tests passed!")
		} else {
			fmt.Println("[ERROR] Tests failed!")
		}
	} else {
		if success {
			total := pt.passedTests + pt.failedTests + pt.skippedTests
			fmt.Printf("[OK] All tests passed! (%d/%d)\n", pt.passedTests, total)
			if pt.skippedTests > 0 {
				fmt.Printf("   [SKIP]  %d skipped\n", pt.skippedTests)
			}
		} else {
			fmt.Printf("[ERROR] Tests failed! (%d passed, %d failed)\n", pt.passedTests, pt.failedTests)
			if pt.skippedTests > 0 {
				fmt.Printf("   [SKIP]  %d skipped\n", pt.skippedTests)
			}
		}
	}

	fmt.Printf(" Total time: %s\n", formatDuration(elapsed))

	if pt.notify {
		pt.sendNotification(success)
	}
}

func (pt *ProgressTracker) sendNotification(success bool) {
	if os.Getenv("GPTCODE_NO_NOTIFY") != "" {
		return
	}

	title := "GPTCode E2E Tests"
	var message string
	var sound string

	if success {
		message = fmt.Sprintf("[OK] All tests passed (%d/%d)", pt.passedTests, pt.totalTests)
		sound = "Glass"
	} else {
		message = fmt.Sprintf("[ERROR] Tests failed (%d passed, %d failed)", pt.passedTests, pt.failedTests)
		sound = "Basso"
	}

	cmd := exec.Command("osascript", "-e",
		fmt.Sprintf(`display notification "%s" with title "%s" sound name "%s"`,
			message, title, sound))

	_ = cmd.Run()
}

func formatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second

	if h > 0 {
		return fmt.Sprintf("%dh%02dm%02ds", h, m, s)
	}
	if m > 0 {
		return fmt.Sprintf("%dm%02ds", m, s)
	}
	return fmt.Sprintf("%ds", s)
}

func RunTestsWithProgress(category, backend, profile string, timeout int, notify bool) error {
	testDir := "tests/e2e"
	if category != "all" {
		testDir = fmt.Sprintf("tests/e2e/%s", category)
	}

	if _, err := os.Stat(testDir); os.IsNotExist(err) {
		return fmt.Errorf("test directory not found: %s\n\nAvailable categories: run, chat, tdd, integration", testDir)
	}

	args := []string{"test", "-tags=e2e", "-v", "-timeout", fmt.Sprintf("%ds", timeout), fmt.Sprintf("./%s/...", testDir)}

	os.Setenv("E2E_BACKEND", backend)
	os.Setenv("E2E_PROFILE", profile)
	os.Setenv("E2E_TIMEOUT", fmt.Sprintf("%d", timeout))

	cmd := exec.Command("go", args...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	categoryName := category
	if category == "all" {
		categoryName = "All"
	} else if len(category) > 0 {
		categoryName = strings.ToUpper(category[:1]) + category[1:]
	}

	fmt.Printf("Running %s tests from %s...\n\n", categoryName, testDir)

	tracker := NewProgressTracker(category, timeout, notify)
	tracker.Start()

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		tracker.ParseOutput(stdout)
	}()

	go func() {
		defer wg.Done()
		tracker.ParseOutput(stderr)
	}()

	if err := cmd.Start(); err != nil {
		tracker.Stop()
		return err
	}

	wg.Wait()

	err = cmd.Wait()
	success := err == nil

	tracker.Finish(success)

	if err != nil {
		return fmt.Errorf("tests failed")
	}

	return nil
}
