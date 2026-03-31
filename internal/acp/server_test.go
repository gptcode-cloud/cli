package acp

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

// --- JSON-RPC Protocol Tests ---

func TestProtocolTypes(t *testing.T) {
	t.Run("NewResponse marshals correctly", func(t *testing.T) {
		resp := NewResponse(1, map[string]string{"key": "value"})
		data, err := json.Marshal(resp)
		if err != nil {
			t.Fatalf("marshal error: %v", err)
		}

		var parsed Response
		if err := json.Unmarshal(data, &parsed); err != nil {
			t.Fatalf("unmarshal error: %v", err)
		}

		if parsed.JSONRPC != "2.0" {
			t.Errorf("expected jsonrpc 2.0, got %s", parsed.JSONRPC)
		}
		if parsed.Error != nil {
			t.Error("expected no error")
		}
	})

	t.Run("NewErrorResponse includes error code", func(t *testing.T) {
		resp := NewErrorResponse(42, CodeMethodNotFound, "test error")
		data, err := json.Marshal(resp)
		if err != nil {
			t.Fatalf("marshal error: %v", err)
		}

		var parsed Response
		json.Unmarshal(data, &parsed)

		if parsed.Error == nil {
			t.Fatal("expected error")
		}
		if parsed.Error.Code != CodeMethodNotFound {
			t.Errorf("expected code %d, got %d", CodeMethodNotFound, parsed.Error.Code)
		}
		if parsed.Error.Message != "test error" {
			t.Errorf("expected message 'test error', got '%s'", parsed.Error.Message)
		}
	})

	t.Run("NewNotification has no ID", func(t *testing.T) {
		notif := NewNotification("session/update", map[string]string{"foo": "bar"})
		data, err := json.Marshal(notif)
		if err != nil {
			t.Fatalf("marshal error: %v", err)
		}

		// Should not contain "id" field
		var raw map[string]interface{}
		json.Unmarshal(data, &raw)
		if _, hasID := raw["id"]; hasID {
			t.Error("notification should not have an id field")
		}
		if raw["method"] != "session/update" {
			t.Errorf("expected method session/update, got %v", raw["method"])
		}
	})
}

// --- Server Handshake Tests ---

type mockHandler struct {
	promptCalled bool
	promptText   string
}

func (m *mockHandler) HandlePrompt(ctx context.Context, sessionID string, content []ContentBlock, emitter UpdateEmitter) (SessionPromptResult, error) {
	m.promptCalled = true
	for _, block := range content {
		if block.Type == "text" {
			m.promptText = block.Text
		}
	}
	emitter.EmitText("Hello from GPTCode!")
	return SessionPromptResult{StopReason: "endTurn"}, nil
}

func TestServerInitialize(t *testing.T) {
	initReq := Request{
		JSONRPC: "2.0",
		ID:      float64(0),
		Method:  "initialize",
	}
	params := InitializeParams{
		ProtocolVersion: 1,
		ClientCapabilities: ClientCapabilities{
			FS:       &FSCapabilities{ReadTextFile: true, WriteTextFile: true},
			Terminal: true,
		},
		ClientInfo: ImplementationInfo{
			Name:    "test-client",
			Title:   "Test Client",
			Version: "1.0.0",
		},
	}
	paramsBytes, _ := json.Marshal(params)
	initReq.Params = paramsBytes

	reqLine, _ := json.Marshal(initReq)
	reqLine = append(reqLine, '\n')

	input := bytes.NewReader(reqLine)
	output := &bytes.Buffer{}

	handler := &mockHandler{}
	server := NewServerWithIO(input, output, handler)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Run will read the init request then EOF
	server.Run(ctx)

	// Parse the response
	respLine := strings.TrimSpace(output.String())
	if respLine == "" {
		t.Fatal("no response received")
	}

	var resp Response
	if err := json.Unmarshal([]byte(respLine), &resp); err != nil {
		t.Fatalf("failed to parse response: %v (raw: %s)", err, respLine)
	}

	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
	}

	// Verify the result contains agent info
	resultBytes, _ := json.Marshal(resp.Result)
	var initResult InitializeResult
	if err := json.Unmarshal(resultBytes, &initResult); err != nil {
		t.Fatalf("failed to parse init result: %v", err)
	}

	if initResult.ProtocolVersion != 1 {
		t.Errorf("expected protocol version 1, got %d", initResult.ProtocolVersion)
	}
	if initResult.AgentInfo.Name != "gptcode" {
		t.Errorf("expected agent name 'gptcode', got '%s'", initResult.AgentInfo.Name)
	}
	if initResult.AgentInfo.Title != "GPTCode" {
		t.Errorf("expected agent title 'GPTCode', got '%s'", initResult.AgentInfo.Title)
	}
}

func TestServerSessionFlow(t *testing.T) {
	// Build a sequence of requests: initialize -> session/new -> session/prompt
	var input bytes.Buffer

	// 1. Initialize
	initParams, _ := json.Marshal(InitializeParams{
		ProtocolVersion: 1,
		ClientCapabilities: ClientCapabilities{
			FS: &FSCapabilities{ReadTextFile: true, WriteTextFile: true},
		},
		ClientInfo: ImplementationInfo{Name: "test", Title: "Test", Version: "1.0"},
	})
	initReq, _ := json.Marshal(Request{JSONRPC: "2.0", ID: float64(1), Method: "initialize", Params: initParams})
	input.Write(initReq)
	input.WriteByte('\n')

	// 2. Session new
	sessionParams, _ := json.Marshal(SessionNewParams{WorkingDirectory: "/tmp"})
	sessionReq, _ := json.Marshal(Request{JSONRPC: "2.0", ID: float64(2), Method: "session/new", Params: sessionParams})
	input.Write(sessionReq)
	input.WriteByte('\n')

	output := &bytes.Buffer{}
	handler := &mockHandler{}
	server := NewServerWithIO(&input, output, handler)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	server.Run(ctx)

	// Parse responses (one per line)
	lines := strings.Split(strings.TrimSpace(output.String()), "\n")
	if len(lines) < 2 {
		t.Fatalf("expected at least 2 responses, got %d (output: %s)", len(lines), output.String())
	}

	// Check session/new response
	// It may be the 2nd or 3rd line (after init response and possibly an availableCommands notification)
	foundSession := false
	for _, line := range lines {
		var resp Response
		if err := json.Unmarshal([]byte(line), &resp); err != nil {
			continue
		}
		if resp.ID == float64(2) && resp.Result != nil {
			resultBytes, _ := json.Marshal(resp.Result)
			var sessionResult SessionNewResult
			if err := json.Unmarshal(resultBytes, &sessionResult); err == nil && sessionResult.SessionID != "" {
				foundSession = true
				if !strings.HasPrefix(sessionResult.SessionID, "session-") {
					t.Errorf("expected session ID to start with 'session-', got '%s'", sessionResult.SessionID)
				}
			}
		}
	}

	if !foundSession {
		t.Error("session/new response not found in output")
	}
}

// --- Slash Commands Tests ---

func TestSlashCommands(t *testing.T) {
	commands := GetSlashCommands()

	if len(commands) == 0 {
		t.Fatal("expected at least one slash command")
	}

	commandNames := make(map[string]bool)
	for _, cmd := range commands {
		commandNames[cmd.Name] = true
		if cmd.Description == "" {
			t.Errorf("command %s has no description", cmd.Name)
		}
	}

	// Verify key commands exist
	expectedCommands := []string{"/plan", "/review", "/research", "/implement", "/tdd"}
	for _, expected := range expectedCommands {
		if !commandNames[expected] {
			t.Errorf("expected command %s not found", expected)
		}
	}
}

// --- Utility Tests ---

func TestTruncate(t *testing.T) {
	tests := []struct {
		input    string
		maxLen   int
		expected string
	}{
		{"short", 10, "short"},
		{"exactly10!", 10, "exactly10!"},
		{"this is longer than ten", 10, "this is lo..."},
		{"", 5, ""},
	}

	for _, tt := range tests {
		result := truncate(tt.input, tt.maxLen)
		if result != tt.expected {
			t.Errorf("truncate(%q, %d) = %q, want %q", tt.input, tt.maxLen, result, tt.expected)
		}
	}
}
