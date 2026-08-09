package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/opsflow/registry-service/internal/domain"
	"github.com/opsflow/registry-service/internal/ports"
)

type PostgresServiceRepository struct {
	pool *pgxpool.Pool
}

func NewServiceRepository(pool *pgxpool.Pool) *PostgresServiceRepository {
	return &PostgresServiceRepository{pool: pool}
}

func (r *PostgresServiceRepository) CreateService(ctx context.Context, s *domain.Service) error {
	query := `
		INSERT INTO services (id, name, description, owner_team, criticality, repository_url, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := r.pool.Exec(ctx, query, s.ID, s.Name, s.Description, s.OwnerTeam, s.Criticality, s.RepositoryURL, s.Status, s.CreatedAt, s.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to insert service: %w", err)
	}
	return nil
}

func (r *PostgresServiceRepository) GetServiceByID(ctx context.Context, id uuid.UUID) (*domain.Service, error) {
	query := `
		SELECT id, name, description, owner_team, criticality, repository_url, status, created_at, updated_at
		FROM services
		WHERE id = $1
	`
	s := &domain.Service{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&s.ID, &s.Name, &s.Description, &s.OwnerTeam, &s.Criticality, &s.RepositoryURL, &s.Status, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrServiceNotFound
		}
		return nil, fmt.Errorf("failed to get service by id: %w", err)
	}

	envs, _ := r.ListEnvironments(ctx, s.ID)
	s.Environments = envs

	deps, _ := r.ListDependencies(ctx, s.ID)
	s.Dependencies = deps

	return s, nil
}

func (r *PostgresServiceRepository) GetServiceByName(ctx context.Context, name string) (*domain.Service, error) {
	query := `
		SELECT id, name, description, owner_team, criticality, repository_url, status, created_at, updated_at
		FROM services
		WHERE name = $1
	`
	s := &domain.Service{}
	err := r.pool.QueryRow(ctx, query, name).Scan(
		&s.ID, &s.Name, &s.Description, &s.OwnerTeam, &s.Criticality, &s.RepositoryURL, &s.Status, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrServiceNotFound
		}
		return nil, fmt.Errorf("failed to get service by name: %w", err)
	}
	return s, nil
}

func (r *PostgresServiceRepository) ListServices(ctx context.Context, filter ports.ServiceFilter) ([]*domain.Service, int, error) {
	limit := filter.Limit
	if limit <= 0 {
		limit = 20
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	whereClause := "WHERE 1=1"
	args := []interface{}{}
	argIdx := 1

	if filter.OwnerTeam != "" {
		whereClause += fmt.Sprintf(" AND owner_team = $%d", argIdx)
		args = append(args, filter.OwnerTeam)
		argIdx++
	}
	if filter.Criticality != "" {
		whereClause += fmt.Sprintf(" AND criticality = $%d", argIdx)
		args = append(args, filter.Criticality)
		argIdx++
	}
	if filter.Status != "" {
		whereClause += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, filter.Status)
		argIdx++
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM services %s", whereClause)
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count services: %w", err)
	}

	query := fmt.Sprintf(`
		SELECT id, name, description, owner_team, criticality, repository_url, status, created_at, updated_at
		FROM services
		%s
		ORDER BY name ASC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list services: %w", err)
	}
	defer rows.Close()

	var res []*domain.Service
	for rows.Next() {
		s := &domain.Service{}
		err := rows.Scan(
			&s.ID, &s.Name, &s.Description, &s.OwnerTeam, &s.Criticality, &s.RepositoryURL, &s.Status, &s.CreatedAt, &s.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		res = append(res, s)
	}

	return res, total, rows.Err()
}

func (r *PostgresServiceRepository) UpdateService(ctx context.Context, s *domain.Service) error {
	query := `
		UPDATE services
		SET name = $1, description = $2, owner_team = $3, criticality = $4, repository_url = $5, status = $6, updated_at = $7
		WHERE id = $8
	`
	_, err := r.pool.Exec(ctx, query, s.Name, s.Description, s.OwnerTeam, s.Criticality, s.RepositoryURL, s.Status, time.Now().UTC(), s.ID)
	if err != nil {
		return fmt.Errorf("failed to update service: %w", err)
	}
	return nil
}

func (r *PostgresServiceRepository) DeleteService(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM services WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}

func (r *PostgresServiceRepository) AddEnvironment(ctx context.Context, env *domain.ServiceEnvironment) error {
	query := `
		INSERT INTO service_environments (id, service_id, environment, base_url, health_endpoint, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.pool.Exec(ctx, query, env.ID, env.ServiceID, env.Environment, env.BaseURL, env.HealthEndpoint, env.CreatedAt, env.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to insert service environment: %w", err)
	}
	return nil
}

func (r *PostgresServiceRepository) ListEnvironments(ctx context.Context, serviceID uuid.UUID) ([]domain.ServiceEnvironment, error) {
	query := `
		SELECT id, service_id, environment, base_url, health_endpoint, created_at, updated_at
		FROM service_environments
		WHERE service_id = $1
		ORDER BY environment ASC
	`
	rows, err := r.pool.Query(ctx, query, serviceID)
	if err != nil {
		return nil, fmt.Errorf("failed to list service environments: %w", err)
	}
	defer rows.Close()

	var envs []domain.ServiceEnvironment
	for rows.Next() {
		var env domain.ServiceEnvironment
		err := rows.Scan(&env.ID, &env.ServiceID, &env.Environment, &env.BaseURL, &env.HealthEndpoint, &env.CreatedAt, &env.UpdatedAt)
		if err != nil {
			return nil, err
		}
		hcs, _ := r.ListHealthChecks(ctx, env.ID)
		env.HealthChecks = hcs
		envs = append(envs, env)
	}
	return envs, rows.Err()
}

func (r *PostgresServiceRepository) GetEnvironmentByID(ctx context.Context, id uuid.UUID) (*domain.ServiceEnvironment, error) {
	query := `
		SELECT id, service_id, environment, base_url, health_endpoint, created_at, updated_at
		FROM service_environments
		WHERE id = $1
	`
	env := &domain.ServiceEnvironment{}
	err := r.pool.QueryRow(ctx, query, id).Scan(&env.ID, &env.ServiceID, &env.Environment, &env.BaseURL, &env.HealthEndpoint, &env.CreatedAt, &env.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrEnvironmentNotFound
		}
		return nil, err
	}
	return env, nil
}

func (r *PostgresServiceRepository) AddDependency(ctx context.Context, dep *domain.ServiceDependency) error {
	query := `
		INSERT INTO service_dependencies (id, service_id, depends_on_service_id, dependency_type, critical, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.pool.Exec(ctx, query, dep.ID, dep.ServiceID, dep.DependsOnServiceID, dep.DependencyType, dep.Critical, dep.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to insert service dependency: %w", err)
	}
	return nil
}

func (r *PostgresServiceRepository) ListDependencies(ctx context.Context, serviceID uuid.UUID) ([]domain.ServiceDependency, error) {
	query := `
		SELECT sd.id, sd.service_id, sd.depends_on_service_id, s.name, sd.dependency_type, sd.critical, sd.created_at
		FROM service_dependencies sd
		JOIN services s ON sd.depends_on_service_id = s.id
		WHERE sd.service_id = $1
	`
	rows, err := r.pool.Query(ctx, query, serviceID)
	if err != nil {
		return nil, fmt.Errorf("failed to list service dependencies: %w", err)
	}
	defer rows.Close()

	var deps []domain.ServiceDependency
	for rows.Next() {
		var dep domain.ServiceDependency
		err := rows.Scan(&dep.ID, &dep.ServiceID, &dep.DependsOnServiceID, &dep.DependsOnName, &dep.DependencyType, &dep.Critical, &dep.CreatedAt)
		if err != nil {
			return nil, err
		}
		deps = append(deps, dep)
	}
	return deps, rows.Err()
}

func (r *PostgresServiceRepository) RemoveDependency(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM service_dependencies WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}

func (r *PostgresServiceRepository) AddHealthCheck(ctx context.Context, hc *domain.HealthCheck) error {
	query := `
		INSERT INTO health_checks (id, service_environment_id, name, method, path, expected_status, timeout_ms, interval_seconds, enabled, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	_, err := r.pool.Exec(ctx, query, hc.ID, hc.ServiceEnvironmentID, hc.Name, hc.Method, hc.Path, hc.ExpectedStatus, hc.TimeoutMS, hc.IntervalSeconds, hc.Enabled, hc.CreatedAt, hc.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to insert health check: %w", err)
	}
	return nil
}

func (r *PostgresServiceRepository) ListHealthChecks(ctx context.Context, envID uuid.UUID) ([]domain.HealthCheck, error) {
	query := `
		SELECT id, service_environment_id, name, method, path, expected_status, timeout_ms, interval_seconds, enabled, created_at, updated_at
		FROM health_checks
		WHERE service_environment_id = $1
	`
	rows, err := r.pool.Query(ctx, query, envID)
	if err != nil {
		return nil, fmt.Errorf("failed to list health checks: %w", err)
	}
	defer rows.Close()

	var hcs []domain.HealthCheck
	for rows.Next() {
		var hc domain.HealthCheck
		err := rows.Scan(&hc.ID, &hc.ServiceEnvironmentID, &hc.Name, &hc.Method, &hc.Path, &hc.ExpectedStatus, &hc.TimeoutMS, &hc.IntervalSeconds, &hc.Enabled, &hc.CreatedAt, &hc.UpdatedAt)
		if err != nil {
			return nil, err
		}
		hcs = append(hcs, hc)
	}
	return hcs, rows.Err()
}
