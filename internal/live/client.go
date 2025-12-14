package live

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"gptcode/internal/crypto"

	"github.com/gorilla/websocket"
)

// Client connects to the GPTCode Live Dashboard via Phoenix WebSocket
type Client struct {
	conn               *websocket.Conn
	agentID            string
	url                string
	mu                 sync.Mutex
	joinRef            int
	msgRef             int
	onEdit             func(contextType, content string)
	e2e                *crypto.E2ESession
	encrypted          bool
	onEncryptedMessage func(data []byte)
}

// NewClient creates a new Live Dashboard client
func NewClient(dashboardURL, agentID string) *Client {
	return &Client{
		url:     dashboardURL,
		agentID: agentID,
		joinRef: 1,
		msgRef:  1,
	}
}

// Connect establishes WebSocket connection to Phoenix
func (c *Client) Connect() error {
	u, err := url.Parse(c.url)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	// Convert http/https to ws/wss
	scheme := "ws"
	if u.Scheme == "https" {
		scheme = "wss"
	}
	u.Scheme = scheme
	u.Path = "/socket/websocket"
	u.RawQuery = "vsn=2.0.0"

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	c.conn = conn

	// Join agent channel
	if err := c.joinChannel(); err != nil {
		conn.Close()
		return fmt.Errorf("failed to join channel: %w", err)
	}

	// Start message handler
	go c.handleMessages()

	return nil
}

func (c *Client) joinChannel() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	topic := fmt.Sprintf("agent:%s", c.agentID)
	msg := []interface{}{
		c.joinRef,
		c.msgRef,
		topic,
		"phx_join",
		map[string]interface{}{},
	}
	c.msgRef++

	return c.conn.WriteJSON(msg)
}

func (c *Client) handleMessages() {
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if !websocket.IsCloseError(err, websocket.CloseNormalClosure) {
				log.Printf("Live: connection error: %v", err)
			}
			return
		}

		var msg []interface{}
		if err := json.Unmarshal(message, &msg); err != nil {
			continue
		}

		if len(msg) < 5 {
			continue
		}

		event, ok := msg[3].(string)
		if !ok {
			continue
		}

		payload, ok := msg[4].(map[string]interface{})
		if !ok {
			continue
		}

		switch event {
		case "context_edit":
			c.handleContextEdit(payload)
		case "phx_reply":
			// Handle join reply
		}
	}
}

func (c *Client) handleContextEdit(payload map[string]interface{}) {
	contextType, _ := payload["type"].(string)
	content, _ := payload["content"].(string)

	if c.onEdit != nil {
		c.onEdit(contextType, content)
	} else {
		// Default: write to .gptcode/context/
		if err := WriteContextFile(contextType, content); err != nil {
			log.Printf("Live: failed to write context: %v", err)
		}
	}
}

// SendContextUpdate sends current project context to Live
func (c *Client) SendContextUpdate(shared, next, roadmap string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	topic := fmt.Sprintf("agent:%s", c.agentID)
	msg := []interface{}{
		c.joinRef,
		c.msgRef,
		topic,
		"context_update",
		map[string]interface{}{
			"shared":  shared,
			"next":    next,
			"roadmap": roadmap,
		},
	}
	c.msgRef++

	return c.conn.WriteJSON(msg)
}

// OnContextEdit sets callback for when Live edits context
func (c *Client) OnContextEdit(fn func(contextType, content string)) {
	c.onEdit = fn
}

// Close closes the WebSocket connection
func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// EnableEncryption initializes E2E encryption and sends public key to server
func (c *Client) EnableEncryption() error {
	session, err := crypto.NewE2ESession()
	if err != nil {
		return fmt.Errorf("failed to create E2E session: %w", err)
	}
	c.e2e = session

	// Send our public key to server
	c.mu.Lock()
	defer c.mu.Unlock()

	topic := fmt.Sprintf("agent:%s", c.agentID)
	msg := []interface{}{
		c.joinRef,
		c.msgRef,
		topic,
		"key_exchange",
		map[string]interface{}{
			"public_key": c.e2e.PublicKey(),
		},
	}
	c.msgRef++

	log.Printf("Live: Initiated key exchange, fingerprint: %s", c.e2e.Fingerprint())
	return c.conn.WriteJSON(msg)
}

// SetRemotePublicKey completes key exchange with browser's public key
func (c *Client) SetRemotePublicKey(publicKey string) error {
	if c.e2e == nil {
		return fmt.Errorf("encryption not initialized")
	}
	if err := c.e2e.SetRemotePublicKey(publicKey); err != nil {
		return err
	}
	c.encrypted = true
	log.Printf("Live: Key exchange complete, remote fingerprint: %s", c.e2e.RemoteFingerprint())
	return nil
}

// SendEncrypted sends an encrypted message
func (c *Client) SendEncrypted(sessionID string, data []byte) error {
	if !c.encrypted || c.e2e == nil {
		return fmt.Errorf("encryption not ready")
	}

	ciphertext, err := c.e2e.Encrypt(data)
	if err != nil {
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	topic := fmt.Sprintf("agent:%s", c.agentID)
	msg := []interface{}{
		c.joinRef,
		c.msgRef,
		topic,
		"encrypted_payload",
		map[string]interface{}{
			"session_id": sessionID,
			"data":       ciphertext,
		},
	}
	c.msgRef++

	return c.conn.WriteJSON(msg)
}

// DecryptMessage decrypts a message from the browser
func (c *Client) DecryptMessage(ciphertext string) ([]byte, error) {
	if !c.encrypted || c.e2e == nil {
		return nil, fmt.Errorf("encryption not ready")
	}
	return c.e2e.Decrypt(ciphertext)
}

// IsEncrypted returns true if E2E encryption is active
func (c *Client) IsEncrypted() bool {
	return c.encrypted && c.e2e != nil && c.e2e.IsReady()
}

// OnEncryptedMessage sets callback for encrypted messages from browser
func (c *Client) OnEncryptedMessage(fn func(data []byte)) {
	c.onEncryptedMessage = fn
}

// ReadContextFile reads a context file from .gptcode/context/
func ReadContextFile(contextType string) (string, error) {
	gptcodeDir, err := findGPTCodeDir()
	if err != nil {
		return "", err
	}

	filename := contextType + ".md"
	path := filepath.Join(gptcodeDir, "context", filename)

	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// WriteContextFile writes to a context file in .gptcode/context/
func WriteContextFile(contextType, content string) error {
	gptcodeDir, err := findGPTCodeDir()
	if err != nil {
		return err
	}

	filename := contextType + ".md"
	path := filepath.Join(gptcodeDir, "context", filename)

	return os.WriteFile(path, []byte(content), 0644)
}

// ReadAllContext reads all context files
func ReadAllContext() (shared, next, roadmap string, err error) {
	shared, _ = ReadContextFile("shared")
	next, _ = ReadContextFile("next")
	roadmap, _ = ReadContextFile("roadmap")
	return shared, next, roadmap, nil
}

func findGPTCodeDir() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	dir := cwd
	for {
		gptcodePath := filepath.Join(dir, ".gptcode")
		if _, err := os.Stat(gptcodePath); err == nil {
			return gptcodePath, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return "", fmt.Errorf(".gptcode directory not found")
}

// AutoSync connects and syncs context automatically
func AutoSync(dashboardURL, agentID string) (*Client, error) {
	client := NewClient(dashboardURL, agentID)

	if err := client.Connect(); err != nil {
		return nil, err
	}

	// Send initial context
	shared, next, roadmap, _ := ReadAllContext()
	if shared != "" || next != "" || roadmap != "" {
		if err := client.SendContextUpdate(shared, next, roadmap); err != nil {
			log.Printf("Live: failed to send initial context: %v", err)
		}
	}

	// Watch for local changes
	go watchContextChanges(client)

	return client, nil
}

func watchContextChanges(client *Client) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	var lastShared, lastNext, lastRoadmap string

	for range ticker.C {
		shared, next, roadmap, _ := ReadAllContext()

		if shared != lastShared || next != lastNext || roadmap != lastRoadmap {
			if err := client.SendContextUpdate(shared, next, roadmap); err != nil {
				log.Printf("Live: failed to sync context: %v", err)
			}
			lastShared, lastNext, lastRoadmap = shared, next, roadmap
		}
	}
}

// GetDashboardURL returns the Live dashboard URL from config or default
func GetDashboardURL() string {
	if url := os.Getenv("GPTCODE_LIVE_URL"); url != "" {
		return url
	}
	return "https://live.gptcode.app"
}

// GetAgentID returns a unique agent identifier
func GetAgentID() string {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}

	// Make it unique per workspace
	cwd, _ := os.Getwd()
	parts := strings.Split(cwd, string(os.PathSeparator))
	if len(parts) > 0 {
		hostname = hostname + "-" + parts[len(parts)-1]
	}

	return hostname
}
