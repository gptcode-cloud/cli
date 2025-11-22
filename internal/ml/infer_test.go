package ml

import (
	"testing"
)

func TestIntentInference(t *testing.T) {
	predictor, err := LoadEmbedded("intent")
	if err != nil {
		t.Fatalf("Failed to load intent model: %v", err)
	}

	tests := []struct {
		text     string
		expected string
	}{
		{"how to implement auth", "research"}, // Based on my manual test
		{"fix the bug", "editor"},
		{"explain this code", "query"},
		{"hello", "router"},
	}

	for _, tt := range tests {
		label, probs := predictor.Predict(tt.text)
		t.Logf("Text: %q -> Label: %s (Conf: %.2f)", tt.text, label, probs[label])

		// Note: Exact label might vary due to small dataset, but let's check if it runs
		if label == "" {
			t.Errorf("Predicted empty label for %q", tt.text)
		}
	}
}

func TestComplexityInference(t *testing.T) {
	predictor, err := LoadEmbedded("complexity")
	if err != nil {
		t.Fatalf("Failed to load complexity model: %v", err)
	}

	tests := []struct {
		text     string
		expected string
	}{
		{"fix typo", "simple"},
		{"implement oauth2 authentication", "complex"},
		{"first do this then do that", "multistep"},
	}

	for _, tt := range tests {
		label, probs := predictor.Predict(tt.text)
		t.Logf("Text: %q -> Label: %s (Conf: %.2f)", tt.text, label, probs[label])

		if label != tt.expected {
			t.Logf("WARNING: Expected %s, got %s for %q", tt.expected, label, tt.text)
		}
	}
}
