package application_test

import (
	"context"
	"testing"

	"github.com/opsflow/notification-service/internal/application"
	"github.com/opsflow/notification-service/internal/domain"
)

type mockNotificationRepository struct {
	processed map[string]*domain.Notification
}

func newMockNotificationRepository() *mockNotificationRepository {
	return &mockNotificationRepository{
		processed: make(map[string]*domain.Notification),
	}
}

func (m *mockNotificationRepository) SaveIfNew(ctx context.Context, n *domain.Notification) (bool, error) {
	if _, exists := m.processed[n.IdempotencyKey]; exists {
		return false, nil // Duplicate detected
	}
	m.processed[n.IdempotencyKey] = n
	return true, nil // Newly processed
}

func (m *mockNotificationRepository) ListRecent(ctx context.Context, limit int) ([]*domain.Notification, error) {
	var res []*domain.Notification
	for _, n := range m.processed {
		res = append(res, n)
	}
	return res, nil
}

func TestNotificationService_Idempotency(t *testing.T) {
	repo := newMockNotificationRepository()
	svc := application.NewNotificationService(repo)
	ctx := context.Background()

	payload := []byte(`{"incident_id":"123","title":"Server Down","created_by":"c0eebc99-9c0b-4ef8-bb6d-6bb9bd380c11"}`)

	// 1. Process First Time (New)
	isNew, err := svc.ProcessEvent(ctx, "EMAIL", "incident.created", payload)
	if err != nil {
		t.Fatalf("failed to process event: %v", err)
	}
	if !isNew {
		t.Error("expected first delivery to be marked as new")
	}

	// 2. Process Duplicate Event (Same payload and event_type)
	isNew, err = svc.ProcessEvent(ctx, "EMAIL", "incident.created", payload)
	if err != nil {
		t.Fatalf("failed to process duplicate event: %v", err)
	}
	if isNew {
		t.Error("expected duplicate delivery to be ignored (idempotent)")
	}

	// 3. Verify Only 1 Notification Record Saved
	recent, err := svc.ListRecentNotifications(ctx, 10)
	if err != nil {
		t.Fatalf("failed to list notifications: %v", err)
	}
	if len(recent) != 1 {
		t.Errorf("expected 1 notification in repository, got %d", len(recent))
	}
	if recent[0].UserID == nil || recent[0].UserID.String() != "c0eebc99-9c0b-4ef8-bb6d-6bb9bd380c11" {
		t.Errorf("expected user_id to be extracted from payload, got %v", recent[0].UserID)
	}
}
