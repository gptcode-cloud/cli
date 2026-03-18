package llm

import "testing"

func TestLoopDetector_ToolLoops(t *testing.T) {
	detector := NewLoopDetector("edit")

	// Call same tool 4 times - should not trigger loop
	for i := 0; i < 4; i++ {
		isLoop, _ := detector.RecordToolCall("read_file", `{"path": "/test.go"}`)
		if isLoop {
			t.Errorf("Expected no loop detection at call %d, threshold is 5", i+1)
		}
	}

	// 5th call should trigger loop
	isLoop, reason := detector.RecordToolCall("read_file", `{"path": "/test.go"}`)
	if !isLoop {
		t.Error("Expected loop detection at 5th identical call")
	}
	if reason == "" {
		t.Error("Expected non-empty loop reason")
	}
}

func TestLoopDetector_ContentLoops(t *testing.T) {
	detector := NewLoopDetector("query")

	// Record same response 2 times - should not trigger
	for i := 0; i < 2; i++ {
		isLoop, _ := detector.RecordResponse("I don't know the answer")
		if isLoop {
			t.Errorf("Expected no content loop at response %d, threshold is 3", i+1)
		}
	}

	// 3rd same response should trigger
	isLoop, reason := detector.RecordResponse("I don't know the answer")
	if !isLoop {
		t.Error("Expected content loop at 3rd identical response")
	}
	if reason == "" {
		t.Error("Expected non-empty content loop reason")
	}
}

func TestLoopDetector_IntentAwareLimits(t *testing.T) {
	tests := []struct {
		intent      string
		expectedMax int
	}{
		{"query", 15},
		{"edit", 25},
		{"plan", 20},
		{"research", 30},
	}

	for _, tt := range tests {
		t.Run(tt.intent, func(t *testing.T) {
			detector := NewLoopDetector(tt.intent)
			max := detector.getMaxIterationsForIntent()
			if max != tt.expectedMax {
				t.Errorf("Intent %s: expected max %d, got %d", tt.intent, tt.expectedMax, max)
			}
		})
	}
}

func TestLoopDetector_DifferentToolCallsNoLoop(t *testing.T) {
	detector := NewLoopDetector("edit")

	// Different tool calls should not trigger loop
	tools := []struct{ name, args string }{
		{"read_file", `{"path": "/a.go"}`},
		{"read_file", `{"path": "/b.go"}`},
		{"write_file", `{"path": "/c.go"}`},
		{"read_file", `{"path": "/a.go"}`}, // Same as first but not consecutive
		{"list_dir", `{"path": "/"}`},
	}

	for _, tc := range tools {
		isLoop, _ := detector.RecordToolCall(tc.name, tc.args)
		if isLoop {
			t.Errorf("Unexpected loop for varied tool calls: %s", tc.name)
		}
	}
}
