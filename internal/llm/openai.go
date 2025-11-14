package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
)

type OpenAIProvider struct {
	APIKey  string
	BaseURL string
}

func NewOpenAI(baseURL, backendName string) *OpenAIProvider {
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}
	if !strings.HasSuffix(baseURL, "/chat/completions") {
		baseURL = baseURL + "/chat/completions"
	}

	envVar := strings.ToUpper(backendName) + "_API_KEY"
	apiKey := os.Getenv(envVar)
	if apiKey == "" && backendName == "openai" {
		apiKey = os.Getenv("OPENAI_API_KEY")
	}

	return &OpenAIProvider{
		APIKey:  apiKey,
		BaseURL: baseURL,
	}
}

type openaiReq struct {
	Model    string          `json:"model"`
	Messages []openaiMessage `json:"messages"`
}

type openaiMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openaiResp struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func (o *OpenAIProvider) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	if o.APIKey == "" {
		return nil, errors.New("API key not defined")
	}

	body := openaiReq{
		Model: req.Model,
		Messages: []openaiMessage{
			{Role: "system", Content: req.SystemPrompt},
			{Role: "user", Content: req.UserPrompt},
		},
	}

	b, _ := json.Marshal(body)

	httpReq, _ := http.NewRequestWithContext(ctx, "POST", o.BaseURL, bytes.NewReader(b))
	httpReq.Header.Set("Authorization", "Bearer "+o.APIKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var or openaiResp
	if err := json.NewDecoder(resp.Body).Decode(&or); err != nil {
		return nil, err
	}

	if or.Error != nil {
		return nil, fmt.Errorf("API error: %s", or.Error.Message)
	}

	if len(or.Choices) == 0 {
		return nil, errors.New("empty response from API")
	}

	return &ChatResponse{Text: or.Choices[0].Message.Content}, nil
}
