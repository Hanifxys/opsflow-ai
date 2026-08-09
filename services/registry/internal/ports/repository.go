package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/opsflow/registry-service/internal/domain"
)

type ServiceFilter struct {
	OwnerTeam   string
	Criticality string
	Status      string
	Limit       int
	Offset      int
}

type ServiceRepository interface {
	CreateService(ctx context.Context, service *domain.Service) error
	GetServiceByID(ctx context.Context, id uuid.UUID) (*domain.Service, error)
	GetServiceByName(ctx context.Context, name string) (*domain.Service, error)
	ListServices(ctx context.Context, filter ServiceFilter) ([]*domain.Service, int, error)
	UpdateService(ctx context.Context, service *domain.Service) error
	DeleteService(ctx context.Context, id uuid.UUID) error

	AddEnvironment(ctx context.Context, env *domain.ServiceEnvironment) error
	ListEnvironments(ctx context.Context, serviceID uuid.UUID) ([]domain.ServiceEnvironment, error)
	GetEnvironmentByID(ctx context.Context, id uuid.UUID) (*domain.ServiceEnvironment, error)

	AddDependency(ctx context.Context, dep *domain.ServiceDependency) error
	ListDependencies(ctx context.Context, serviceID uuid.UUID) ([]domain.ServiceDependency, error)
	RemoveDependency(ctx context.Context, id uuid.UUID) error

	AddHealthCheck(ctx context.Context, hc *domain.HealthCheck) error
	ListHealthChecks(ctx context.Context, envID uuid.UUID) ([]domain.HealthCheck, error)
}
