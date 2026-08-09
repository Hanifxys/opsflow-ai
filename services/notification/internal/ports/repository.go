package ports

import (
	"context"

	"github.com/opsflow/notification-service/internal/domain"
)

type NotificationRepository interface {
	SaveIfNew(ctx context.Context, n *domain.Notification) (bool, error)
	ListRecent(ctx context.Context, limit int) ([]*domain.Notification, error)
}
