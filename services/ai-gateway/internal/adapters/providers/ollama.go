package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/opsflow/ai-gateway/internal/ports"
)

type OllamaProvider struct {
	baseURL string
	client  *http.Client
}

func NewOllamaProvider(baseURL string) *OllamaProvider {
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}
	return &OllamaProvider{
		baseURL: baseURL,
		client:  &http.Client{Timeout: 60 * time.Second},
	}
}

func (p *OllamaProvider) Name() string {
	return "ollama"
}

type ollamaMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ollamaRequest struct {
	Model    string          `json:"model"`
	Messages []ollamaMessage `json:"messages"`
	Stream   bool            `json:"stream"`
}

type ollamaResponse struct {
	Message ollamaMessage `json:"message"`
}

func (p *OllamaProvider) GenerateContent(ctx context.Context, req ports.LLMRequest) (*ports.LLMResponse, error) {
	model := req.Model
	if model == "" {
		model = "llama3"
	}

	msgs := make([]ollamaMessage, len(req.Messages))
	for i, m := range req.Messages {
		msgs[i] = ollamaMessage{
			Role:    m.Role,
			Content: m.Content,
		}
	}

	bodyData, _ := json.Marshal(ollamaRequest{
		Model:    model,
		Messages: msgs,
		Stream:   false,
	})

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/api/chat", bytes.NewBuffer(bodyData))
	if err != nil {
		return nil, fmt.Errorf("failed to create ollama request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("ollama request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ollama returned status code %d", resp.StatusCode)
	}

	var res ollamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, fmt.Errorf("failed to unmarshal ollama response: %w", err)
	}

	return &ports.LLMResponse{
		Content: res.Message.Content,
	}, nil
}
