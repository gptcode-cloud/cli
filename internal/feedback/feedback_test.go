package feedback

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRecord(t *testing.T) {
	tempDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", origHome)

	event := Event{
		Sentiment: SentimentGood,
		Backend:   "groq",
		Model:     "llama-3.3-70b",
		Agent:     "query",
		Context:   "test context",
	}

	err := Record(event)
	if err != nil {
		t.Fatalf("Record() error = %v", err)
	}

	dir := filepath.Join(tempDir, ".gptcode", "feedback")
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Errorf("Feedback directory not created")
	}
}

func TestLoadAll(t *testing.T) {
	tempDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", origHome)

	event1 := Event{
		Sentiment: SentimentGood,
		Backend:   "groq",
		Model:     "model-a",
		Agent:     "query",
	}

	event2 := Event{
		Sentiment: SentimentBad,
		Backend:   "groq",
		Model:     "model-b",
		Agent:     "editor",
	}

	if err := Record(event1); err != nil {
		t.Fatalf("Record(event1) error = %v", err)
	}

	time.Sleep(time.Millisecond * 10)

	if err := Record(event2); err != nil {
		t.Fatalf("Record(event2) error = %v", err)
	}

	events, err := LoadAll()
	if err != nil {
		t.Fatalf("LoadAll() error = %v", err)
	}

	if len(events) != 2 {
		t.Errorf("LoadAll() got %d events, want 2", len(events))
	}
}

func TestAnalyze(t *testing.T) {
	events := []Event{
		{Sentiment: SentimentGood, Backend: "groq", Model: "model-a", Agent: "query"},
		{Sentiment: SentimentGood, Backend: "groq", Model: "model-a", Agent: "query"},
		{Sentiment: SentimentBad, Backend: "groq", Model: "model-a", Agent: "query"},
		{Sentiment: SentimentGood, Backend: "ollama", Model: "model-b", Agent: "editor"},
	}

	stats := Analyze(events)

	if stats.TotalEvents != 4 {
		t.Errorf("TotalEvents = %d, want 4", stats.TotalEvents)
	}

	if stats.GoodCount != 3 {
		t.Errorf("GoodCount = %d, want 3", stats.GoodCount)
	}

	if stats.BadCount != 1 {
		t.Errorf("BadCount = %d, want 1", stats.BadCount)
	}

	groqStats, ok := stats.ByBackend["groq"]
	if !ok {
		t.Fatal("groq stats not found")
	}

	if groqStats.Total != 3 {
		t.Errorf("groq total = %d, want 3", groqStats.Total)
	}

	if groqStats.Ratio < 0.66 || groqStats.Ratio > 0.67 {
		t.Errorf("groq ratio = %f, want ~0.67", groqStats.Ratio)
	}

	modelAStats, ok := stats.ByModel["model-a"]
	if !ok {
		t.Fatal("model-a stats not found")
	}

	if modelAStats.Total != 3 {
		t.Errorf("model-a total = %d, want 3", modelAStats.Total)
	}

	if modelAStats.Good != 2 {
		t.Errorf("model-a good = %d, want 2", modelAStats.Good)
	}
}

func TestGetBestModels(t *testing.T) {
	tempDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", origHome)

	events := []Event{
		{Sentiment: SentimentGood, Model: "model-a", Agent: "query"},
		{Sentiment: SentimentGood, Model: "model-a", Agent: "query"},
		{Sentiment: SentimentGood, Model: "model-a", Agent: "query"},
		{Sentiment: SentimentGood, Model: "model-b", Agent: "query"},
		{Sentiment: SentimentBad, Model: "model-b", Agent: "query"},
		{Sentiment: SentimentBad, Model: "model-b", Agent: "query"},
	}

	for _, e := range events {
		if err := Record(e); err != nil {
			t.Fatalf("Record() error = %v", err)
		}
	}

	best := GetBestModels("query", 3)
	if len(best) != 2 {
		t.Errorf("GetBestModels() returned %d models, want 2", len(best))
	}

	if len(best) > 0 && best[0] != "model-a" {
		t.Errorf("Best model = %s, want model-a", best[0])
	}
}

func TestEmptyFeedback(t *testing.T) {
	tempDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", origHome)

	events, err := LoadAll()
	if err != nil {
		t.Fatalf("LoadAll() on empty dir error = %v", err)
	}

	if len(events) != 0 {
		t.Errorf("LoadAll() on empty dir got %d events, want 0", len(events))
	}

	stats := Analyze(events)
	if stats.TotalEvents != 0 {
		t.Errorf("Analyze(empty) TotalEvents = %d, want 0", stats.TotalEvents)
	}
}

func TestRecordWithExtendedFields(t *testing.T) {
	tempDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", origHome)

	event := Event{
		Sentiment:       SentimentBad,
		Backend:         "groq",
		Model:           "llama-3.3-70b",
		Agent:           "editor",
		Task:            "list files",
		WrongResponse:   "ls -la",
		CorrectResponse: "ls -lah",
		Source:          "shell",
		Kind:            "command",
		Files:           []string{"main.go", "config.yaml"},
		DiffPath:        "/tmp/test.patch",
	}

	err := Record(event)
	if err != nil {
		t.Fatalf("Record() error = %v", err)
	}

	events, err := LoadAll()
	if err != nil {
		t.Fatalf("LoadAll() error = %v", err)
	}

	if len(events) != 1 {
		t.Fatalf("LoadAll() got %d events, want 1", len(events))
	}

	loaded := events[0]
	if loaded.WrongResponse != "ls -la" {
		t.Errorf("WrongResponse = %s, want 'ls -la'", loaded.WrongResponse)
	}
	if loaded.CorrectResponse != "ls -lah" {
		t.Errorf("CorrectResponse = %s, want 'ls -lah'", loaded.CorrectResponse)
	}
	if loaded.Source != "shell" {
		t.Errorf("Source = %s, want 'shell'", loaded.Source)
	}
	if string(loaded.Kind) != "command" {
		t.Errorf("Kind = %s, want 'command'", loaded.Kind)
	}
	if len(loaded.Files) != 2 {
		t.Errorf("Files length = %d, want 2", len(loaded.Files))
	}
	if loaded.DiffPath != "/tmp/test.patch" {
		t.Errorf("DiffPath = %s, want '/tmp/test.patch'", loaded.DiffPath)
	}
}

func TestRecordWithPartialFields(t *testing.T) {
	tempDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", origHome)

	event := Event{
		Sentiment:       SentimentGood,
		Backend:         "ollama",
		Model:           "qwen-3",
		Agent:           "query",
		CorrectResponse: "helpful response",
	}

	err := Record(event)
	if err != nil {
		t.Fatalf("Record() error = %v", err)
	}

	events, err := LoadAll()
	if err != nil {
		t.Fatalf("LoadAll() error = %v", err)
	}

	if len(events) != 1 {
		t.Fatalf("LoadAll() got %d events, want 1", len(events))
	}

	loaded := events[0]
	if loaded.WrongResponse != "" {
		t.Errorf("WrongResponse should be empty, got %s", loaded.WrongResponse)
	}
	if loaded.CorrectResponse != "helpful response" {
		t.Errorf("CorrectResponse = %s, want 'helpful response'", loaded.CorrectResponse)
	}
	if len(loaded.Files) != 0 {
		t.Errorf("Files should be empty, got %v", loaded.Files)
	}
}
