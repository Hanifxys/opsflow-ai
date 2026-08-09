package application_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/opsflow/ai-gateway/internal/adapters/providers"
	"github.com/opsflow/ai-gateway/internal/adapters/tools"
	"github.com/opsflow/ai-gateway/internal/application"
	"github.com/opsflow/ai-gateway/internal/domain"
)

type mockAIRepository struct {
	conversations map[uuid.UUID]*domain.Conversation
	messages      map[uuid.UUID][]domain.Message
	approvals     map[uuid.UUID]*domain.ApprovalRequest
	knowledge     []*domain.KnowledgeDocument
}

func newMockAIRepository() *mockAIRepository {
	return &mockAIRepository{
		conversations: make(map[uuid.UUID]*domain.Conversation),
		messages:      make(map[uuid.UUID][]domain.Message),
		approvals:     make(map[uuid.UUID]*domain.ApprovalRequest),
	}
}

func (m *mockAIRepository) CreateConversation(ctx context.Context, conv *domain.Conversation) error {
	m.conversations[conv.ID] = conv
	return nil
}

func (m *mockAIRepository) GetConversationByID(ctx context.Context, id uuid.UUID) (*domain.Conversation, error) {
	conv, ok := m.conversations[id]
	if !ok {
		return nil, domain.ErrConversationNotFound
	}
	return conv, nil
}

func (m *mockAIRepository) ListConversationsByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.Conversation, error) {
	var res []*domain.Conversation
	for _, c := range m.conversations {
		if c.UserID == userID {
			res = append(res, c)
		}
	}
	return res, nil
}

func (m *mockAIRepository) SaveMessage(ctx context.Context, msg *domain.Message) error {
	m.messages[msg.ConversationID] = append(m.messages[msg.ConversationID], *msg)
	return nil
}

func (m *mockAIRepository) ListMessagesByConversationID(ctx context.Context, convID uuid.UUID) ([]domain.Message, error) {
	return m.messages[convID], nil
}

func (m *mockAIRepository) CreateApprovalRequest(ctx context.Context, app *domain.ApprovalRequest) error {
	m.approvals[app.ID] = app
	return nil
}

func (m *mockAIRepository) GetApprovalRequestByID(ctx context.Context, id uuid.UUID) (*domain.ApprovalRequest, error) {
	app, ok := m.approvals[id]
	if !ok {
		return nil, domain.ErrApprovalNotFound
	}
	return app, nil
}

func (m *mockAIRepository) UpdateApprovalStatus(ctx context.Context, id uuid.UUID, status domain.ApprovalStatus, approvedBy uuid.UUID) error {
	app, ok := m.approvals[id]
	if !ok {
		return domain.ErrApprovalNotFound
	}
	if app.Status != domain.ApprovalStatusPending {
		return domain.ErrApprovalNotPending
	}
	app.Status = status
	app.ApprovedBy = &approvedBy
	return nil
}

func (m *mockAIRepository) ListPendingApprovals(ctx context.Context) ([]*domain.ApprovalRequest, error) {
	var list []*domain.ApprovalRequest
	for _, a := range m.approvals {
		if a.Status == domain.ApprovalStatusPending {
			list = append(list, a)
		}
	}
	return list, nil
}

func (m *mockAIRepository) SaveKnowledgeDocument(ctx context.Context, doc *domain.KnowledgeDocument) error {
	m.knowledge = append(m.knowledge, doc)
	return nil
}

func (m *mockAIRepository) SearchKnowledge(ctx context.Context, query string, limit int) ([]*domain.KnowledgeDocument, error) {
	return m.knowledge, nil
}

func TestAIService_HumanApprovalWorkflow(t *testing.T) {
	repo := newMockAIRepository()
	mockP := providers.NewMockProvider()
	router := application.NewModelRouter("mock", mockP)
	toolReg := tools.NewToolRegistry()
	svc := application.NewAIService(repo, router, toolReg)

	ctx := context.Background()
	userID := uuid.New()

	// 1. Start Conversation
	conv, err := svc.CreateConversation(ctx, userID, "Incident Mitigation")
	if err != nil {
		t.Fatalf("failed to create conversation: %v", err)
	}

	// 2. Send Mutation Prompt requiring approval ("restart payment-service")
	msg, approvals, err := svc.SendMessage(ctx, conv.ID, userID, "Please restart payment-service")
	if err != nil {
		t.Fatalf("failed to send message: %v", err)
	}
	if msg == nil {
		t.Fatal("expected assistant message response")
	}
	if len(approvals) != 1 {
		t.Fatalf("expected 1 PENDING human approval request, got %d", len(approvals))
	}

	approvalID := approvals[0].ID
	if approvals[0].Status != domain.ApprovalStatusPending {
		t.Errorf("expected status PENDING, got %s", approvals[0].Status)
	}

	// 3. Human Approves Action
	approverID := uuid.New()
	output, err := svc.ApproveToolAction(ctx, approvalID, approverID)
	if err != nil {
		t.Fatalf("failed to approve tool action: %v", err)
	}
	if output == "" {
		t.Error("expected non-empty tool execution output")
	}

	// 4. Verify Approval status updated to APPROVED
	app, _ := repo.GetApprovalRequestByID(ctx, approvalID)
	if app.Status != domain.ApprovalStatusApproved {
		t.Errorf("expected status APPROVED, got %s", app.Status)
	}
}

func TestAIService_RAGKnowledgeIngestion(t *testing.T) {
	repo := newMockAIRepository()
	mockP := providers.NewMockProvider()
	router := application.NewModelRouter("mock", mockP)
	toolReg := tools.NewToolRegistry()
	svc := application.NewAIService(repo, router, toolReg)

	ctx := context.Background()

	// 1. Ingest Knowledge Document
	doc, err := svc.IngestKnowledge(ctx, "Runbook Payment Service", "To resolve high latency, verify DB connection pool settings.", map[string]interface{}{"author": "SRE"})
	if err != nil {
		t.Fatalf("failed to ingest knowledge: %v", err)
	}
	if doc.ID == uuid.Nil {
		t.Error("expected valid doc UUID")
	}

	// 2. Send Message -> verifies RAG docs retrieved
	conv, _ := svc.CreateConversation(ctx, uuid.New(), "RAG Test")
	msg, _, err := svc.SendMessage(ctx, conv.ID, uuid.New(), "How to fix payment service latency?")
	if err != nil {
		t.Fatalf("failed to send message: %v", err)
	}
	if msg == nil {
		t.Error("expected message response")
	}
}
