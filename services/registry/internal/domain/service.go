package domain

import (
	"time"

	"github.com/google/uuid"
)

type ServiceStatus string

const (
	ServiceStatusActive     ServiceStatus = "ACTIVE"
	ServiceStatusDeprecated ServiceStatus = "DEPRECATED"
	ServiceStatusArchived   ServiceStatus = "ARCHIVED"
)

type Criticality string

const (
	CriticalityCritical Criticality = "CRITICAL"
	CriticalityHigh     Criticality = "HIGH"
	CriticalityMedium   Criticality = "MEDIUM"
	CriticalityLow      Criticality = "LOW"
)

type Service struct {
	ID            uuid.UUID             `json:"id"`
	Name          string                `json:"name"`
	Description   string                `json:"description"`
	OwnerTeam     string                `json:"owner_team"`
	Criticality   Criticality           `json:"criticality"`
	RepositoryURL string                `json:"repository_url"`
	Status        ServiceStatus         `json:"status"`
	CreatedAt     time.Time             `json:"created_at"`
	UpdatedAt     time.Time             `json:"updated_at"`
	Environments  []ServiceEnvironment  `json:"environments,omitempty"`
	Dependencies  []ServiceDependency   `json:"dependencies,omitempty"`
}

type ServiceEnvironment struct {
	ID             uuid.UUID     `json:"id"`
	ServiceID      uuid.UUID     `json:"service_id"`
	Environment    string        `json:"environment"` // production, staging, dev
	BaseURL        string        `json:"base_url"`
	HealthEndpoint string        `json:"health_endpoint"`
	CreatedAt      time.Time     `json:"created_at"`
	UpdatedAt      time.Time     `json:"updated_at"`
	HealthChecks   []HealthCheck `json:"health_checks,omitempty"`
}

type ServiceDependency struct {
	ID                 uuid.UUID `json:"id"`
	ServiceID          uuid.UUID `json:"service_id"`
	DependsOnServiceID uuid.UUID `json:"depends_on_service_id"`
	DependsOnName      string    `json:"depends_on_name,omitempty"`
	DependencyType     string    `json:"dependency_type"` // HTTP, gRPC, MQ, DB
	Critical           bool      `json:"critical"`
	CreatedAt          time.Time `json:"created_at"`
}

type HealthCheck struct {
	ID                   uuid.UUID `json:"id"`
	ServiceEnvironmentID uuid.UUID `json:"service_environment_id"`
	Name                 string    `json:"name"`
	Method               string    `json:"method"`
	Path                 string    `json:"path"`
	ExpectedStatus       int       `json:"expected_status"`
	TimeoutMS            int       `json:"timeout_ms"`
	IntervalSeconds      int       `json:"interval_seconds"`
	Enabled              bool      `json:"enabled"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}
