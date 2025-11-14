package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

type OllamaProvider struct {
	BaseURL string
}

func NewOllama(baseURL string) *OllamaProvider {
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}
	if !strings.HasSuffix(baseURL, "/api/chat") {
		baseURL = baseURL + "/api/chat"
	}
	return &OllamaProvider{BaseURL: baseURL}
}

type ollamaReq struct {
	Model    string          `json:"model"`
	Messages []ollamaMessage `json:"messages"`
	Stream   bool            `json:"stream"`
}

type ollamaMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ollamaResp struct {
	Message struct {
		Content string `json:"content"`
	} `json:"message"`
}

func (o *OllamaProvider) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	body := ollamaReq{
		Model: req.Model,
		Messages: []ollamaMessage{
			{Role: "system", Content: req.SystemPrompt},
			{Role: "user", Content: req.UserPrompt},
		},
		Stream: false,
	}
	b, _ := json.Marshal(body)

	httpReq, _ := http.NewRequestWithContext(ctx, "POST", o.BaseURL, bytes.NewReader(b))
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var or ollamaResp
	if err := json.NewDecoder(resp.Body).Decode(&or); err != nil {
		return nil, err
	}
	if or.Message.Content == "" {
		return nil, errors.New("resposta vazia do Ollama")
	}

	return &ChatResponse{Text: or.Message.Content}, nil
}
