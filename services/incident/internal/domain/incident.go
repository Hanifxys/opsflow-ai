package domain

import (
	"time"

	"github.com/google/uuid"
)

type Severity string

const (
	SeveritySEV1 Severity = "SEV1" // Critical
	SeveritySEV2 Severity = "SEV2" // Major
	SeveritySEV3 Severity = "SEV3" // Minor
	SeveritySEV4 Severity = "SEV4" // Low
)

func (s Severity) IsValid() bool {
	switch s {
	case SeveritySEV1, SeveritySEV2, SeveritySEV3, SeveritySEV4:
		return true
	}
	return false
}

type Status string

const (
	StatusOpen          Status = "OPEN"
	StatusInvestigating Status = "INVESTIGATING"
	StatusMitigating    Status = "MITIGATING"
	StatusResolved      Status = "RESOLVED"
	StatusClosed        Status = "CLOSED"
)

func (s Status) IsValid() bool {
	switch s {
	case StatusOpen, StatusInvestigating, StatusMitigating, StatusResolved, StatusClosed:
		return true
	}
	return false
}

// CanTransitionTo enforces the Incident State Machine rules (design.md § 4).
func (s Status) CanTransitionTo(target Status) bool {
	if s == target {
		return true
	}

	switch s {
	case StatusOpen:
		return target == StatusInvestigating || target == StatusMitigating || target == StatusResolved
	case StatusInvestigating:
		return target == StatusMitigating || target == StatusResolved
	case StatusMitigating:
		return target == StatusResolved || target == StatusInvestigating
	case StatusResolved:
		return target == StatusClosed || target == StatusInvestigating
	case StatusClosed:
		return false // Terminal state
	}

	return false
}

type Incident struct {
	ID          uuid.UUID  `json:"id"`
	IncidentKey string     `json:"incident_key"`
	ServiceID   *uuid.UUID `json:"service_id,omitempty"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Severity    Severity   `json:"severity"`
	Status      Status     `json:"status"`
	AssigneeID  *uuid.UUID `json:"assignee_id,omitempty"`
	StartedAt   time.Time  `json:"started_at"`
	ResolvedAt  *time.Time `json:"resolved_at,omitempty"`
	CreatedBy   uuid.UUID  `json:"created_by"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type IncidentEvent struct {
	ID         uuid.UUID              `json:"id"`
	IncidentID uuid.UUID              `json:"incident_id"`
	EventType  string                 `json:"event_type"`
	Message    string                 `json:"message"`
	ActorID    *uuid.UUID             `json:"actor_id,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt  time.Time              `json:"created_at"`
}

type IncidentComment struct {
	ID         uuid.UUID `json:"id"`
	IncidentID uuid.UUID `json:"incident_id"`
	AuthorID   uuid.UUID `json:"author_id"`
	Content    string    `json:"content"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
