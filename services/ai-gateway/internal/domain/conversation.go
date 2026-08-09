package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Conversation struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Messages  []Message `json:"messages,omitempty"`
}

type Message struct {
	ID             uuid.UUID       `json:"id"`
	ConversationID uuid.UUID       `json:"conversation_id"`
	Role           string          `json:"role"` // user, assistant, system, tool
	Content        string          `json:"content"`
	ToolCalls      json.RawMessage `json:"tool_calls,omitempty"`
	CreatedAt      time.Time       `json:"created_at"`
}

type ApprovalStatus string

const (
	ApprovalStatusPending  ApprovalStatus = "PENDING"
	ApprovalStatusApproved ApprovalStatus = "APPROVED"
	ApprovalStatusRejected ApprovalStatus = "REJECTED"
)

type ApprovalRequest struct {
	ID             uuid.UUID       `json:"id"`
	ConversationID uuid.UUID       `json:"conversation_id"`
	ToolName       string          `json:"tool_name"`
	Arguments      json.RawMessage `json:"arguments"`
	Status         ApprovalStatus  `json:"status"`
	RequestedBy    uuid.UUID       `json:"requested_by"`
	ApprovedBy     *uuid.UUID      `json:"approved_by,omitempty"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
}

type KnowledgeDocument struct {
	ID        uuid.UUID              `json:"id"`
	Title     string                 `json:"title"`
	Content   string                 `json:"content"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
}
