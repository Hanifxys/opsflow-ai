package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/opsflow/ai-gateway/internal/domain"
)

type AIRepository interface {
	CreateConversation(ctx context.Context, conv *domain.Conversation) error
	GetConversationByID(ctx context.Context, id uuid.UUID) (*domain.Conversation, error)
	ListConversationsByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.Conversation, error)

	SaveMessage(ctx context.Context, msg *domain.Message) error
	ListMessagesByConversationID(ctx context.Context, convID uuid.UUID) ([]domain.Message, error)

	CreateApprovalRequest(ctx context.Context, app *domain.ApprovalRequest) error
	GetApprovalRequestByID(ctx context.Context, id uuid.UUID) (*domain.ApprovalRequest, error)
	UpdateApprovalStatus(ctx context.Context, id uuid.UUID, status domain.ApprovalStatus, approvedBy uuid.UUID) error
	ListPendingApprovals(ctx context.Context) ([]*domain.ApprovalRequest, error)

	SaveKnowledgeDocument(ctx context.Context, doc *domain.KnowledgeDocument) error
	SearchKnowledge(ctx context.Context, query string, limit int) ([]*domain.KnowledgeDocument, error)
}
