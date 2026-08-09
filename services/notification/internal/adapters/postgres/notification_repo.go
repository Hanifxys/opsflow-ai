package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/opsflow/notification-service/internal/domain"
)

type PostgresNotificationRepository struct {
	pool *pgxpool.Pool
}

func NewNotificationRepository(pool *pgxpool.Pool) *PostgresNotificationRepository {
	return &PostgresNotificationRepository{pool: pool}
}

func (r *PostgresNotificationRepository) SaveIfNew(ctx context.Context, n *domain.Notification) (bool, error) {
	query := `
		INSERT INTO notifications (id, user_id, channel, event_type, payload, status, attempts, idempotency_key, sent_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (idempotency_key) DO NOTHING
	`
	cmdTag, err := r.pool.Exec(ctx, query,
		n.ID, n.UserID, n.Channel, n.EventType, n.Payload, n.Status, n.Attempts, n.IdempotencyKey, n.SentAt, n.CreatedAt,
	)
	if err != nil {
		return false, fmt.Errorf("failed to insert notification: %w", err)
	}

	// RowsAffected == 1 means it was newly inserted, 0 means duplicate ignored (idempotent)
	return cmdTag.RowsAffected() > 0, nil
}

func (r *PostgresNotificationRepository) ListRecent(ctx context.Context, limit int) ([]*domain.Notification, error) {
	if limit <= 0 {
		limit = 20
	}
	query := `
		SELECT id, user_id, channel, event_type, payload, status, attempts, idempotency_key, sent_at, created_at
		FROM notifications
		ORDER BY created_at DESC
		LIMIT $1
	`
	rows, err := r.pool.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query notifications: %w", err)
	}
	defer rows.Close()

	var list []*domain.Notification
	for rows.Next() {
		n := &domain.Notification{}
		err := rows.Scan(&n.ID, &n.UserID, &n.Channel, &n.EventType, &n.Payload, &n.Status, &n.Attempts, &n.IdempotencyKey, &n.SentAt, &n.CreatedAt)
		if err != nil {
			return nil, err
		}
		list = append(list, n)
	}

	return list, rows.Err()
}
