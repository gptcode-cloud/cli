package live

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	PhoenixChannel       = "agent:lobby"
	HeartbeatInterval    = 30 * time.Second
	ReconnectDelay       = 5 * time.Second
	MaxReconnectAttempts = 3
)

type WsClient struct {
	url       string
	model     string
	agentID   string
	agentType string
	task      string
	hostname  string
	workspace string

	conn      *websocket.Conn
	mu        sync.RWMutex
	connected bool
	wsEnabled bool

	httpFallback *ReportConfig

	callbacks struct {
		onUpdatePrompt func(string)
		onPause        func()
		onResume       func()
		onKill         func()
	}

	done     chan struct{}
	receive  chan []byte
	stopChan chan struct{}
}

type WsMessage struct {
	Event   string          `json:"event"`
	Topic   string          `json:"topic"`
	Payload json.RawMessage `json:"payload"`
	Ref     int             `json:"ref"`
}

type PhoenixMessage struct {
	Event   string `json:"event"`
	Payload struct {
		Body string `json:"body,omitempty"`
	} `json:"payload"`
}

func NewWsClient(baseURL string) *WsClient {
	ws := &WsClient{
		receive:  make(chan []byte, 100),
		stopChan: make(chan struct{}),
		done:     make(chan struct{}),
	}

	if baseURL == "" {
		baseURL = "wss://gptcode.live"
	}

	ws.url = ws.convertToWs(baseURL)

	hostname, _ := os.Hostname()
	cwd, _ := os.Getwd()

	ws.hostname = hostname
	ws.workspace = filepath.Base(cwd)

	ws.httpFallback = &ReportConfig{
		BaseURL: ws.convertToHttp(baseURL),
		Client:  &http.Client{Timeout: 5 * time.Second},
		AgentID: ws.agentID,
	}

	return ws
}

func (w *WsClient) convertToWs(url string) string {
	url = strings.TrimSuffix(url, "/")
	if strings.HasPrefix(url, "https://") {
		return "wss://" + strings.TrimPrefix(url, "https://") + "/agent/websocket"
	}
	if strings.HasPrefix(url, "http://") {
		return "ws://" + strings.TrimPrefix(url, "http://") + "/agent/websocket"
	}
	if !strings.HasPrefix(url, "ws") {
		return "wss://" + url + "/agent/websocket"
	}
	return url
}

func (w *WsClient) convertToHttp(url string) string {
	url = strings.TrimSuffix(url, "/")
	url = strings.TrimSuffix(url, "/agent/websocket")
	if strings.HasPrefix(url, "wss://") {
		return "https://" + strings.TrimPrefix(url, "wss://")
	}
	if strings.HasPrefix(url, "ws://") {
		return "http://" + strings.TrimPrefix(url, "ws://")
	}
	return "https://" + url
}

func (w *WsClient) SetModel(model string) {
	w.model = model
}

func (w *WsClient) SetCallbacks(callbacks map[string]interface{}) {
	if fn, ok := callbacks["update_prompt"].(func(string)); ok {
		w.callbacks.onUpdatePrompt = fn
	}
	if fn, ok := callbacks["pause"].(func()); ok {
		w.callbacks.onPause = fn
	}
	if fn, ok := callbacks["resume"].(func()); ok {
		w.callbacks.onResume = fn
	}
	if fn, ok := callbacks["kill"].(func()); ok {
		w.callbacks.onKill = fn
	}
}

func (w *WsClient) Connect(agentID, agentType, task string) error {
	w.agentID = agentID
	w.agentType = agentType
	w.task = task

	if cwd, err := os.Getwd(); err == nil {
		w.workspace = filepath.Base(cwd)
	}

	if !w.wsEnabled {
		return w.httpFallback.Connect(agentID, agentType, task)
	}

	return w.wsConnect()
}

func (w *WsClient) wsConnect() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	var err error
	header := http.Header{}
	header.Set("Origin", "https://gptcode.live")

	w.conn, _, err = websocket.DefaultDialer.Dial(w.url, header)
	if err != nil {
		log.Printf("[WS] Connection failed: %v, falling back to HTTP", err)
		return w.httpFallback.Connect(w.agentID, w.agentType, w.task)
	}

	payload := map[string]interface{}{
		"agent_id": w.agentID,
		"engine":   "gptcode",
		"type":     w.agentType,
		"context":  w.workspace,
		"task":     w.task,
		"hostname": w.hostname,
		"model":    w.model,
	}

	if w.model != "" {
		payload["model"] = w.model
	}

	phoenixJoin := map[string]interface{}{
		"topic":   PhoenixChannel,
		"event":   "phx_join",
		"payload": payload,
		"ref":     1,
	}

	if err := w.conn.WriteJSON(phoenixJoin); err != nil {
		w.conn.Close()
		return fmt.Errorf("phx_join failed: %w", err)
	}

	w.connected = true
	go w.readLoop()
	go w.heartbeat()

	log.Printf("[WS] Connected to %s", w.url)
	return nil
}

func (w *WsClient) readLoop() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[WS] Read loop recovered: %v", r)
		}
	}()

	for {
		select {
		case <-w.stopChan:
			return
		default:
			_, data, err := w.conn.ReadMessage()
			if err != nil {
				log.Printf("[WS] Read error: %v", err)
				w.handleDisconnect()
				return
			}
			w.handleMessage(data)
		}
	}
}

func (w *WsClient) handleMessage(data []byte) {
	var msg WsMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		log.Printf("[WS] Failed to parse message: %v", err)
		return
	}

	switch msg.Event {
	case "update_prompt":
		var payload struct {
			Body string `json:"body"`
		}
		if err := json.Unmarshal(msg.Payload, &payload); err == nil && payload.Body != "" {
			if w.callbacks.onUpdatePrompt != nil {
				w.callbacks.onUpdatePrompt(payload.Body)
			}
		}

	case "update_context":
		var payload struct {
			Context string `json:"context"`
		}
		if err := json.Unmarshal(msg.Payload, &payload); err == nil && payload.Context != "" {
			if w.callbacks.onUpdatePrompt != nil {
				w.callbacks.onUpdatePrompt(payload.Context)
			}
		}

	case "pause":
		if w.callbacks.onPause != nil {
			w.callbacks.onPause()
		}

	case "resume":
		if w.callbacks.onResume != nil {
			w.callbacks.onResume()
		}

	case "kill":
		if w.callbacks.onKill != nil {
			w.callbacks.onKill()
		}

	case "phx_reply":
		log.Printf("[WS] Phoenix reply received")

	case "heartbeat":
		w.sendHeartbeat()
	}
}

func (w *WsClient) handleDisconnect() {
	w.mu.Lock()
	if !w.connected {
		w.mu.Unlock()
		return
	}
	w.connected = false
	w.mu.Unlock()

	log.Printf("[WS] Disconnected, attempting reconnect...")

	for i := 0; i < MaxReconnectAttempts; i++ {
		time.Sleep(ReconnectDelay)
		if err := w.wsConnect(); err == nil {
			log.Printf("[WS] Reconnected successfully")
			return
		}
		log.Printf("[WS] Reconnect attempt %d failed", i+1)
	}

	log.Printf("[WS] All reconnect attempts failed, using HTTP fallback")
}

func (w *WsClient) heartbeat() {
	ticker := time.NewTicker(HeartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-w.stopChan:
			return
		case <-ticker.C:
			w.sendHeartbeat()
		}
	}
}

func (w *WsClient) sendHeartbeat() {
	w.mu.RLock()
	connected := w.connected
	conn := w.conn
	w.mu.RUnlock()

	if !connected || conn == nil {
		return
	}

	heartbeat := map[string]interface{}{
		"topic": PhoenixChannel,
		"event": "heartbeat",
		"payload": map[string]string{
			"status": "ok",
		},
		"ref": time.Now().UnixNano() % 100000,
	}

	if err := conn.WriteJSON(heartbeat); err != nil {
		log.Printf("[WS] Heartbeat failed: %v", err)
	}
}

func (w *WsClient) Step(description string, stepType string) error {
	w.mu.RLock()
	connected := w.connected
	w.mu.RUnlock()

	if !connected {
		return w.httpFallback.Step(description, stepType)
	}

	payload := map[string]interface{}{
		"agent_id":    w.agentID,
		"description": description,
		"step_type":   stepType,
		"timestamp":   time.Now().Unix(),
	}

	msg := map[string]interface{}{
		"topic":   PhoenixChannel,
		"event":   "step",
		"payload": payload,
		"ref":     time.Now().UnixNano() % 100000,
	}

	w.mu.RLock()
	err := w.conn.WriteJSON(msg)
	w.mu.RUnlock()

	if err != nil {
		log.Printf("[WS] Step failed: %v, using HTTP fallback", err)
		return w.httpFallback.Step(description, stepType)
	}

	log.Printf("[WS] Step: %s", description)
	return nil
}

func (w *WsClient) Disconnect() error {
	close(w.stopChan)

	w.mu.Lock()
	defer w.mu.Unlock()

	if !w.connected {
		return nil
	}

	if w.conn != nil {
		phoenixLeave := map[string]interface{}{
			"topic":   PhoenixChannel,
			"event":   "phx_leave",
			"payload": map[string]interface{}{},
			"ref":     time.Now().UnixNano() % 100000,
		}
		w.conn.WriteJSON(phoenixLeave)
		w.conn.Close()
		w.connected = false
	}

	w.httpFallback.Disconnect()
	log.Printf("[WS] Disconnected")
	return nil
}

func (w *WsClient) EnableWS(enabled bool) {
	w.wsEnabled = enabled
}

func (w *WsClient) IsConnected() bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.connected
}

type Reporter interface {
	Connect(agentID, agentType, task string) error
	Step(description string, stepType string) error
	Disconnect() error
	SetModel(model string)
	SetCallbacks(callbacks map[string]interface{})
	EnableWS(enabled bool)
	IsConnected() bool
}

type ReporterClient struct {
	ws    *WsClient
	http  *ReportConfig
	useWs bool
}

func NewReporterClient(baseURL string) *ReporterClient {
	return &ReporterClient{
		ws:   NewWsClient(baseURL),
		http: DefaultReportConfig(),
	}
}

func (r *ReporterClient) Connect(agentID, agentType, task string) error {
	if r.useWs {
		return r.ws.Connect(agentID, agentType, task)
	}
	return r.http.Connect(agentID, agentType, task)
}

func (r *ReporterClient) Step(description string, stepType string) error {
	if r.useWs {
		return r.ws.Step(description, stepType)
	}
	return r.http.Step(description, stepType)
}

func (r *ReporterClient) Disconnect() error {
	if r.useWs {
		return r.ws.Disconnect()
	}
	return r.http.Disconnect()
}

func (r *ReporterClient) SetModel(model string) {
	r.ws.SetModel(model)
}

func (r *ReporterClient) SetCallbacks(callbacks map[string]interface{}) {
	r.ws.SetCallbacks(callbacks)
}

func (r *ReporterClient) EnableWS(enabled bool) {
	r.useWs = enabled
	r.ws.EnableWS(enabled)
}
