package autonomous

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type Watcher struct {
	cwd          string
	paths        []string
	interval     time.Duration
	onFileChange func(path string)
	executor     *Executor
	stopCh       chan struct{}
}

func NewWatcher(cwd string, paths []string, interval time.Duration) *Watcher {
	if len(paths) == 0 {
		paths = []string{"."}
	}
	return &Watcher{
		cwd:      cwd,
		paths:    paths,
		interval: interval,
		stopCh:   make(chan struct{}),
	}
}

func (w *Watcher) SetExecutor(executor *Executor) {
	w.executor = executor
}

func (w *Watcher) OnFileChange(fn func(path string)) {
	w.onFileChange = fn
}

func (w *Watcher) Start(ctx context.Context, task string) error {
	fmt.Printf("[WATCH] Starting watcher on: %s\n", strings.Join(w.paths, ", "))
	fmt.Printf("[WATCH] Interval: %v\n", w.interval)
	fmt.Printf("[WATCH] Task: %s\n\n", task)

	// Track file modifications
	modTimes := make(map[string]time.Time)
	for _, path := range w.paths {
		w.updateModTimes(path, modTimes)
	}

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	// Run initial execution
	fmt.Println("[WATCH] Running initial task...")
	if err := w.runTask(ctx, task); err != nil {
		fmt.Printf("[WATCH] Initial task failed: %v\n", err)
	}

	for {
		select {
		case <-ctx.Done():
			fmt.Println("[WATCH] Context cancelled, stopping...")
			return nil
		case <-w.stopCh:
			fmt.Println("[WATCH] Stopped by user")
			return nil
		case <-ticker.C:
			if w.checkChanges(modTimes) {
				fmt.Println("[WATCH] File changes detected, running task...")
				if err := w.runTask(ctx, task); err != nil {
					fmt.Printf("[WATCH] Task failed: %v\n", err)
				}
				w.updateModTimes(w.paths[0], modTimes)
			}
		}
	}
}

func (w *Watcher) Stop() {
	close(w.stopCh)
}

func (w *Watcher) checkChanges(modTimes map[string]time.Time) bool {
	for _, path := range w.paths {
		changed := w.hasChanges(path, modTimes)
		if changed {
			return true
		}
	}
	return false
}

func (w *Watcher) hasChanges(path string, modTimes map[string]time.Time) bool {
	absPath := w.cwd
	if !strings.HasPrefix(path, "/") {
		absPath = filepath.Join(w.cwd, path)
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return false
	}

	if info.IsDir() {
		entries, err := os.ReadDir(absPath)
		if err != nil {
			return false
		}
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			fullPath := filepath.Join(absPath, entry.Name())
			if w.hasChanges(fullPath, modTimes) {
				return true
			}
		}
		return false
	}

	modTime, ok := modTimes[absPath]
	if !ok || info.ModTime().After(modTime) {
		return true
	}
	return false
}

func (w *Watcher) updateModTimes(path string, modTimes map[string]time.Time) {
	absPath := w.cwd
	if !strings.HasPrefix(path, "/") {
		absPath = filepath.Join(w.cwd, path)
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return
	}

	if info.IsDir() {
		entries, err := os.ReadDir(absPath)
		if err != nil {
			return
		}
		for _, entry := range entries {
			if entry.IsDir() {
				w.updateModTimes(filepath.Join(absPath, entry.Name()), modTimes)
			} else {
				fullPath := filepath.Join(absPath, entry.Name())
				info, _ := os.Stat(fullPath)
				if info != nil {
					modTimes[fullPath] = info.ModTime()
				}
			}
		}
		return
	}

	modTimes[absPath] = info.ModTime()
}

func (w *Watcher) runTask(ctx context.Context, task string) error {
	if w.executor == nil {
		return fmt.Errorf("no executor configured")
	}
	return w.executor.Execute(ctx, task)
}

type TaskQueue struct {
	tasks     []string
	pending   []string
	completed []string
	failed    []string
	mu        int
}

func NewTaskQueue(tasks []string) *TaskQueue {
	return &TaskQueue{
		tasks:     tasks,
		pending:   make([]string, len(tasks)),
		completed: []string{},
		failed:    []string{},
	}
}

func (q *TaskQueue) Next() string {
	if len(q.tasks) == 0 {
		return ""
	}
	task := q.tasks[0]
	q.tasks = q.tasks[1:]
	return task
}

func (q *TaskQueue) MarkCompleted(task string) {
	q.completed = append(q.completed, task)
}

func (q *TaskQueue) MarkFailed(task string) {
	q.failed = append(q.failed, task)
}

func (q *TaskQueue) Stats() (total, pending, completed, failed int) {
	return len(q.tasks) + len(q.pending) + len(q.completed) + len(q.failed),
		len(q.pending),
		len(q.completed),
		len(q.failed)
}

func (q *TaskQueue) Summary() string {
	total, pending, completed, failed := q.Stats()
	return fmt.Sprintf("Total: %d, Pending: %d, Completed: %d, Failed: %d",
		total, pending, completed, failed)
}

type BatchExecutor struct {
	concurrency int
	executor    *Executor
}

func NewBatchExecutor(concurrency int) *BatchExecutor {
	return &BatchExecutor{
		concurrency: concurrency,
	}
}

func (be *BatchExecutor) SetExecutor(executor *Executor) {
	be.executor = executor
}

func (be *BatchExecutor) ExecuteAll(ctx context.Context, tasks []string) *BatchResult {
	result := &BatchResult{
		Tasks:     make([]TaskResult, len(tasks)),
		StartTime: time.Now(),
	}

	sem := make(chan struct{}, be.concurrency)
	taskCh := make(chan int, len(tasks))

	// Start workers
	for i := range tasks {
		taskCh <- i
	}
	close(taskCh)

	for i := range taskCh {
		select {
		case sem <- struct{}{}:
			go func(idx int) {
				task := tasks[idx]
				start := time.Now()

				err := be.executor.Execute(ctx, task)

				result.mu.Lock()
				result.Tasks[idx] = TaskResult{
					Task:     task,
					Success:  err == nil,
					Error:    err,
					Duration: time.Since(start),
				}
				if err == nil {
					result.Completed++
				} else {
					result.Failed++
				}
				result.mu.Unlock()

				<-sem
			}(i)
		}
	}

	result.Duration = time.Since(result.StartTime)
	return result
}

type BatchResult struct {
	Tasks     []TaskResult
	Completed int
	Failed    int
	Duration  time.Duration
	StartTime time.Time
	mu        sync.Mutex
}

type TaskResult struct {
	Task     string
	Success  bool
	Error    error
	Duration time.Duration
}

func (r *BatchResult) Summary() string {
	return fmt.Sprintf("Completed: %d/%d, Failed: %d/%d, Duration: %v",
		r.Completed, len(r.Tasks), r.Failed, len(r.Tasks), r.Duration)
}

func (r *BatchResult) SuccessRate() float64 {
	if len(r.Tasks) == 0 {
		return 0
	}
	return float64(r.Completed) / float64(len(r.Tasks)) * 100
}
