package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Notification struct {
	ID             uuid.UUID       `json:"id"`
	UserID         *uuid.UUID      `json:"user_id,omitempty"`
	Channel        string          `json:"channel"`
	EventType      string          `json:"event_type"`
	Payload        json.RawMessage `json:"payload"`
	Status         string          `json:"status"`
	Attempts       int             `json:"attempts"`
	IdempotencyKey string          `json:"idempotency_key"`
	SentAt         time.Time       `json:"sent_at"`
	CreatedAt      time.Time       `json:"created_at"`
}
