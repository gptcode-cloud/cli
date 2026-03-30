package acp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
	"sync/atomic"
)

// Server implements the ACP agent-side protocol over stdio.
// It reads JSON-RPC requests from stdin and writes responses/notifications to stdout.
// Logging goes to stderr per the ACP spec.
type Server struct {
	reader  *bufio.Reader
	writer  io.Writer
	logger  io.Writer
	mu      sync.Mutex // protects writer
	nextID  atomic.Int64

	// State
	initialized      bool
	clientCaps       ClientCapabilities
	sessions         map[string]*Session
	cancelFuncs      map[string]context.CancelFunc
	sessionsMu       sync.Mutex

	// Handler
	handler SessionHandler
}

// SessionHandler is implemented by the application to handle prompt execution.
type SessionHandler interface {
	// HandlePrompt processes a user prompt and streams updates back via the emitter.
	HandlePrompt(ctx context.Context, sessionID string, content []ContentBlock, emitter UpdateEmitter) (SessionPromptResult, error)
}

// UpdateEmitter is the interface passed to SessionHandler for streaming updates.
type UpdateEmitter interface {
	// EmitText sends a text chunk to the client.
	EmitText(text string)
	// EmitToolCallStart notifies that a tool call has started.
	EmitToolCallStart(toolCallID, toolName, input string)
	// EmitToolCallOutput sends incremental output for a running tool.
	EmitToolCallOutput(toolCallID, output string)
	// EmitToolCallComplete marks a tool call as completed.
	EmitToolCallComplete(toolCallID, output string)
	// EmitToolCallError marks a tool call as errored.
	EmitToolCallError(toolCallID, errorMsg string)
	// EmitPlan sends the current execution plan.
	EmitPlan(title string, steps []PlanStep)
}

// NewServer creates a new ACP server reading from stdin, writing to stdout.
func NewServer(handler SessionHandler) *Server {
	return &Server{
		reader:      bufio.NewReader(os.Stdin),
		writer:      os.Stdout,
		logger:      os.Stderr,
		sessions:    make(map[string]*Session),
		cancelFuncs: make(map[string]context.CancelFunc),
		handler:     handler,
	}
}

// NewServerWithIO creates an ACP server with custom I/O (for testing).
func NewServerWithIO(reader io.Reader, writer io.Writer, handler SessionHandler) *Server {
	return &Server{
		reader:      bufio.NewReader(reader),
		writer:      writer,
		logger:      io.Discard,
		sessions:    make(map[string]*Session),
		cancelFuncs: make(map[string]context.CancelFunc),
		handler:     handler,
	}
}

// Session holds state for an active conversation session.
type Session struct {
	ID               string
	WorkingDirectory string
	ConfigOptions    map[string]string
}

// Run starts the main server loop. It blocks until stdin is closed or ctx is cancelled.
func (s *Server) Run(ctx context.Context) error {
	s.log("GPTCode ACP server starting (protocol version %d)", ProtocolVersion)

	for {
		select {
		case <-ctx.Done():
			s.log("Server context cancelled, shutting down")
			return ctx.Err()
		default:
		}

		line, err := s.reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				s.log("stdin closed, shutting down")
				return nil
			}
			s.log("Error reading stdin: %v", err)
			return fmt.Errorf("read error: %w", err)
		}

		if len(line) == 0 || (len(line) == 1 && line[0] == '\n') {
			continue
		}

		go s.handleMessage(ctx, line)
	}
}

// handleMessage parses and dispatches a single JSON-RPC message.
func (s *Server) handleMessage(ctx context.Context, data []byte) {
	var msg struct {
		JSONRPC string          `json:"jsonrpc"`
		ID      interface{}     `json:"id,omitempty"`
		Method  string          `json:"method"`
		Params  json.RawMessage `json:"params,omitempty"`
	}

	if err := json.Unmarshal(data, &msg); err != nil {
		s.log("Parse error: %v", err)
		if msg.ID != nil {
			s.sendResponse(NewErrorResponse(msg.ID, CodeParseError, "Parse error"))
		}
		return
	}

	if msg.JSONRPC != "2.0" {
		s.log("Invalid JSON-RPC version: %s", msg.JSONRPC)
		if msg.ID != nil {
			s.sendResponse(NewErrorResponse(msg.ID, CodeInvalidRequest, "Expected jsonrpc 2.0"))
		}
		return
	}

	// Notifications have no ID
	isNotification := msg.ID == nil

	s.log("← %s (id=%v, notification=%v)", msg.Method, msg.ID, isNotification)

	switch msg.Method {
	case MethodInitialize:
		s.handleInitialize(msg.ID, msg.Params)
	case MethodSessionNew:
		s.handleSessionNew(msg.ID, msg.Params)
	case MethodSessionPrompt:
		s.handleSessionPrompt(ctx, msg.ID, msg.Params)
	case MethodSessionCancel:
		s.handleSessionCancel(msg.Params)
	case MethodSessionSetMode:
		s.handleSessionSetMode(msg.ID, msg.Params)
	default:
		s.log("Unknown method: %s", msg.Method)
		if !isNotification {
			s.sendResponse(NewErrorResponse(msg.ID, CodeMethodNotFound, fmt.Sprintf("Unknown method: %s", msg.Method)))
		}
	}
}

// handleInitialize processes the initialize handshake.
func (s *Server) handleInitialize(id interface{}, params json.RawMessage) {
	var p InitializeParams
	if err := json.Unmarshal(params, &p); err != nil {
		s.sendResponse(NewErrorResponse(id, CodeInvalidParams, "Invalid initialize params"))
		return
	}

	s.log("Client: %s %s (protocol v%d)", p.ClientInfo.Name, p.ClientInfo.Version, p.ProtocolVersion)

	// Store client capabilities
	s.clientCaps = p.ClientCapabilities
	s.initialized = true

	// Negotiate protocol version (we only support v1)
	version := ProtocolVersion
	if p.ProtocolVersion < version {
		version = p.ProtocolVersion
	}

	result := InitializeResult{
		ProtocolVersion: version,
		AgentCapabilities: AgentCapabilities{
			LoadSession: false, // Not implemented yet
			PromptCapabilities: PromptCapabilities{
				Image:           false,
				Audio:           false,
				EmbeddedContext: true,
			},
			MCPCapabilities: MCPCapabilities{
				HTTP: false,
				SSE:  false,
			},
		},
		AgentInfo: ImplementationInfo{
			Name:    "gptcode",
			Title:   "GPTCode",
			Version: "1.0.0",
		},
		AuthMethods: []interface{}{},
	}

	s.sendResponse(NewResponse(id, result))
	s.log("Initialization complete (negotiated protocol v%d)", version)
}

// handleSessionNew creates a new session.
func (s *Server) handleSessionNew(id interface{}, params json.RawMessage) {
	if !s.initialized {
		s.sendResponse(NewErrorResponse(id, CodeInternalError, "Not initialized"))
		return
	}

	var p SessionNewParams
	if params != nil {
		if err := json.Unmarshal(params, &p); err != nil {
			s.sendResponse(NewErrorResponse(id, CodeInvalidParams, "Invalid session/new params"))
			return
		}
	}

	sessionID := fmt.Sprintf("session-%d", s.nextID.Add(1))

	cwd := p.WorkingDirectory
	if cwd == "" {
		cwd, _ = os.Getwd()
	}

	session := &Session{
		ID:               sessionID,
		WorkingDirectory: cwd,
		ConfigOptions:    p.ConfigOptions,
	}

	s.sessionsMu.Lock()
	s.sessions[sessionID] = session
	s.sessionsMu.Unlock()

	s.log("Session created: %s (cwd: %s)", sessionID, cwd)

	// Send available commands after session creation
	s.sendAvailableCommands(sessionID)

	s.sendResponse(NewResponse(id, SessionNewResult{SessionID: sessionID}))
}

// handleSessionPrompt processes a user prompt.
func (s *Server) handleSessionPrompt(parentCtx context.Context, id interface{}, params json.RawMessage) {
	var p SessionPromptParams
	if err := json.Unmarshal(params, &p); err != nil {
		s.sendResponse(NewErrorResponse(id, CodeInvalidParams, "Invalid session/prompt params"))
		return
	}

	s.sessionsMu.Lock()
	session := s.sessions[p.SessionID]
	s.sessionsMu.Unlock()

	if session == nil {
		s.sendResponse(NewErrorResponse(id, CodeInvalidParams, "Unknown session"))
		return
	}

	// Create cancellable context for this prompt turn
	ctx, cancel := context.WithCancel(parentCtx)
	s.sessionsMu.Lock()
	s.cancelFuncs[p.SessionID] = cancel
	s.sessionsMu.Unlock()

	defer func() {
		cancel()
		s.sessionsMu.Lock()
		delete(s.cancelFuncs, p.SessionID)
		s.sessionsMu.Unlock()
	}()

	// Create an emitter that sends session/update notifications
	emitter := &serverEmitter{
		server:    s,
		sessionID: p.SessionID,
	}

	// Log the incoming prompt
	for _, block := range p.Content {
		if block.Type == "text" {
			s.log("Prompt [%s]: %s", p.SessionID, truncate(block.Text, 100))
		}
	}

	// Delegate to the handler
	result, err := s.handler.HandlePrompt(ctx, p.SessionID, p.Content, emitter)
	if err != nil {
		if ctx.Err() != nil {
			// Cancelled
			s.sendResponse(NewResponse(id, SessionPromptResult{StopReason: "cancelled"}))
			return
		}
		s.log("Prompt error: %v", err)
		emitter.EmitText(fmt.Sprintf("\n\n❌ Error: %v", err))
		result.StopReason = "endTurn"
	}

	if result.StopReason == "" {
		result.StopReason = "endTurn"
	}

	s.sendResponse(NewResponse(id, result))
}

// handleSessionCancel cancels an in-progress prompt.
func (s *Server) handleSessionCancel(params json.RawMessage) {
	var p struct {
		SessionID string `json:"sessionId"`
	}
	if err := json.Unmarshal(params, &p); err != nil {
		s.log("Invalid session/cancel params: %v", err)
		return
	}

	s.sessionsMu.Lock()
	cancel, ok := s.cancelFuncs[p.SessionID]
	s.sessionsMu.Unlock()

	if ok {
		s.log("Cancelling session %s", p.SessionID)
		cancel()
	}
}

// handleSessionSetMode changes the agent's operating mode.
func (s *Server) handleSessionSetMode(id interface{}, params json.RawMessage) {
	var p struct {
		SessionID string `json:"sessionId"`
		Mode      string `json:"mode"`
	}
	if err := json.Unmarshal(params, &p); err != nil {
		s.sendResponse(NewErrorResponse(id, CodeInvalidParams, "Invalid set_mode params"))
		return
	}

	s.log("Mode changed to: %s (session %s)", p.Mode, p.SessionID)

	// Store mode in session config
	s.sessionsMu.Lock()
	if session, ok := s.sessions[p.SessionID]; ok {
		if session.ConfigOptions == nil {
			session.ConfigOptions = make(map[string]string)
		}
		session.ConfigOptions["mode"] = p.Mode
	}
	s.sessionsMu.Unlock()

	s.sendResponse(NewResponse(id, map[string]interface{}{"ok": true}))
}

// sendAvailableCommands notifies the client about available slash commands.
func (s *Server) sendAvailableCommands(sessionID string) {
	commands := GetSlashCommands()
	s.sendNotification(MethodSessionUpdate, SessionUpdateParams{
		SessionID: sessionID,
		Update: SessionUpdate{
			Kind:     "availableCommands",
			Commands: commands,
		},
	})
}

// --- Transport helpers ---

// sendResponse marshals and writes a response to stdout.
func (s *Server) sendResponse(resp *Response) {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := json.Marshal(resp)
	if err != nil {
		fmt.Fprintf(s.logger, "[ACP] marshal error: %v\n", err)
		return
	}

	data = append(data, '\n')
	if _, err := s.writer.Write(data); err != nil {
		fmt.Fprintf(s.logger, "[ACP] write error: %v\n", err)
	}
}

// sendNotification marshals and writes a notification to stdout.
func (s *Server) sendNotification(method string, params interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()

	notif := NewNotification(method, params)
	data, err := json.Marshal(notif)
	if err != nil {
		fmt.Fprintf(s.logger, "[ACP] marshal error: %v\n", err)
		return
	}

	data = append(data, '\n')
	if _, err := s.writer.Write(data); err != nil {
		fmt.Fprintf(s.logger, "[ACP] write error: %v\n", err)
	}
}

// SendRequest sends a JSON-RPC request to the client and returns the raw response.
// This is used for agent→client methods like fs/read_text_file.
func (s *Server) SendRequest(method string, params interface{}) (*Response, error) {
	reqID := s.nextID.Add(1)

	paramsBytes, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("marshal params: %w", err)
	}

	req := Request{
		JSONRPC: "2.0",
		ID:      reqID,
		Method:  method,
		Params:  paramsBytes,
	}

	s.mu.Lock()
	data, err := json.Marshal(req)
	if err != nil {
		s.mu.Unlock()
		return nil, fmt.Errorf("marshal request: %w", err)
	}
	data = append(data, '\n')
	if _, err := s.writer.Write(data); err != nil {
		s.mu.Unlock()
		return nil, fmt.Errorf("write request: %w", err)
	}
	s.mu.Unlock()

	// Read response from stdin
	// Note: In a production implementation, we'd need a proper multiplexer
	// to distinguish between client-initiated requests and responses to our requests.
	// For now, we read the next line and parse it as a response.
	line, err := s.reader.ReadBytes('\n')
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	var resp Response
	if err := json.Unmarshal(line, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	return &resp, nil
}

// log writes a log message to stderr.
func (s *Server) log(format string, args ...interface{}) {
	fmt.Fprintf(s.logger, "[ACP] "+format+"\n", args...)
}

// ClientCapabilitiesFor returns the stored client capabilities.
func (s *Server) ClientCapabilitiesFor() ClientCapabilities {
	return s.clientCaps
}

// GetSession returns a session by ID.
func (s *Server) GetSession(sessionID string) *Session {
	s.sessionsMu.Lock()
	defer s.sessionsMu.Unlock()
	return s.sessions[sessionID]
}

// --- serverEmitter implements UpdateEmitter ---

type serverEmitter struct {
	server    *Server
	sessionID string
}

func (e *serverEmitter) EmitText(text string) {
	e.server.sendNotification(MethodSessionUpdate, SessionUpdateParams{
		SessionID: e.sessionID,
		Update: SessionUpdate{
			Kind: "message",
			MessageChunk: &MessageChunk{
				Content: []ContentBlock{
					{Type: "text", Text: text},
				},
			},
		},
	})
}

func (e *serverEmitter) EmitToolCallStart(toolCallID, toolName, input string) {
	e.server.sendNotification(MethodSessionUpdate, SessionUpdateParams{
		SessionID: e.sessionID,
		Update: SessionUpdate{
			Kind: "toolCall",
			ToolCall: &ToolCallUpdate{
				ToolCallID: toolCallID,
				ToolName:   toolName,
				Status:     "running",
				Input:      input,
			},
		},
	})
}

func (e *serverEmitter) EmitToolCallOutput(toolCallID, output string) {
	e.server.sendNotification(MethodSessionUpdate, SessionUpdateParams{
		SessionID: e.sessionID,
		Update: SessionUpdate{
			Kind: "toolCallUpdate",
			ToolCallDelta: &ToolCallDelta{
				ToolCallID: toolCallID,
				Output:     output,
			},
		},
	})
}

func (e *serverEmitter) EmitToolCallComplete(toolCallID, output string) {
	e.server.sendNotification(MethodSessionUpdate, SessionUpdateParams{
		SessionID: e.sessionID,
		Update: SessionUpdate{
			Kind: "toolCall",
			ToolCall: &ToolCallUpdate{
				ToolCallID: toolCallID,
				Status:     "completed",
				Output:     output,
			},
		},
	})
}

func (e *serverEmitter) EmitToolCallError(toolCallID, errorMsg string) {
	e.server.sendNotification(MethodSessionUpdate, SessionUpdateParams{
		SessionID: e.sessionID,
		Update: SessionUpdate{
			Kind: "toolCall",
			ToolCall: &ToolCallUpdate{
				ToolCallID: toolCallID,
				Status:     "error",
				Output:     errorMsg,
			},
		},
	})
}

func (e *serverEmitter) EmitPlan(title string, steps []PlanStep) {
	e.server.sendNotification(MethodSessionUpdate, SessionUpdateParams{
		SessionID: e.sessionID,
		Update: SessionUpdate{
			Kind: "plan",
			Plan: &PlanUpdate{
				Title: title,
				Steps: steps,
			},
		},
	})
}

// --- Utilities ---

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
