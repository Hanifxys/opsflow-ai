package providers

import (
	"context"
	"fmt"
	"strings"

	"github.com/opsflow/ai-gateway/internal/domain"
	"github.com/opsflow/ai-gateway/internal/ports"
)

type MockProvider struct{}

func NewMockProvider() *MockProvider {
	return &MockProvider{}
}

func (p *MockProvider) Name() string {
	return "mock"
}

func (p *MockProvider) GenerateContent(ctx context.Context, req ports.LLMRequest) (*ports.LLMResponse, error) {
	lastMsg := ""
	if len(req.Messages) > 0 {
		lastMsg = req.Messages[len(req.Messages)-1].Content
	}

	lower := strings.ToLower(lastMsg)

	// Simulate tool invocation intent
	if strings.Contains(lower, "restart") || strings.Contains(lower, "scale") {
		return &ports.LLMResponse{
			Content: "I can help restart that service, but it requires human approval.",
			ToolCalls: []domain.ToolCall{
				{
					ID:   "call_restart_1",
					Name: "restart_service",
					Arguments: map[string]interface{}{
						"service_name": "payment-service",
						"environment":  "production",
					},
				},
			},
		}, nil
	}

	if strings.Contains(lower, "status") || strings.Contains(lower, "health") {
		return &ports.LLMResponse{
			Content: "Checking system status...",
			ToolCalls: []domain.ToolCall{
				{
					ID:   "call_status_1",
					Name: "get_service_status",
					Arguments: map[string]interface{}{
						"service_name": "payment-service",
					},
				},
			},
		}, nil
	}

	return &ports.LLMResponse{
		Content: fmt.Sprintf("AI Assistant Response to: %s", lastMsg),
	}, nil
}
