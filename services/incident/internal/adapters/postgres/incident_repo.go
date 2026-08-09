package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/opsflow/incident-service/internal/domain"
	"github.com/opsflow/incident-service/internal/ports"
)

type PostgresIncidentRepository struct {
	pool *pgxpool.Pool
}

func NewIncidentRepository(pool *pgxpool.Pool) *PostgresIncidentRepository {
	return &PostgresIncidentRepository{pool: pool}
}

func (r *PostgresIncidentRepository) Create(ctx context.Context, inc *domain.Incident) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	query := `
		INSERT INTO incidents (id, incident_key, service_id, title, description, severity, status, assignee_id, started_at, resolved_at, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`
	_, err = tx.Exec(ctx, query,
		inc.ID, inc.IncidentKey, inc.ServiceID, inc.Title, inc.Description,
		inc.Severity, inc.Status, inc.AssigneeID, inc.StartedAt, inc.ResolvedAt,
		inc.CreatedBy, inc.CreatedAt, inc.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert incident: %w", err)
	}

	// Insert transactional outbox event for incident.created
	outboxQuery := `
		INSERT INTO outbox_events (id, aggregate_type, aggregate_id, event_type, payload, status, attempts, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	payloadBytes, _ := json.Marshal(map[string]interface{}{
		"incident_id":  inc.ID,
		"incident_key": inc.IncidentKey,
		"title":        inc.Title,
		"severity":     inc.Severity,
		"created_by":   inc.CreatedBy,
	})
	_, err = tx.Exec(ctx, outboxQuery, uuid.New(), "INCIDENT", inc.ID, "INCIDENT_CREATED", payloadBytes, "PENDING", 0, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("failed to insert outbox event: %w", err)
	}

	return tx.Commit(ctx)
}

func (r *PostgresIncidentRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Incident, error) {
	query := `
		SELECT id, incident_key, service_id, title, description, severity, status, assignee_id, started_at, resolved_at, created_by, created_at, updated_at
		FROM incidents
		WHERE id = $1
	`
	inc := &domain.Incident{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&inc.ID, &inc.IncidentKey, &inc.ServiceID, &inc.Title, &inc.Description,
		&inc.Severity, &inc.Status, &inc.AssigneeID, &inc.StartedAt, &inc.ResolvedAt,
		&inc.CreatedBy, &inc.CreatedAt, &inc.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrIncidentNotFound
		}
		return nil, fmt.Errorf("failed to get incident by id: %w", err)
	}
	return inc, nil
}

func (r *PostgresIncidentRepository) GetByKey(ctx context.Context, key string) (*domain.Incident, error) {
	query := `
		SELECT id, incident_key, service_id, title, description, severity, status, assignee_id, started_at, resolved_at, created_by, created_at, updated_at
		FROM incidents
		WHERE incident_key = $1
	`
	inc := &domain.Incident{}
	err := r.pool.QueryRow(ctx, query, key).Scan(
		&inc.ID, &inc.IncidentKey, &inc.ServiceID, &inc.Title, &inc.Description,
		&inc.Severity, &inc.Status, &inc.AssigneeID, &inc.StartedAt, &inc.ResolvedAt,
		&inc.CreatedBy, &inc.CreatedAt, &inc.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrIncidentNotFound
		}
		return nil, fmt.Errorf("failed to get incident by key: %w", err)
	}
	return inc, nil
}

func (r *PostgresIncidentRepository) List(ctx context.Context, filter ports.ListFilter) ([]*domain.Incident, int, error) {
	limit := filter.Limit
	if limit <= 0 {
		limit = 20
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	whereClause := "WHERE 1=1"
	args := []interface{}{}
	argIdx := 1

	if filter.Status != nil {
		whereClause += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, *filter.Status)
		argIdx++
	}
	if filter.Severity != nil {
		whereClause += fmt.Sprintf(" AND severity = $%d", argIdx)
		args = append(args, *filter.Severity)
		argIdx++
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM incidents %s", whereClause)
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count incidents: %w", err)
	}

	query := fmt.Sprintf(`
		SELECT id, incident_key, service_id, title, description, severity, status, assignee_id, started_at, resolved_at, created_by, created_at, updated_at
		FROM incidents
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list incidents: %w", err)
	}
	defer rows.Close()

	var res []*domain.Incident
	for rows.Next() {
		inc := &domain.Incident{}
		err := rows.Scan(
			&inc.ID, &inc.IncidentKey, &inc.ServiceID, &inc.Title, &inc.Description,
			&inc.Severity, &inc.Status, &inc.AssigneeID, &inc.StartedAt, &inc.ResolvedAt,
			&inc.CreatedBy, &inc.CreatedAt, &inc.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		res = append(res, inc)
	}

	return res, total, rows.Err()
}

func (r *PostgresIncidentRepository) Update(ctx context.Context, inc *domain.Incident) error {
	query := `
		UPDATE incidents
		SET title = $1, description = $2, severity = $3, status = $4, assignee_id = $5, resolved_at = $6, updated_at = $7
		WHERE id = $8
	`
	_, err := r.pool.Exec(ctx, query,
		inc.Title, inc.Description, inc.Severity, inc.Status, inc.AssigneeID, inc.ResolvedAt, time.Now().UTC(), inc.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update incident: %w", err)
	}
	return nil
}

func (r *PostgresIncidentRepository) AddEvent(ctx context.Context, ev *domain.IncidentEvent) error {
	metaJSON, err := json.Marshal(ev.Metadata)
	if err != nil {
		metaJSON = []byte("{}")
	}

	query := `
		INSERT INTO incident_events (id, incident_id, event_type, message, actor_id, metadata, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err = r.pool.Exec(ctx, query, ev.ID, ev.IncidentID, ev.EventType, ev.Message, ev.ActorID, metaJSON, ev.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to insert incident event: %w", err)
	}
	return nil
}

func (r *PostgresIncidentRepository) ListEvents(ctx context.Context, incidentID uuid.UUID) ([]*domain.IncidentEvent, error) {
	query := `
		SELECT id, incident_id, event_type, message, actor_id, metadata, created_at
		FROM incident_events
		WHERE incident_id = $1
		ORDER BY created_at ASC
	`
	rows, err := r.pool.Query(ctx, query, incidentID)
	if err != nil {
		return nil, fmt.Errorf("failed to query incident events: %w", err)
	}
	defer rows.Close()

	var events []*domain.IncidentEvent
	for rows.Next() {
		ev := &domain.IncidentEvent{}
		var metaBytes []byte
		err := rows.Scan(&ev.ID, &ev.IncidentID, &ev.EventType, &ev.Message, &ev.ActorID, &metaBytes, &ev.CreatedAt)
		if err != nil {
			return nil, err
		}
		if len(metaBytes) > 0 {
			_ = json.Unmarshal(metaBytes, &ev.Metadata)
		}
		events = append(events, ev)
	}

	return events, rows.Err()
}

func (r *PostgresIncidentRepository) AddComment(ctx context.Context, c *domain.IncidentComment) error {
	query := `
		INSERT INTO incident_comments (id, incident_id, author_id, content, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.pool.Exec(ctx, query, c.ID, c.IncidentID, c.AuthorID, c.Content, c.CreatedAt, c.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to insert comment: %w", err)
	}
	return nil
}

func (r *PostgresIncidentRepository) ListComments(ctx context.Context, incidentID uuid.UUID) ([]*domain.IncidentComment, error) {
	query := `
		SELECT id, incident_id, author_id, content, created_at, updated_at
		FROM incident_comments
		WHERE incident_id = $1
		ORDER BY created_at ASC
	`
	rows, err := r.pool.Query(ctx, query, incidentID)
	if err != nil {
		return nil, fmt.Errorf("failed to query comments: %w", err)
	}
	defer rows.Close()

	var comments []*domain.IncidentComment
	for rows.Next() {
		c := &domain.IncidentComment{}
		err := rows.Scan(&c.ID, &c.IncidentID, &c.AuthorID, &c.Content, &c.CreatedAt, &c.UpdatedAt)
		if err != nil {
			return nil, err
		}
		comments = append(comments, c)
	}

	return comments, rows.Err()
}
