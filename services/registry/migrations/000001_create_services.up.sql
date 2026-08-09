-- 000001_create_services.up.sql

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS services (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    owner_team VARCHAR(100) NOT NULL,
    criticality VARCHAR(50) NOT NULL DEFAULT 'MEDIUM',
    repository_url TEXT,
    status VARCHAR(50) NOT NULL DEFAULT 'ACTIVE',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS service_environments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    service_id UUID NOT NULL REFERENCES services(id) ON DELETE CASCADE,
    environment VARCHAR(50) NOT NULL,
    base_url TEXT,
    health_endpoint TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_service_environment UNIQUE(service_id, environment)
);

CREATE TABLE IF NOT EXISTS service_dependencies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    service_id UUID NOT NULL REFERENCES services(id) ON DELETE CASCADE,
    depends_on_service_id UUID NOT NULL REFERENCES services(id) ON DELETE CASCADE,
    dependency_type VARCHAR(50) NOT NULL DEFAULT 'HTTP',
    critical BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_service_dependency UNIQUE(service_id, depends_on_service_id)
);

CREATE TABLE IF NOT EXISTS health_checks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    service_environment_id UUID NOT NULL REFERENCES service_environments(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    method VARCHAR(10) NOT NULL DEFAULT 'GET',
    path TEXT NOT NULL DEFAULT '/health',
    expected_status INTEGER NOT NULL DEFAULT 200,
    timeout_ms INTEGER NOT NULL DEFAULT 5000,
    interval_seconds INTEGER NOT NULL DEFAULT 30,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_services_name ON services(name);
CREATE INDEX IF NOT EXISTS idx_services_owner_team ON services(owner_team);
CREATE INDEX IF NOT EXISTS idx_service_envs_service ON service_environments(service_id);
CREATE INDEX IF NOT EXISTS idx_service_deps_service ON service_dependencies(service_id);
CREATE INDEX IF NOT EXISTS idx_health_checks_env ON health_checks(service_environment_id);
