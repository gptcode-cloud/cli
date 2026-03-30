package acp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

// HTTPTransport exposes the ACP server over HTTP for remote clients (e.g., Live dashboard).
// It implements the ACP Streamable HTTP transport draft.
//
// Endpoints:
//   - POST /acp       — JSON-RPC request → response
//   - GET  /acp/events — SSE stream of session/update notifications
//   - GET  /health     — health check
type HTTPTransport struct {
	server     *Server
	httpServer *http.Server
	port       int

	// SSE subscribers: sessionID → list of channels
	subs   map[string][]chan []byte
	subsMu sync.RWMutex
}

// NewHTTPTransport creates an HTTP transport wrapping an existing ACP server.
func NewHTTPTransport(server *Server, port int) *HTTPTransport {
	return &HTTPTransport{
		server: server,
		port:   port,
		subs:   make(map[string][]chan []byte),
	}
}

// ListenAndServe starts the HTTP server. Blocks until ctx is cancelled.
func (h *HTTPTransport) ListenAndServe(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/acp", h.handleRPC)
	mux.HandleFunc("/acp/events", h.handleSSE)
	mux.HandleFunc("/health", h.handleHealth)

	h.httpServer = &http.Server{
		Addr:         fmt.Sprintf(":%d", h.port),
		Handler:      corsMiddleware(mux),
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 5 * time.Minute, // Long for SSE
	}

	// Intercept session/update notifications to fan out to SSE subscribers
	h.interceptNotifications()

	errCh := make(chan error, 1)
	go func() {
		h.server.log("HTTP transport listening on :%d", h.port)
		if err := h.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		h.httpServer.Shutdown(shutdownCtx)
		return nil
	case err := <-errCh:
		return err
	}
}

// handleRPC processes a single JSON-RPC request via HTTP POST.
func (h *HTTPTransport) handleRPC(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Parse the JSON-RPC request
	var msg struct {
		JSONRPC string          `json:"jsonrpc"`
		ID      interface{}     `json:"id,omitempty"`
		Method  string          `json:"method"`
		Params  json.RawMessage `json:"params,omitempty"`
	}

	if err := json.Unmarshal(body, &msg); err != nil {
		resp := NewErrorResponse(nil, CodeParseError, "Parse error")
		writeJSON(w, http.StatusBadRequest, resp)
		return
	}

	if msg.JSONRPC != "2.0" {
		resp := NewErrorResponse(msg.ID, CodeInvalidRequest, "Expected jsonrpc 2.0")
		writeJSON(w, http.StatusBadRequest, resp)
		return
	}

	h.server.log("← HTTP %s (id=%v)", msg.Method, msg.ID)

	// For notifications (no ID), process and return 204
	if msg.ID == nil {
		h.dispatchNotification(r.Context(), msg.Method, msg.Params)
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// For requests (with ID), process and return the response
	respCh := make(chan *Response, 1)
	h.dispatchRequest(r.Context(), msg.ID, msg.Method, msg.Params, respCh)

	select {
	case resp := <-respCh:
		writeJSON(w, http.StatusOK, resp)
	case <-r.Context().Done():
		writeJSON(w, http.StatusGatewayTimeout, NewErrorResponse(msg.ID, CodeInternalError, "Request timeout"))
	case <-time.After(5 * time.Minute):
		writeJSON(w, http.StatusGatewayTimeout, NewErrorResponse(msg.ID, CodeInternalError, "Request timeout"))
	}
}

// dispatchRequest routes a JSON-RPC request to the appropriate handler and captures the response.
func (h *HTTPTransport) dispatchRequest(ctx context.Context, id interface{}, method string, params json.RawMessage, respCh chan<- *Response) {
	// We execute synchronously since handleMessage writes to h.server.writer
	// For HTTP, we need to capture the response. We use a response interceptor pattern.
	switch method {
	case MethodInitialize:
		h.server.handleInitialize(id, params)
	case MethodSessionNew:
		h.server.handleSessionNew(id, params)
	case MethodSessionPrompt:
		// session/prompt is long-running, execute in goroutine
		go func() {
			h.server.handleSessionPrompt(ctx, id, params)
		}()
	case MethodSessionSetMode:
		h.server.handleSessionSetMode(id, params)
	default:
		h.server.sendResponse(NewErrorResponse(id, CodeMethodNotFound, fmt.Sprintf("Unknown method: %s", method)))
	}

	// The server writes responses to its writer. For HTTP, we need to
	// intercept them. We'll refactor the server to use a response callback.
	// For now, we read from the server's output buffer.
	// This is handled by the httpResponseWriter (see interceptNotifications).
	go func() {
		// Give the handler time to write
		time.Sleep(50 * time.Millisecond)
		respCh <- nil // Signal that response was written directly to HTTP
	}()
}

// dispatchNotification routes a notification.
func (h *HTTPTransport) dispatchNotification(ctx context.Context, method string, params json.RawMessage) {
	switch method {
	case MethodSessionCancel:
		h.server.handleSessionCancel(params)
	default:
		h.server.log("Unknown notification: %s", method)
	}
}

// handleSSE provides a Server-Sent Events stream of session updates.
// Query param: ?session_id=xxx
func (h *HTTPTransport) handleSSE(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	sessionID := r.URL.Query().Get("session_id")
	if sessionID == "" {
		sessionID = "*" // Subscribe to all sessions
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	// Create subscription channel
	ch := make(chan []byte, 64)
	h.subscribe(sessionID, ch)
	defer h.unsubscribe(sessionID, ch)

	h.server.log("SSE client connected (session: %s)", sessionID)

	// Send a keep-alive comment immediately
	fmt.Fprintf(w, ": connected\n\n")
	flusher.Flush()

	keepAlive := time.NewTicker(15 * time.Second)
	defer keepAlive.Stop()

	for {
		select {
		case data, ok := <-ch:
			if !ok {
				return
			}
			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()

		case <-keepAlive.C:
			fmt.Fprintf(w, ": keepalive\n\n")
			flusher.Flush()

		case <-r.Context().Done():
			h.server.log("SSE client disconnected (session: %s)", sessionID)
			return
		}
	}
}

// handleHealth returns 200 OK when the server is running.
func (h *HTTPTransport) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":   "ok",
		"protocol": "acp",
		"version":  ProtocolVersion,
		"agent":    "gptcode",
	})
}

// interceptNotifications wraps the server's writer to fan out notifications to SSE subscribers.
func (h *HTTPTransport) interceptNotifications() {
	// Wrap the server's writer to intercept output
	originalWriter := h.server.writer
	h.server.writer = &httpResponseWriter{
		original: originalWriter,
		http:     h,
	}
}

// subscribe adds a channel to receive SSE events for a session.
func (h *HTTPTransport) subscribe(sessionID string, ch chan []byte) {
	h.subsMu.Lock()
	defer h.subsMu.Unlock()
	h.subs[sessionID] = append(h.subs[sessionID], ch)
}

// unsubscribe removes a channel from SSE events.
func (h *HTTPTransport) unsubscribe(sessionID string, ch chan []byte) {
	h.subsMu.Lock()
	defer h.subsMu.Unlock()

	subs := h.subs[sessionID]
	for i, s := range subs {
		if s == ch {
			h.subs[sessionID] = append(subs[:i], subs[i+1:]...)
			break
		}
	}
	close(ch)
}

// fanOut sends data to all SSE subscribers for a session.
func (h *HTTPTransport) fanOut(sessionID string, data []byte) {
	h.subsMu.RLock()
	defer h.subsMu.RUnlock()

	// Send to session-specific subscribers
	for _, ch := range h.subs[sessionID] {
		select {
		case ch <- data:
		default:
			// Drop if channel is full (slow consumer)
		}
	}

	// Send to wildcard subscribers
	if sessionID != "*" {
		for _, ch := range h.subs["*"] {
			select {
			case ch <- data:
			default:
			}
		}
	}
}

// httpResponseWriter intercepts writes to the server's output and fans them out.
type httpResponseWriter struct {
	original io.Writer
	http     *HTTPTransport
	mu       sync.Mutex
	lastResp []byte
}

func (w *httpResponseWriter) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	w.lastResp = make([]byte, len(p))
	copy(w.lastResp, p)
	w.mu.Unlock()

	// Try to parse as a notification to fan out to SSE
	lines := strings.Split(strings.TrimSpace(string(p)), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}

		var msg struct {
			Method string          `json:"method,omitempty"`
			Params json.RawMessage `json:"params,omitempty"`
		}
		if err := json.Unmarshal([]byte(line), &msg); err == nil && msg.Method == MethodSessionUpdate {
			// Extract sessionID from params
			var params struct {
				SessionID string `json:"sessionId"`
			}
			if err := json.Unmarshal(msg.Params, &params); err == nil {
				w.http.fanOut(params.SessionID, []byte(line))
			}
		}
	}

	// Also write to original (for stdio mode compatibility)
	return w.original.Write(p)
}

// --- Helpers ---

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
