package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/opsflow/ai-gateway/internal/domain"
)

type PostgresAIRepository struct {
	pool *pgxpool.Pool
}

func NewAIRepository(pool *pgxpool.Pool) *PostgresAIRepository {
	return &PostgresAIRepository{pool: pool}
}

func (r *PostgresAIRepository) CreateConversation(ctx context.Context, conv *domain.Conversation) error {
	query := `
		INSERT INTO ai_conversations (id, user_id, title, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := r.pool.Exec(ctx, query, conv.ID, conv.UserID, conv.Title, conv.CreatedAt, conv.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to insert conversation: %w", err)
	}
	return nil
}

func (r *PostgresAIRepository) GetConversationByID(ctx context.Context, id uuid.UUID) (*domain.Conversation, error) {
	query := `
		SELECT id, user_id, title, created_at, updated_at
		FROM ai_conversations
		WHERE id = $1
	`
	conv := &domain.Conversation{}
	err := r.pool.QueryRow(ctx, query, id).Scan(&conv.ID, &conv.UserID, &conv.Title, &conv.CreatedAt, &conv.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrConversationNotFound
		}
		return nil, fmt.Errorf("failed to query conversation: %w", err)
	}
	return conv, nil
}

func (r *PostgresAIRepository) ListConversationsByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.Conversation, error) {
	query := `
		SELECT id, user_id, title, created_at, updated_at
		FROM ai_conversations
		WHERE user_id = $1
		ORDER BY updated_at DESC
	`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query conversations: %w", err)
	}
	defer rows.Close()

	var list []*domain.Conversation
	for rows.Next() {
		c := &domain.Conversation{}
		if err := rows.Scan(&c.ID, &c.UserID, &c.Title, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, c)
	}
	return list, rows.Err()
}

func (r *PostgresAIRepository) SaveMessage(ctx context.Context, msg *domain.Message) error {
	query := `
		INSERT INTO ai_messages (id, conversation_id, role, content, tool_calls, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	toolCallsJSON := msg.ToolCalls
	if toolCallsJSON == nil {
		toolCallsJSON = []byte("[]")
	}
	_, err := r.pool.Exec(ctx, query, msg.ID, msg.ConversationID, msg.Role, msg.Content, toolCallsJSON, msg.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to insert message: %w", err)
	}

	// Update conversation updated_at
	_, _ = r.pool.Exec(ctx, "UPDATE ai_conversations SET updated_at = NOW() WHERE id = $1", msg.ConversationID)
	return nil
}

func (r *PostgresAIRepository) ListMessagesByConversationID(ctx context.Context, convID uuid.UUID) ([]domain.Message, error) {
	query := `
		SELECT id, conversation_id, role, content, tool_calls, created_at
		FROM ai_messages
		WHERE conversation_id = $1
		ORDER BY created_at ASC
	`
	rows, err := r.pool.Query(ctx, query, convID)
	if err != nil {
		return nil, fmt.Errorf("failed to query messages: %w", err)
	}
	defer rows.Close()

	var list []domain.Message
	for rows.Next() {
		var m domain.Message
		if err := rows.Scan(&m.ID, &m.ConversationID, &m.Role, &m.Content, &m.ToolCalls, &m.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, m)
	}
	return list, rows.Err()
}

func (r *PostgresAIRepository) CreateApprovalRequest(ctx context.Context, app *domain.ApprovalRequest) error {
	query := `
		INSERT INTO ai_approvals (id, conversation_id, tool_name, arguments, status, requested_by, approved_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := r.pool.Exec(ctx, query, app.ID, app.ConversationID, app.ToolName, app.Arguments, app.Status, app.RequestedBy, app.ApprovedBy, app.CreatedAt, app.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to insert approval request: %w", err)
	}
	return nil
}

func (r *PostgresAIRepository) GetApprovalRequestByID(ctx context.Context, id uuid.UUID) (*domain.ApprovalRequest, error) {
	query := `
		SELECT id, conversation_id, tool_name, arguments, status, requested_by, approved_by, created_at, updated_at
		FROM ai_approvals
		WHERE id = $1
	`
	app := &domain.ApprovalRequest{}
	err := r.pool.QueryRow(ctx, query, id).Scan(&app.ID, &app.ConversationID, &app.ToolName, &app.Arguments, &app.Status, &app.RequestedBy, &app.ApprovedBy, &app.CreatedAt, &app.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrApprovalNotFound
		}
		return nil, fmt.Errorf("failed to query approval: %w", err)
	}
	return app, nil
}

func (r *PostgresAIRepository) UpdateApprovalStatus(ctx context.Context, id uuid.UUID, status domain.ApprovalStatus, approvedBy uuid.UUID) error {
	query := `
		UPDATE ai_approvals
		SET status = $1, approved_by = $2, updated_at = NOW()
		WHERE id = $3 AND status = 'PENDING'
	`
	cmdTag, err := r.pool.Exec(ctx, query, status, approvedBy, id)
	if err != nil {
		return fmt.Errorf("failed to update approval: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return domain.ErrApprovalNotPending
	}
	return nil
}

func (r *PostgresAIRepository) ListPendingApprovals(ctx context.Context) ([]*domain.ApprovalRequest, error) {
	query := `
		SELECT id, conversation_id, tool_name, arguments, status, requested_by, approved_by, created_at, updated_at
		FROM ai_approvals
		WHERE status = 'PENDING'
		ORDER BY created_at ASC
	`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query pending approvals: %w", err)
	}
	defer rows.Close()

	var list []*domain.ApprovalRequest
	for rows.Next() {
		app := &domain.ApprovalRequest{}
		if err := rows.Scan(&app.ID, &app.ConversationID, &app.ToolName, &app.Arguments, &app.Status, &app.RequestedBy, &app.ApprovedBy, &app.CreatedAt, &app.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, app)
	}
	return list, rows.Err()
}

func (r *PostgresAIRepository) SaveKnowledgeDocument(ctx context.Context, doc *domain.KnowledgeDocument) error {
	query := `
		INSERT INTO knowledge_documents (id, title, content, metadata, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	metaBytes, _ := json.Marshal(doc.Metadata)
	_, err := r.pool.Exec(ctx, query, doc.ID, doc.Title, doc.Content, metaBytes, doc.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to insert knowledge document: %w", err)
	}
	return nil
}

func (r *PostgresAIRepository) SearchKnowledge(ctx context.Context, queryStr string, limit int) ([]*domain.KnowledgeDocument, error) {
	if limit <= 0 {
		limit = 5
	}
	query := `
		SELECT id, title, content, metadata, created_at
		FROM knowledge_documents
		WHERE title ILIKE '%' || $1 || '%' OR content ILIKE '%' || $1 || '%'
		ORDER BY created_at DESC
		LIMIT $2
	`
	rows, err := r.pool.Query(ctx, query, queryStr, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query knowledge: %w", err)
	}
	defer rows.Close()

	var list []*domain.KnowledgeDocument
	for rows.Next() {
		doc := &domain.KnowledgeDocument{}
		var metaBytes []byte
		if err := rows.Scan(&doc.ID, &doc.Title, &doc.Content, &metaBytes, &doc.CreatedAt); err != nil {
			return nil, err
		}
		_ = json.Unmarshal(metaBytes, &doc.Metadata)
		list = append(list, doc)
	}
	return list, rows.Err()
}
