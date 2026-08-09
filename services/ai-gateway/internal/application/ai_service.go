package application

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/opsflow/ai-gateway/internal/adapters/tools"
	"github.com/opsflow/ai-gateway/internal/domain"
	"github.com/opsflow/ai-gateway/internal/ports"
)

type AIService struct {
	repo   ports.AIRepository
	router *ModelRouter
	tools  *tools.ToolRegistry
}

func NewAIService(repo ports.AIRepository, router *ModelRouter, toolReg *tools.ToolRegistry) *AIService {
	return &AIService{
		repo:   repo,
		router: router,
		tools:  toolReg,
	}
}

func (s *AIService) CreateConversation(ctx context.Context, userID uuid.UUID, title string) (*domain.Conversation, error) {
	if title == "" {
		title = "OpsFlow Pro AI Session"
	}
	now := time.Now().UTC()
	conv := &domain.Conversation{
		ID:        uuid.New(),
		UserID:    userID,
		Title:     title,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.repo.CreateConversation(ctx, conv); err != nil {
		return nil, err
	}
	return conv, nil
}

func (s *AIService) SendMessage(ctx context.Context, convID, userID uuid.UUID, content string) (*domain.Message, []*domain.ApprovalRequest, error) {
	conv, err := s.repo.GetConversationByID(ctx, convID)
	if err != nil {
		return nil, nil, err
	}

	now := time.Now().UTC()
	userMsg := &domain.Message{
		ID:             uuid.New(),
		ConversationID: conv.ID,
		Role:           "user",
		Content:        content,
		CreatedAt:      now,
	}
	if err := s.repo.SaveMessage(ctx, userMsg); err != nil {
		return nil, nil, err
	}

	history, err := s.repo.ListMessagesByConversationID(ctx, convID)
	if err != nil {
		return nil, nil, err
	}

	// Fine-tuned Senior L3 DevOps / Reliability Engineer Persona & System Context Injection
	systemContext := "SYSTEM PERSONA: You are OpsFlow Pro AI, a Senior L3 Infrastructure & Reliability Engineer fine-tuned for high-availability systems, Kubernetes, PostgreSQL tuning, and incident remediation.\n"

	// Search RAG knowledge base for operational context
	docs, _ := s.repo.SearchKnowledge(ctx, content, 2)
	if len(docs) > 0 {
		systemContext += "\nRELEVANT KNOWLEDGE BASE CONTEXT:\n"
		for _, d := range docs {
			systemContext += fmt.Sprintf("- [%s]: %s\n", d.Title, d.Content)
		}
	}

	llmMsgs := make([]domain.Message, 0, len(history)+1)
	llmMsgs = append(llmMsgs, domain.Message{Role: "system", Content: systemContext})
	llmMsgs = append(llmMsgs, history...)

	// Route prompt to LLM provider
	llmResp, err := s.router.Route(ctx, ports.LLMRequest{
		Messages: llmMsgs,
		Tools:    s.tools.Definitions(),
	})
	if err != nil {
		return nil, nil, fmt.Errorf("llm generate error: %w", err)
	}

	var createdApprovals []*domain.ApprovalRequest

	// Handle tool calls if returned by LLM
	for _, tc := range llmResp.ToolCalls {
		if s.tools.IsSensitive(tc.Name) {
			// Sensitive mutation -> Create Human Approval Request (PENDING)
			argsBytes, _ := json.Marshal(tc.Arguments)
			appReq := &domain.ApprovalRequest{
				ID:             uuid.New(),
				ConversationID: convID,
				ToolName:       tc.Name,
				Arguments:      argsBytes,
				Status:         domain.ApprovalStatusPending,
				RequestedBy:    userID,
				CreatedAt:      now,
				UpdatedAt:      now,
			}
			_ = s.repo.CreateApprovalRequest(ctx, appReq)
			createdApprovals = append(createdApprovals, appReq)
		} else {
			// Read-Only tool -> Execute immediately
			output, _ := s.tools.ExecuteReadOnlyTool(ctx, tc.Name, tc.Arguments)
			toolMsg := &domain.Message{
				ID:             uuid.New(),
				ConversationID: convID,
				Role:           "tool",
				Content:        output,
				CreatedAt:      time.Now().UTC(),
			}
			_ = s.repo.SaveMessage(ctx, toolMsg)
		}
	}

	assistantContent := llmResp.Content
	if len(createdApprovals) > 0 {
		assistantContent += fmt.Sprintf("\n[HUMAN APPROVAL REQUIRED] Proposed action '%s' requires approval. Approval Request ID: %s", createdApprovals[0].ToolName, createdApprovals[0].ID.String())
	}

	toolCallsJSON, _ := json.Marshal(llmResp.ToolCalls)
	assistantMsg := &domain.Message{
		ID:             uuid.New(),
		ConversationID: convID,
		Role:           "assistant",
		Content:        assistantContent,
		ToolCalls:      toolCallsJSON,
		CreatedAt:      time.Now().UTC(),
	}
	if err := s.repo.SaveMessage(ctx, assistantMsg); err != nil {
		return nil, nil, err
	}

	return assistantMsg, createdApprovals, nil
}

func (s *AIService) ApproveToolAction(ctx context.Context, approvalID, approverID uuid.UUID) (string, error) {
	app, err := s.repo.GetApprovalRequestByID(ctx, approvalID)
	if err != nil {
		return "", err
	}

	if err := s.repo.UpdateApprovalStatus(ctx, approvalID, domain.ApprovalStatusApproved, approverID); err != nil {
		return "", err
	}

	var args map[string]interface{}
	_ = json.Unmarshal(app.Arguments, &args)

	// Execute approved mutation
	output, err := s.tools.ExecuteApprovedMutation(ctx, app.ToolName, args)
	if err != nil {
		return "", fmt.Errorf("failed to execute approved tool: %w", err)
	}

	// Persist result into conversation
	toolMsg := &domain.Message{
		ID:             uuid.New(),
		ConversationID: app.ConversationID,
		Role:           "tool",
		Content:        output,
		CreatedAt:      time.Now().UTC(),
	}
	_ = s.repo.SaveMessage(ctx, toolMsg)

	return output, nil
}

func (s *AIService) RejectToolAction(ctx context.Context, approvalID, approverID uuid.UUID) error {
	return s.repo.UpdateApprovalStatus(ctx, approvalID, domain.ApprovalStatusRejected, approverID)
}

func (s *AIService) ListPendingApprovals(ctx context.Context) ([]*domain.ApprovalRequest, error) {
	return s.repo.ListPendingApprovals(ctx)
}

func (s *AIService) IngestKnowledge(ctx context.Context, title, content string, metadata map[string]interface{}) (*domain.KnowledgeDocument, error) {
	doc := &domain.KnowledgeDocument{
		ID:        uuid.New(),
		Title:     title,
		Content:   content,
		Metadata:  metadata,
		CreatedAt: time.Now().UTC(),
	}
	if err := s.repo.SaveKnowledgeDocument(ctx, doc); err != nil {
		return nil, err
	}
	return doc, nil
}

func (s *AIService) ListConversations(ctx context.Context, userID uuid.UUID) ([]*domain.Conversation, error) {
	return s.repo.ListConversationsByUserID(ctx, userID)
}

func (s *AIService) GetMessages(ctx context.Context, convID uuid.UUID) ([]domain.Message, error) {
	return s.repo.ListMessagesByConversationID(ctx, convID)
}
