// Package acp implements the Agent Client Protocol (ACP) for GPTCode CLI.
//
// ACP is a JSON-RPC 2.0 based protocol that standardizes communication between
// code editors/IDEs and AI coding agents. It is analogous to LSP but for AI agents.
//
// Transport: stdio (newline-delimited JSON-RPC messages)
// Spec: https://agentclientprotocol.com
package acp

import (
	"encoding/json"
	"fmt"
)

// ProtocolVersion is the ACP protocol version we support.
const ProtocolVersion = 1

// --- JSON-RPC 2.0 types ---

// Request represents a JSON-RPC 2.0 request from client or agent.
type Request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// Response represents a JSON-RPC 2.0 response.
type Response struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *RPCError   `json:"error,omitempty"`
}

// Notification represents a JSON-RPC 2.0 notification (no ID, no response expected).
type Notification struct {
	JSONRPC string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

// RPCError represents a JSON-RPC 2.0 error object.
type RPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func (e *RPCError) Error() string {
	return fmt.Sprintf("RPC error %d: %s", e.Code, e.Message)
}

// Standard JSON-RPC 2.0 error codes.
const (
	CodeParseError     = -32700
	CodeInvalidRequest = -32600
	CodeMethodNotFound = -32601
	CodeInvalidParams  = -32602
	CodeInternalError  = -32603
)

// --- ACP Method constants ---

const (
	// Agent baseline methods (client → agent)
	MethodInitialize    = "initialize"
	MethodAuthenticate  = "authenticate"
	MethodSessionNew    = "session/new"
	MethodSessionPrompt = "session/prompt"

	// Agent optional methods
	MethodSessionLoad    = "session/load"
	MethodSessionSetMode = "session/set_mode"

	// Agent notifications (client → agent)
	MethodSessionCancel = "session/cancel"

	// Client baseline methods (agent → client)
	MethodRequestPermission = "session/request_permission"

	// Client optional methods (agent → client)
	MethodFSReadTextFile  = "fs/read_text_file"
	MethodFSWriteTextFile = "fs/write_text_file"
	MethodTerminalCreate  = "terminal/create"
	MethodTerminalOutput  = "terminal/output"
	MethodTerminalRelease = "terminal/release"
	MethodTerminalWaitExit = "terminal/wait_for_exit"
	MethodTerminalKill    = "terminal/kill"

	// Client notifications (agent → client)
	MethodSessionUpdate = "session/update"
)

// --- Initialization types ---

// InitializeParams is sent by the client in the initialize request.
type InitializeParams struct {
	ProtocolVersion    int                `json:"protocolVersion"`
	ClientCapabilities ClientCapabilities `json:"clientCapabilities"`
	ClientInfo         ImplementationInfo `json:"clientInfo"`
}

// InitializeResult is the agent's response to initialize.
type InitializeResult struct {
	ProtocolVersion   int                `json:"protocolVersion"`
	AgentCapabilities AgentCapabilities  `json:"agentCapabilities"`
	AgentInfo         ImplementationInfo `json:"agentInfo"`
	AuthMethods       []interface{}      `json:"authMethods"`
}

// ClientCapabilities describes what the client can do.
type ClientCapabilities struct {
	FS       *FSCapabilities `json:"fs,omitempty"`
	Terminal bool            `json:"terminal,omitempty"`
}

// FSCapabilities describes file system capabilities of the client.
type FSCapabilities struct {
	ReadTextFile  bool `json:"readTextFile,omitempty"`
	WriteTextFile bool `json:"writeTextFile,omitempty"`
}

// AgentCapabilities describes what the agent supports.
type AgentCapabilities struct {
	LoadSession       bool               `json:"loadSession"`
	PromptCapabilities PromptCapabilities `json:"promptCapabilities"`
	MCPCapabilities   MCPCapabilities    `json:"mcpCapabilities"`
}

// PromptCapabilities describes what content types the agent accepts.
type PromptCapabilities struct {
	Image           bool `json:"image"`
	Audio           bool `json:"audio"`
	EmbeddedContext bool `json:"embeddedContext"`
}

// MCPCapabilities describes MCP support.
type MCPCapabilities struct {
	HTTP bool `json:"http"`
	SSE  bool `json:"sse"`
}

// ImplementationInfo identifies an agent or client implementation.
type ImplementationInfo struct {
	Name    string `json:"name"`
	Title   string `json:"title"`
	Version string `json:"version"`
}

// --- Session types ---

// SessionNewParams is the params for session/new.
type SessionNewParams struct {
	WorkingDirectory string            `json:"workingDirectory,omitempty"`
	ConfigOptions    map[string]string `json:"configOptions,omitempty"`
}

// SessionNewResult is the result of session/new.
type SessionNewResult struct {
	SessionID string `json:"sessionId"`
}

// SessionPromptParams is the params for session/prompt.
type SessionPromptParams struct {
	SessionID string         `json:"sessionId"`
	Content   []ContentBlock `json:"content"`
}

// ContentBlock represents a piece of content in a prompt or update.
type ContentBlock struct {
	Type string `json:"type"` // "text", "image", "resource", "resourceLink"
	Text string `json:"text,omitempty"`
}

// SessionPromptResult is the result of session/prompt (returned when the turn ends).
type SessionPromptResult struct {
	StopReason string `json:"stopReason"` // "endTurn", "cancelled", "maxTurns"
}

// --- Update notification types ---

// SessionUpdateParams is sent as a notification from agent to client during a turn.
type SessionUpdateParams struct {
	SessionID string        `json:"sessionId"`
	Update    SessionUpdate `json:"update"`
}

// SessionUpdate is the polymorphic update payload.
type SessionUpdate struct {
	Kind string `json:"kind"` // "message", "toolCall", "toolCallUpdate", "plan", "availableCommands", "modeChange"

	// For kind="message"
	MessageChunk *MessageChunk `json:"messageChunk,omitempty"`

	// For kind="toolCall"
	ToolCall *ToolCallUpdate `json:"toolCall,omitempty"`

	// For kind="toolCallUpdate"
	ToolCallDelta *ToolCallDelta `json:"toolCallDelta,omitempty"`

	// For kind="plan"
	Plan *PlanUpdate `json:"plan,omitempty"`

	// For kind="availableCommands"
	Commands []SlashCommand `json:"commands,omitempty"`
}

// MessageChunk is a text chunk streamed to the client.
type MessageChunk struct {
	Content []ContentBlock `json:"content"`
}

// ToolCallUpdate notifies the client about a tool invocation.
type ToolCallUpdate struct {
	ToolCallID string `json:"toolCallId"`
	ToolName   string `json:"toolName"`
	Status     string `json:"status"` // "running", "completed", "error"
	Input      string `json:"input,omitempty"`
	Output     string `json:"output,omitempty"`
}

// ToolCallDelta is a partial update for a long-running tool call.
type ToolCallDelta struct {
	ToolCallID string `json:"toolCallId"`
	Output     string `json:"output,omitempty"`
}

// PlanUpdate provides plan information to the client.
type PlanUpdate struct {
	Title string     `json:"title"`
	Steps []PlanStep `json:"steps"`
}

// PlanStep is a single step in an agent plan.
type PlanStep struct {
	Title  string `json:"title"`
	Status string `json:"status"` // "pending", "running", "completed", "error"
}

// SlashCommand describes an available slash command.
type SlashCommand struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// --- Permission types ---

// RequestPermissionParams is sent from agent to client.
type RequestPermissionParams struct {
	SessionID   string `json:"sessionId"`
	Permissions []Permission `json:"permissions"`
}

// Permission is a single permission request.
type Permission struct {
	Type        string `json:"type"` // "fileWrite", "command"
	Description string `json:"description"`
	FilePath    string `json:"filePath,omitempty"`
	Command     string `json:"command,omitempty"`
}

// RequestPermissionResult is the client's response.
type RequestPermissionResult struct {
	Granted bool `json:"granted"`
}

// --- File system types ---

// FSReadTextFileParams for fs/read_text_file request.
type FSReadTextFileParams struct {
	Path string `json:"path"`
}

// FSReadTextFileResult for fs/read_text_file response.
type FSReadTextFileResult struct {
	Content string `json:"content"`
}

// FSWriteTextFileParams for fs/write_text_file request.
type FSWriteTextFileParams struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

// FSWriteTextFileResult for fs/write_text_file response.
type FSWriteTextFileResult struct {
	Success bool `json:"success"`
}

// --- Terminal types ---

// TerminalCreateParams for terminal/create request.
type TerminalCreateParams struct {
	Command string `json:"command"`
	Cwd     string `json:"cwd,omitempty"`
}

// TerminalCreateResult for terminal/create response.
type TerminalCreateResult struct {
	TerminalID string `json:"terminalId"`
}

// TerminalWaitExitResult for terminal/wait_for_exit response.
type TerminalWaitExitResult struct {
	ExitCode int `json:"exitCode"`
}

// --- Helpers ---

// NewResponse creates a successful JSON-RPC response.
func NewResponse(id interface{}, result interface{}) *Response {
	return &Response{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}
}

// NewErrorResponse creates an error JSON-RPC response.
func NewErrorResponse(id interface{}, code int, message string) *Response {
	return &Response{
		JSONRPC: "2.0",
		ID:      id,
		Error: &RPCError{
			Code:    code,
			Message: message,
		},
	}
}

// NewNotification creates a JSON-RPC notification.
func NewNotification(method string, params interface{}) *Notification {
	return &Notification{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
	}
}
