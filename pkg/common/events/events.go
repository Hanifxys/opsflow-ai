package events

import (
	"time"

	"github.com/google/uuid"
)

const (
	ExchangeEvents = "opsflow.events"
	ExchangeDLX    = "opsflow.dlx"

	RoutingIncidentCreated      = "incident.created"
	RoutingIncidentResolved     = "incident.resolved"
	RoutingDeploymentCreated    = "deployment.created"
	RoutingDeploymentValidated  = "deployment.validated"
	RoutingServiceHealthChanged = "service.health_changed"
	RoutingNotificationReq      = "notification.requested"
)

type Header struct {
	EventID       uuid.UUID `json:"event_id"`
	AggregateType string    `json:"aggregate_type"`
	AggregateID   uuid.UUID `json:"aggregate_id"`
	EventType     string    `json:"event_type"`
	OccurredAt    time.Time `json:"occurred_at"`
	CorrelationID string    `json:"correlation_id,omitempty"`
}

type Event[T any] struct {
	Header  Header `json:"header"`
	Payload T      `json:"payload"`
}

type IncidentCreatedPayload struct {
	IncidentID  uuid.UUID `json:"incident_id"`
	IncidentKey string    `json:"incident_key"`
	ServiceID   *uuid.UUID`json:"service_id,omitempty"`
	Title       string    `json:"title"`
	Severity    string    `json:"severity"`
	CreatedBy   uuid.UUID `json:"created_by"`
}

type IncidentResolvedPayload struct {
	IncidentID      uuid.UUID `json:"incident_id"`
	IncidentKey     string    `json:"incident_key"`
	ResolvedBy      uuid.UUID `json:"resolved_by"`
	ResolutionNotes string    `json:"resolution_notes,omitempty"`
	ResolvedAt      time.Time `json:"resolved_at"`
}

type ServiceHealthChangedPayload struct {
	ServiceID     uuid.UUID `json:"service_id"`
	ServiceName   string    `json:"service_name"`
	Environment   string    `json:"environment"`
	PreviousState string    `json:"previous_state"`
	NewState      string    `json:"new_state"`
}
