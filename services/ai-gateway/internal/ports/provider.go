package ports

import (
	"context"

	"github.com/opsflow/ai-gateway/internal/domain"
)

type LLMRequest struct {
	Model     string
	Messages  []domain.Message
	Tools     []domain.ToolDefinition
	MaxTokens int
}

type LLMResponse struct {
	Content   string
	ToolCalls []domain.ToolCall
}

type LLMProvider interface {
	Name() string
	GenerateContent(ctx context.Context, req LLMRequest) (*LLMResponse, error)
}
