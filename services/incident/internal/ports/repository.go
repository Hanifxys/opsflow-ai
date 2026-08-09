package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/opsflow/incident-service/internal/domain"
)

type ListFilter struct {
	Status   *domain.Status
	Severity *domain.Severity
	Limit    int
	Offset   int
}

type IncidentRepository interface {
	Create(ctx context.Context, incident *domain.Incident) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Incident, error)
	GetByKey(ctx context.Context, key string) (*domain.Incident, error)
	List(ctx context.Context, filter ListFilter) ([]*domain.Incident, int, error)
	Update(ctx context.Context, incident *domain.Incident) error

	AddEvent(ctx context.Context, event *domain.IncidentEvent) error
	ListEvents(ctx context.Context, incidentID uuid.UUID) ([]*domain.IncidentEvent, error)

	AddComment(ctx context.Context, comment *domain.IncidentComment) error
	ListComments(ctx context.Context, incidentID uuid.UUID) ([]*domain.IncidentComment, error)
}
