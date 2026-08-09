package application_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/opsflow/incident-service/internal/application"
	"github.com/opsflow/incident-service/internal/domain"
	"github.com/opsflow/incident-service/internal/ports"
)

type mockIncidentRepository struct {
	incidents map[uuid.UUID]*domain.Incident
	events    map[uuid.UUID][]*domain.IncidentEvent
	comments  map[uuid.UUID][]*domain.IncidentComment
}

func newMockIncidentRepository() *mockIncidentRepository {
	return &mockIncidentRepository{
		incidents: make(map[uuid.UUID]*domain.Incident),
		events:    make(map[uuid.UUID][]*domain.IncidentEvent),
		comments:  make(map[uuid.UUID][]*domain.IncidentComment),
	}
}

func (m *mockIncidentRepository) Create(ctx context.Context, inc *domain.Incident) error {
	m.incidents[inc.ID] = inc
	return nil
}

func (m *mockIncidentRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Incident, error) {
	inc, ok := m.incidents[id]
	if !ok {
		return nil, domain.ErrIncidentNotFound
	}
	return inc, nil
}

func (m *mockIncidentRepository) GetByKey(ctx context.Context, key string) (*domain.Incident, error) {
	for _, inc := range m.incidents {
		if inc.IncidentKey == key {
			return inc, nil
		}
	}
	return nil, domain.ErrIncidentNotFound
}

func (m *mockIncidentRepository) List(ctx context.Context, filter ports.ListFilter) ([]*domain.Incident, int, error) {
	var list []*domain.Incident
	for _, inc := range m.incidents {
		if filter.Status != nil && inc.Status != *filter.Status {
			continue
		}
		if filter.Severity != nil && inc.Severity != *filter.Severity {
			continue
		}
		list = append(list, inc)
	}
	return list, len(list), nil
}

func (m *mockIncidentRepository) Update(ctx context.Context, inc *domain.Incident) error {
	if _, ok := m.incidents[inc.ID]; !ok {
		return domain.ErrIncidentNotFound
	}
	m.incidents[inc.ID] = inc
	return nil
}

func (m *mockIncidentRepository) AddEvent(ctx context.Context, ev *domain.IncidentEvent) error {
	m.events[ev.IncidentID] = append(m.events[ev.IncidentID], ev)
	return nil
}

func (m *mockIncidentRepository) ListEvents(ctx context.Context, incidentID uuid.UUID) ([]*domain.IncidentEvent, error) {
	return m.events[incidentID], nil
}

func (m *mockIncidentRepository) AddComment(ctx context.Context, c *domain.IncidentComment) error {
	m.comments[c.IncidentID] = append(m.comments[c.IncidentID], c)
	return nil
}

func (m *mockIncidentRepository) ListComments(ctx context.Context, incidentID uuid.UUID) ([]*domain.IncidentComment, error) {
	return m.comments[incidentID], nil
}

func TestIncidentService_LifecycleAndStateMachine(t *testing.T) {
	repo := newMockIncidentRepository()
	svc := application.NewIncidentService(repo)
	ctx := context.Background()
	user1 := uuid.New()

	// 1. Create Incident
	inc, err := svc.CreateIncident(ctx, "High Latency on Payment Gateway", "p99 > 2s", domain.SeveritySEV1, nil, user1)
	if err != nil {
		t.Fatalf("failed to create incident: %v", err)
	}

	if inc.Status != domain.StatusOpen {
		t.Errorf("expected initial status OPEN, got %s", inc.Status)
	}
	if inc.Severity != domain.SeveritySEV1 {
		t.Errorf("expected SEV1, got %s", inc.Severity)
	}

	// 2. Transition OPEN -> INVESTIGATING (Allowed)
	inc, err = svc.UpdateStatus(ctx, inc.ID, domain.StatusInvestigating, user1, "Investigating payment DB pool")
	if err != nil {
		t.Fatalf("failed to transition to INVESTIGATING: %v", err)
	}
	if inc.Status != domain.StatusInvestigating {
		t.Errorf("expected status INVESTIGATING, got %s", inc.Status)
	}

	// 3. Transition INVESTIGATING -> CLOSED (Invalid Transition according to state machine)
	_, err = svc.UpdateStatus(ctx, inc.ID, domain.StatusClosed, user1, "Closing directly")
	if err != domain.ErrInvalidTransition {
		t.Errorf("expected ErrInvalidTransition for INVESTIGATING -> CLOSED, got %v", err)
	}

	// 4. Transition INVESTIGATING -> MITIGATING (Allowed)
	inc, err = svc.UpdateStatus(ctx, inc.ID, domain.StatusMitigating, user1, "Scaled DB replicas")
	if err != nil {
		t.Fatalf("failed to transition to MITIGATING: %v", err)
	}

	// 5. Resolve Incident MITIGATING -> RESOLVED
	inc, err = svc.ResolveIncident(ctx, inc.ID, user1, "Added 3 replicas, latency normal")
	if err != nil {
		t.Fatalf("failed to resolve incident: %v", err)
	}
	if inc.Status != domain.StatusResolved {
		t.Errorf("expected status RESOLVED, got %s", inc.Status)
	}
	if inc.ResolvedAt == nil {
		t.Error("expected ResolvedAt timestamp to be set")
	}

	// 6. Transition RESOLVED -> CLOSED (Allowed)
	inc, err = svc.UpdateStatus(ctx, inc.ID, domain.StatusClosed, user1, "RCA completed")
	if err != nil {
		t.Fatalf("failed to transition to CLOSED: %v", err)
	}

	// 7. Verify Timeline Events
	events, err := svc.ListEvents(ctx, inc.ID)
	if err != nil {
		t.Fatalf("failed to list events: %v", err)
	}
	if len(events) < 4 {
		t.Errorf("expected at least 4 events, got %d", len(events))
	}

	// 8. Verify Comments
	comments, err := svc.ListComments(ctx, inc.ID)
	if err != nil {
		t.Fatalf("failed to list comments: %v", err)
	}
	if len(comments) != 1 {
		t.Errorf("expected 1 resolution comment, got %d", len(comments))
	}
}
