package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/opsflow/registry-service/internal/domain"
	"github.com/redis/go-redis/v9"
)

type ServiceCache struct {
	client *redis.Client
	logger *slog.Logger
	ttl    time.Duration
}

func NewServiceCache(client *redis.Client, logger *slog.Logger) *ServiceCache {
	return &ServiceCache{
		client: client,
		logger: logger,
		ttl:    5 * time.Minute,
	}
}

func (c *ServiceCache) GetService(ctx context.Context, id uuid.UUID) (*domain.Service, bool) {
	if c.client == nil {
		return nil, false
	}

	key := fmt.Sprintf("cache:services:id:%s", id.String())
	val, err := c.client.Get(ctx, key).Result()
	if err != nil {
		return nil, false
	}

	var s domain.Service
	if err := json.Unmarshal([]byte(val), &s); err != nil {
		c.logger.Warn("failed to unmarshal cached service JSON", slog.String("key", key))
		return nil, false
	}

	c.logger.Info("cache hit for service", slog.String("service_id", id.String()))
	return &s, true
}

func (c *ServiceCache) SetService(ctx context.Context, s *domain.Service) {
	if c.client == nil || s == nil {
		return
	}

	key := fmt.Sprintf("cache:services:id:%s", s.ID.String())
	data, err := json.Marshal(s)
	if err != nil {
		return
	}

	if err := c.client.Set(ctx, key, data, c.ttl).Err(); err != nil {
		c.logger.Warn("failed to cache service", slog.String("key", key), slog.String("error", err.Error()))
	}
}

func (c *ServiceCache) InvalidateService(ctx context.Context, id uuid.UUID) {
	if c.client == nil {
		return
	}

	key := fmt.Sprintf("cache:services:id:%s", id.String())
	_ = c.client.Del(ctx, key, "cache:services:list").Err()
	c.logger.Info("invalidated service cache", slog.String("service_id", id.String()))
}
