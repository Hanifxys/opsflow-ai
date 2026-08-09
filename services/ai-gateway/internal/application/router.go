package application

import (
	"context"

	"github.com/opsflow/ai-gateway/internal/ports"
)

type ModelRouter struct {
	providers map[string]ports.LLMProvider
	defaultP  string
}

func NewModelRouter(defaultProvider string, providers ...ports.LLMProvider) *ModelRouter {
	pMap := make(map[string]ports.LLMProvider)
	for _, p := range providers {
		pMap[p.Name()] = p
	}
	return &ModelRouter{
		providers: pMap,
		defaultP:  defaultProvider,
	}
}

func (r *ModelRouter) Route(ctx context.Context, req ports.LLMRequest) (*ports.LLMResponse, error) {
	providerName := r.defaultP
	if req.Model == "mock" || req.Model == "ollama" || req.Model == "cloud" {
		providerName = req.Model
	}

	p, exists := r.providers[providerName]
	if !exists {
		// Fallback to mock provider
		if mockP, ok := r.providers["mock"]; ok {
			return mockP.GenerateContent(ctx, req)
		}
	}

	resp, err := p.GenerateContent(ctx, req)
	if err != nil {
		// Resilience: Fallback to mock provider on error
		if mockP, ok := r.providers["mock"]; ok && p.Name() != "mock" {
			return mockP.GenerateContent(ctx, req)
		}
		return nil, err
	}

	return resp, nil
}
