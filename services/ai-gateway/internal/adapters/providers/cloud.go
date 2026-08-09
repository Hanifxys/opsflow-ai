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

type CloudProvider struct {
	apiKey  string
	baseURL string
	client  *http.Client
}

func NewCloudProvider(apiKey, baseURL string) *CloudProvider {
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}
	return &CloudProvider{
		apiKey:  apiKey,
		baseURL: baseURL,
		client:  &http.Client{Timeout: 60 * time.Second},
	}
}

func (p *CloudProvider) Name() string {
	return "cloud"
}

type cloudMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type cloudRequest struct {
	Model    string         `json:"model"`
	Messages []cloudMessage `json:"messages"`
}

type cloudResponse struct {
	Choices []struct {
		Message cloudMessage `json:"message"`
	} `json:"choices"`
}

func (p *CloudProvider) GenerateContent(ctx context.Context, req ports.LLMRequest) (*ports.LLMResponse, error) {
	if p.apiKey == "" {
		return nil, fmt.Errorf("cloud provider API key not configured")
	}

	model := req.Model
	if model == "" {
		model = "gpt-4o-mini"
	}

	msgs := make([]cloudMessage, len(req.Messages))
	for i, m := range req.Messages {
		msgs[i] = cloudMessage{
			Role:    m.Role,
			Content: m.Content,
		}
	}

	bodyData, _ := json.Marshal(cloudRequest{
		Model:    model,
		Messages: msgs,
	})

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/chat/completions", bytes.NewBuffer(bodyData))
	if err != nil {
		return nil, fmt.Errorf("failed to create cloud LLM request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("cloud LLM request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("cloud LLM returned status code %d", resp.StatusCode)
	}

	var res cloudResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, fmt.Errorf("failed to decode cloud LLM response: %w", err)
	}

	if len(res.Choices) == 0 {
		return nil, fmt.Errorf("empty choice returned from cloud LLM")
	}

	return &ports.LLMResponse{
		Content: res.Choices[0].Message.Content,
	}, nil
}
