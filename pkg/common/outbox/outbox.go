package outbox

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/opsflow/common/rabbitmq"
)

type EventStatus string

const (
	StatusPending   EventStatus = "PENDING"
	StatusPublished EventStatus = "PUBLISHED"
	StatusFailed    EventStatus = "FAILED"
)

type OutboxEvent struct {
	ID            uuid.UUID       `json:"id"`
	AggregateType string          `json:"aggregate_type"`
	AggregateID   uuid.UUID       `json:"aggregate_id"`
	EventType     string          `json:"event_type"`
	Payload       json.RawMessage `json:"payload"`
	Status        EventStatus     `json:"status"`
	Attempts      int             `json:"attempts"`
	PublishedAt   *time.Time      `json:"published_at,omitempty"`
	CreatedAt     time.Time       `json:"created_at"`
}

// SaveTransactional inserts an outbox record inside an active database transaction.
func SaveTransactional(ctx context.Context, tx pgx.Tx, aggregateType string, aggregateID uuid.UUID, eventType string, payload any) error {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal outbox payload: %w", err)
	}

	query := `
		INSERT INTO outbox_events (id, aggregate_type, aggregate_id, event_type, payload, status, attempts, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	now := time.Now().UTC()
	_, err = tx.Exec(ctx, query, uuid.New(), aggregateType, aggregateID, eventType, payloadBytes, StatusPending, 0, now)
	if err != nil {
		return fmt.Errorf("failed to insert outbox event: %w", err)
	}
	return nil
}

type Publisher struct {
	pool    *pgxpool.Pool
	rabbit  *rabbitmq.Client
	logger  *slog.Logger
	stopCh  chan struct{}
}

func NewPublisher(pool *pgxpool.Pool, rabbit *rabbitmq.Client, logger *slog.Logger) *Publisher {
	return &Publisher{
		pool:   pool,
		rabbit: rabbit,
		logger: logger,
		stopCh: make(chan struct{}),
	}
}

func (p *Publisher) Start(pollInterval time.Duration) {
	ticker := time.NewTicker(pollInterval)
	go func() {
		for {
			select {
			case <-ticker.C:
				p.processPendingEvents()
			case <-p.stopCh:
				ticker.Stop()
				return
			}
		}
	}()
}

func (p *Publisher) Stop() {
	close(p.stopCh)
}

func (p *Publisher) processPendingEvents() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	query := `
		SELECT id, aggregate_type, aggregate_id, event_type, payload, attempts
		FROM outbox_events
		WHERE status = 'PENDING' AND attempts < 5
		ORDER BY created_at ASC
		LIMIT 20
		FOR UPDATE SKIP LOCKED
	`

	rows, err := p.pool.Query(ctx, query)
	if err != nil {
		p.logger.Error("failed to fetch outbox events", slog.String("error", err.Error()))
		return
	}
	defer rows.Close()

	var events []*OutboxEvent
	for rows.Next() {
		ev := &OutboxEvent{}
		if err := rows.Scan(&ev.ID, &ev.AggregateType, &ev.AggregateID, &ev.EventType, &ev.Payload, &ev.Attempts); err != nil {
			p.logger.Error("failed to scan outbox event", slog.String("error", err.Error()))
			continue
		}
		events = append(events, ev)
	}

	for _, ev := range events {
		routingKey := formatRoutingKey(ev.EventType)
		if err := p.rabbit.Publish(ctx, routingKey, ev.Payload); err != nil {
			p.logger.Error("failed to publish outbox event",
				slog.String("event_id", ev.ID.String()),
				slog.String("routing_key", routingKey),
				slog.String("error", err.Error()),
			)
			p.incrementAttempt(ctx, ev.ID, ev.Attempts+1)
			continue
		}

		p.markPublished(ctx, ev.ID)
		p.logger.Info("published outbox event to rabbitmq",
			slog.String("event_id", ev.ID.String()),
			slog.String("routing_key", routingKey),
		)
	}
}

func (p *Publisher) markPublished(ctx context.Context, id uuid.UUID) {
	now := time.Now().UTC()
	query := `UPDATE outbox_events SET status = 'PUBLISHED', published_at = $1 WHERE id = $2`
	_, _ = p.pool.Exec(ctx, query, now, id)
}

func (p *Publisher) incrementAttempt(ctx context.Context, id uuid.UUID, attempts int) {
	status := StatusPending
	if attempts >= 5 {
		status = StatusFailed
	}
	query := `UPDATE outbox_events SET attempts = $1, status = $2 WHERE id = $3`
	_, _ = p.pool.Exec(ctx, query, attempts, status, id)
}

func formatRoutingKey(eventType string) string {
	switch eventType {
	case "INCIDENT_CREATED":
		return "incident.created"
	case "INCIDENT_RESOLVED":
		return "incident.resolved"
	case "SERVICE_HEALTH_CHANGED":
		return "service.health_changed"
	default:
		return "notification.requested"
	}
}
