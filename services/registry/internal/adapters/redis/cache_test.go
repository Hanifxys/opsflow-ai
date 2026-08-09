package redis_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/opsflow/common/logging"
	"github.com/opsflow/registry-service/internal/adapters/redis"
	"github.com/opsflow/registry-service/internal/domain"
)

func TestServiceCache_FallbackWhenNilClient(t *testing.T) {
	logger := logging.New("info")
	cache := redis.NewServiceCache(nil, logger)
	ctx := context.Background()

	id := uuid.New()
	svc := &domain.Service{
		ID:        id,
		Name:      "test-service",
		OwnerTeam: "Core",
	}

	// 1. SetService with nil client should not panic
	cache.SetService(ctx, svc)

	// 2. GetService with nil client should return false cleanly
	cached, ok := cache.GetService(ctx, id)
	if ok || cached != nil {
		t.Errorf("expected cache miss when client is nil, got ok=%v", ok)
	}

	// 3. InvalidateService with nil client should not panic
	cache.InvalidateService(ctx, id)
}
