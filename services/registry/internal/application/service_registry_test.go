package application_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/opsflow/registry-service/internal/application"
	"github.com/opsflow/registry-service/internal/domain"
	"github.com/opsflow/registry-service/internal/ports"
)

type mockServiceRepository struct {
	services     map[uuid.UUID]*domain.Service
	environments map[uuid.UUID]*domain.ServiceEnvironment
	envByService map[uuid.UUID][]domain.ServiceEnvironment
	dependencies map[uuid.UUID][]domain.ServiceDependency
	healthChecks map[uuid.UUID][]domain.HealthCheck
}

func newMockServiceRepository() *mockServiceRepository {
	return &mockServiceRepository{
		services:     make(map[uuid.UUID]*domain.Service),
		environments: make(map[uuid.UUID]*domain.ServiceEnvironment),
		envByService: make(map[uuid.UUID][]domain.ServiceEnvironment),
		dependencies: make(map[uuid.UUID][]domain.ServiceDependency),
		healthChecks: make(map[uuid.UUID][]domain.HealthCheck),
	}
}

func (m *mockServiceRepository) CreateService(ctx context.Context, s *domain.Service) error {
	m.services[s.ID] = s
	return nil
}

func (m *mockServiceRepository) GetServiceByID(ctx context.Context, id uuid.UUID) (*domain.Service, error) {
	s, ok := m.services[id]
	if !ok {
		return nil, domain.ErrServiceNotFound
	}
	return s, nil
}

func (m *mockServiceRepository) GetServiceByName(ctx context.Context, name string) (*domain.Service, error) {
	for _, s := range m.services {
		if s.Name == name {
			return s, nil
		}
	}
	return nil, domain.ErrServiceNotFound
}

func (m *mockServiceRepository) ListServices(ctx context.Context, filter ports.ServiceFilter) ([]*domain.Service, int, error) {
	var list []*domain.Service
	for _, s := range m.services {
		if filter.OwnerTeam != "" && s.OwnerTeam != filter.OwnerTeam {
			continue
		}
		list = append(list, s)
	}
	return list, len(list), nil
}

func (m *mockServiceRepository) UpdateService(ctx context.Context, s *domain.Service) error {
	m.services[s.ID] = s
	return nil
}

func (m *mockServiceRepository) DeleteService(ctx context.Context, id uuid.UUID) error {
	delete(m.services, id)
	return nil
}

func (m *mockServiceRepository) AddEnvironment(ctx context.Context, env *domain.ServiceEnvironment) error {
	m.environments[env.ID] = env
	m.envByService[env.ServiceID] = append(m.envByService[env.ServiceID], *env)
	return nil
}

func (m *mockServiceRepository) ListEnvironments(ctx context.Context, serviceID uuid.UUID) ([]domain.ServiceEnvironment, error) {
	return m.envByService[serviceID], nil
}

func (m *mockServiceRepository) GetEnvironmentByID(ctx context.Context, id uuid.UUID) (*domain.ServiceEnvironment, error) {
	env, ok := m.environments[id]
	if !ok {
		return nil, domain.ErrEnvironmentNotFound
	}
	return env, nil
}

func (m *mockServiceRepository) AddDependency(ctx context.Context, dep *domain.ServiceDependency) error {
	m.dependencies[dep.ServiceID] = append(m.dependencies[dep.ServiceID], *dep)
	return nil
}

func (m *mockServiceRepository) ListDependencies(ctx context.Context, serviceID uuid.UUID) ([]domain.ServiceDependency, error) {
	return m.dependencies[serviceID], nil
}

func (m *mockServiceRepository) RemoveDependency(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (m *mockServiceRepository) AddHealthCheck(ctx context.Context, hc *domain.HealthCheck) error {
	m.healthChecks[hc.ServiceEnvironmentID] = append(m.healthChecks[hc.ServiceEnvironmentID], *hc)
	return nil
}

func (m *mockServiceRepository) ListHealthChecks(ctx context.Context, envID uuid.UUID) ([]domain.HealthCheck, error) {
	return m.healthChecks[envID], nil
}

func TestServiceRegistry_TopologyAndDependencies(t *testing.T) {
	repo := newMockServiceRepository()
	svc := application.NewServiceRegistryService(repo)
	ctx := context.Background()

	// 1. Create payment-service
	payment, err := svc.CreateService(ctx, "payment-service", "Processes transactions", "Core Payments", domain.CriticalityCritical, "https://github.com/opsflow/payment-service")
	if err != nil {
		t.Fatalf("failed to create payment-service: %v", err)
	}

	// 2. Create auth-service
	auth, err := svc.CreateService(ctx, "auth-service", "Handles user auth & JWT", "Platform Security", domain.CriticalityHigh, "https://github.com/opsflow/auth-service")
	if err != nil {
		t.Fatalf("failed to create auth-service: %v", err)
	}

	// 3. Add Environment to payment-service
	env, err := svc.AddEnvironment(ctx, payment.ID, "production", "https://payment.opsflow.local", "/health")
	if err != nil {
		t.Fatalf("failed to add environment: %v", err)
	}
	if env.Environment != "production" {
		t.Errorf("expected production, got %s", env.Environment)
	}

	// 4. Add Dependency (payment-service -> auth-service)
	dep, err := svc.AddDependency(ctx, payment.ID, auth.ID, "HTTP", true)
	if err != nil {
		t.Fatalf("failed to add dependency: %v", err)
	}
	if dep.DependsOnName != "auth-service" {
		t.Errorf("expected auth-service, got %s", dep.DependsOnName)
	}

	// 5. Self Dependency Error check
	_, err = svc.AddDependency(ctx, payment.ID, payment.ID, "HTTP", true)
	if err != domain.ErrSelfDependency {
		t.Errorf("expected ErrSelfDependency, got %v", err)
	}

	// 6. Add Health Check Definition
	hc, err := svc.AddHealthCheck(ctx, env.ID, "HTTP Health Probe", "GET", "/health", 200, 3000, 15)
	if err != nil {
		t.Fatalf("failed to add health check: %v", err)
	}
	if hc.ExpectedStatus != 200 {
		t.Errorf("expected 200, got %d", hc.ExpectedStatus)
	}
}
