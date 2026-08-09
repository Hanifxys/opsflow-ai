package application

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/opsflow/notification-service/internal/domain"
	"github.com/opsflow/notification-service/internal/ports"
)

type NotificationService struct {
	repo ports.NotificationRepository
}

func NewNotificationService(repo ports.NotificationRepository) *NotificationService {
	return &NotificationService{repo: repo}
}

func (s *NotificationService) ProcessEvent(ctx context.Context, channel, eventType string, rawPayload []byte) (bool, error) {
	// Generate unique idempotency key from event type + payload content hash
	hash := sha256.Sum256(rawPayload)
	idempotencyKey := fmt.Sprintf("%s:%s", eventType, hex.EncodeToString(hash[:16]))

	now := time.Now().UTC()
	n := &domain.Notification{
		ID:             uuid.New(),
		Channel:        channel,
		EventType:      eventType,
		Payload:        json.RawMessage(rawPayload),
		Status:         "SENT",
		Attempts:       1,
		IdempotencyKey: idempotencyKey,
		SentAt:         now,
		CreatedAt:      now,
	}

	// Try extracting user_id or actor_id from payload
	var meta map[string]interface{}
	if err := json.Unmarshal(rawPayload, &meta); err == nil {
		if uidStr, ok := meta["created_by"].(string); ok {
			if uid, err := uuid.Parse(uidStr); err == nil {
				n.UserID = &uid
			}
		}
	}

	isNew, err := s.repo.SaveIfNew(ctx, n)
	if err != nil {
		return false, fmt.Errorf("failed to save notification: %w", err)
	}

	return isNew, nil
}

func (s *NotificationService) ListRecentNotifications(ctx context.Context, limit int) ([]*domain.Notification, error) {
	return s.repo.ListRecent(ctx, limit)
}
