package application

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/opsflow/registry-service/internal/domain"
	"github.com/opsflow/registry-service/internal/ports"
	"time"
)

type ServiceRegistryService struct {
	repo ports.ServiceRepository
}

func NewServiceRegistryService(repo ports.ServiceRepository) *ServiceRegistryService {
	return &ServiceRegistryService{repo: repo}
}

func (s *ServiceRegistryService) CreateService(ctx context.Context, name, description, ownerTeam string, criticality domain.Criticality, repoURL string) (*domain.Service, error) {
	existing, err := s.repo.GetServiceByName(ctx, name)
	if err == nil && existing != nil {
		return nil, domain.ErrServiceNameExists
	} else if err != nil && !errors.Is(err, domain.ErrServiceNotFound) {
		return nil, err
	}

	if criticality == "" {
		criticality = domain.CriticalityMedium
	}

	now := time.Now().UTC()
	svc := &domain.Service{
		ID:            uuid.New(),
		Name:          name,
		Description:   description,
		OwnerTeam:     ownerTeam,
		Criticality:   criticality,
		RepositoryURL: repoURL,
		Status:        domain.ServiceStatusActive,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := s.repo.CreateService(ctx, svc); err != nil {
		return nil, err
	}

	return svc, nil
}

func (s *ServiceRegistryService) GetService(ctx context.Context, id uuid.UUID) (*domain.Service, error) {
	return s.repo.GetServiceByID(ctx, id)
}

func (s *ServiceRegistryService) ListServices(ctx context.Context, filter ports.ServiceFilter) ([]*domain.Service, int, error) {
	return s.repo.ListServices(ctx, filter)
}

func (s *ServiceRegistryService) UpdateService(ctx context.Context, svc *domain.Service) error {
	svc.UpdatedAt = time.Now().UTC()
	return s.repo.UpdateService(ctx, svc)
}

func (s *ServiceRegistryService) DeleteService(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeleteService(ctx, id)
}

func (s *ServiceRegistryService) AddEnvironment(ctx context.Context, serviceID uuid.UUID, envName, baseURL, healthEndpoint string) (*domain.ServiceEnvironment, error) {
	_, err := s.repo.GetServiceByID(ctx, serviceID)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	env := &domain.ServiceEnvironment{
		ID:             uuid.New(),
		ServiceID:      serviceID,
		Environment:    envName,
		BaseURL:        baseURL,
		HealthEndpoint: healthEndpoint,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := s.repo.AddEnvironment(ctx, env); err != nil {
		return nil, err
	}

	return env, nil
}

func (s *ServiceRegistryService) ListEnvironments(ctx context.Context, serviceID uuid.UUID) ([]domain.ServiceEnvironment, error) {
	return s.repo.ListEnvironments(ctx, serviceID)
}

func (s *ServiceRegistryService) AddDependency(ctx context.Context, serviceID, dependsOnID uuid.UUID, depType string, critical bool) (*domain.ServiceDependency, error) {
	if serviceID == dependsOnID {
		return nil, domain.ErrSelfDependency
	}

	_, err := s.repo.GetServiceByID(ctx, serviceID)
	if err != nil {
		return nil, err
	}

	targetSvc, err := s.repo.GetServiceByID(ctx, dependsOnID)
	if err != nil {
		return nil, err
	}

	if depType == "" {
		depType = "HTTP"
	}

	dep := &domain.ServiceDependency{
		ID:                 uuid.New(),
		ServiceID:          serviceID,
		DependsOnServiceID: dependsOnID,
		DependsOnName:      targetSvc.Name,
		DependencyType:     depType,
		Critical:           critical,
		CreatedAt:          time.Now().UTC(),
	}

	if err := s.repo.AddDependency(ctx, dep); err != nil {
		return nil, err
	}

	return dep, nil
}

func (s *ServiceRegistryService) ListDependencies(ctx context.Context, serviceID uuid.UUID) ([]domain.ServiceDependency, error) {
	return s.repo.ListDependencies(ctx, serviceID)
}

func (s *ServiceRegistryService) RemoveDependency(ctx context.Context, id uuid.UUID) error {
	return s.repo.RemoveDependency(ctx, id)
}

func (s *ServiceRegistryService) AddHealthCheck(ctx context.Context, envID uuid.UUID, name, method, path string, expectedStatus, timeoutMS, intervalSec int) (*domain.HealthCheck, error) {
	_, err := s.repo.GetEnvironmentByID(ctx, envID)
	if err != nil {
		return nil, err
	}

	if method == "" {
		method = "GET"
	}
	if path == "" {
		path = "/health"
	}
	if expectedStatus <= 0 {
		expectedStatus = 200
	}
	if timeoutMS <= 0 {
		timeoutMS = 5000
	}
	if intervalSec <= 0 {
		intervalSec = 30
	}

	now := time.Now().UTC()
	hc := &domain.HealthCheck{
		ID:                   uuid.New(),
		ServiceEnvironmentID: envID,
		Name:                 name,
		Method:               method,
		Path:                 path,
		ExpectedStatus:       expectedStatus,
		TimeoutMS:            timeoutMS,
		IntervalSeconds:      intervalSec,
		Enabled:              true,
		CreatedAt:            now,
		UpdatedAt:            now,
	}

	if err := s.repo.AddHealthCheck(ctx, hc); err != nil {
		return nil, err
	}

	return hc, nil
}

func (s *ServiceRegistryService) ListHealthChecks(ctx context.Context, envID uuid.UUID) ([]domain.HealthCheck, error) {
	return s.repo.ListHealthChecks(ctx, envID)
}
