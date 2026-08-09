package application

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/opsflow/incident-service/internal/domain"
	"github.com/opsflow/incident-service/internal/ports"
)

type IncidentService struct {
	repo ports.IncidentRepository
}

func NewIncidentService(repo ports.IncidentRepository) *IncidentService {
	return &IncidentService{repo: repo}
}

func (s *IncidentService) CreateIncident(ctx context.Context, title, description string, severity domain.Severity, serviceID *uuid.UUID, createdBy uuid.UUID) (*domain.Incident, error) {
	if !severity.IsValid() {
		severity = domain.SeveritySEV3
	}

	incID := uuid.New()
	key := fmt.Sprintf("INC-%s", incID.String()[:8])
	now := time.Now().UTC()

	inc := &domain.Incident{
		ID:          incID,
		IncidentKey: key,
		ServiceID:   serviceID,
		Title:       title,
		Description: description,
		Severity:    severity,
		Status:      domain.StatusOpen,
		StartedAt:   now,
		CreatedBy:   createdBy,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.repo.Create(ctx, inc); err != nil {
		return nil, err
	}

	// Create initial event
	_ = s.repo.AddEvent(ctx, &domain.IncidentEvent{
		ID:         uuid.New(),
		IncidentID: inc.ID,
		EventType:  "INCIDENT_CREATED",
		Message:    fmt.Sprintf("Incident created: %s", title),
		ActorID:    &createdBy,
		CreatedAt:  now,
	})

	return inc, nil
}

func (s *IncidentService) GetIncident(ctx context.Context, id uuid.UUID) (*domain.Incident, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *IncidentService) ListIncidents(ctx context.Context, filter ports.ListFilter) ([]*domain.Incident, int, error) {
	return s.repo.List(ctx, filter)
}

func (s *IncidentService) UpdateStatus(ctx context.Context, id uuid.UUID, targetStatus domain.Status, actorID uuid.UUID, message string) (*domain.Incident, error) {
	if !targetStatus.IsValid() {
		return nil, domain.ErrInvalidStatus
	}

	inc, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if !inc.Status.CanTransitionTo(targetStatus) {
		return nil, domain.ErrInvalidTransition
	}

	oldStatus := inc.Status
	inc.Status = targetStatus
	now := time.Now().UTC()
	inc.UpdatedAt = now

	if targetStatus == domain.StatusResolved && inc.ResolvedAt == nil {
		inc.ResolvedAt = &now
	}

	if err := s.repo.Update(ctx, inc); err != nil {
		return nil, err
	}

	if message == "" {
		message = fmt.Sprintf("Status changed from %s to %s", oldStatus, targetStatus)
	}

	_ = s.repo.AddEvent(ctx, &domain.IncidentEvent{
		ID:         uuid.New(),
		IncidentID: inc.ID,
		EventType:  "STATUS_CHANGED",
		Message:    message,
		ActorID:    &actorID,
		Metadata:   map[string]interface{}{"old_status": oldStatus, "new_status": targetStatus},
		CreatedAt:  now,
	})

	return inc, nil
}

func (s *IncidentService) ResolveIncident(ctx context.Context, id uuid.UUID, actorID uuid.UUID, resolutionNotes string) (*domain.Incident, error) {
	inc, err := s.UpdateStatus(ctx, id, domain.StatusResolved, actorID, fmt.Sprintf("Incident resolved: %s", resolutionNotes))
	if err != nil {
		return nil, err
	}

	if resolutionNotes != "" {
		_ = s.repo.AddComment(ctx, &domain.IncidentComment{
			ID:         uuid.New(),
			IncidentID: id,
			AuthorID:   actorID,
			Content:    fmt.Sprintf("Resolution notes: %s", resolutionNotes),
			CreatedAt:  time.Now().UTC(),
			UpdatedAt:  time.Now().UTC(),
		})
	}

	return inc, nil
}

func (s *IncidentService) AddEvent(ctx context.Context, incidentID uuid.UUID, eventType, message string, actorID *uuid.UUID, metadata map[string]interface{}) (*domain.IncidentEvent, error) {
	_, err := s.repo.GetByID(ctx, incidentID)
	if err != nil {
		return nil, err
	}

	ev := &domain.IncidentEvent{
		ID:         uuid.New(),
		IncidentID: incidentID,
		EventType:  eventType,
		Message:    message,
		ActorID:    actorID,
		Metadata:   metadata,
		CreatedAt:  time.Now().UTC(),
	}

	if err := s.repo.AddEvent(ctx, ev); err != nil {
		return nil, err
	}

	return ev, nil
}

func (s *IncidentService) ListEvents(ctx context.Context, incidentID uuid.UUID) ([]*domain.IncidentEvent, error) {
	return s.repo.ListEvents(ctx, incidentID)
}

func (s *IncidentService) AddComment(ctx context.Context, incidentID uuid.UUID, authorID uuid.UUID, content string) (*domain.IncidentComment, error) {
	_, err := s.repo.GetByID(ctx, incidentID)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	c := &domain.IncidentComment{
		ID:         uuid.New(),
		IncidentID: incidentID,
		AuthorID:   authorID,
		Content:    content,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	if err := s.repo.AddComment(ctx, c); err != nil {
		return nil, err
	}

	return c, nil
}

func (s *IncidentService) ListComments(ctx context.Context, incidentID uuid.UUID) ([]*domain.IncidentComment, error) {
	return s.repo.ListComments(ctx, incidentID)
}
