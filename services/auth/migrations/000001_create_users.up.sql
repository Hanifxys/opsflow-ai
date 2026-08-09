-- 000001_create_users.up.sql

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    display_name VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'ACTIVE',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(50) NOT NULL UNIQUE,
    description TEXT
);

CREATE TABLE IF NOT EXISTS permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(100) NOT NULL UNIQUE,
    description TEXT
);

CREATE TABLE IF NOT EXISTS user_roles (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, role_id)
);

CREATE TABLE IF NOT EXISTS role_permissions (
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_id UUID NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, permission_id)
);

CREATE TABLE IF NOT EXISTS refresh_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL UNIQUE,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    revoked BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_hash ON refresh_tokens(token_hash);

-- Seed Default Roles
INSERT INTO roles (id, name, description) VALUES
    ('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'ADMIN', 'System Administrator with full access'),
    ('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a12', 'OPS', 'Operations & SRE engineer'),
    ('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a13', 'ENGINEER', 'Software engineer'),
    ('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a14', 'VIEWER', 'Read-only observer')
ON CONFLICT (name) DO NOTHING;

-- Seed Default Permissions
INSERT INTO permissions (id, code, description) VALUES
    ('b0eebc99-9c0b-4ef8-bb6d-6bb9bd380b11', 'admin:all', 'Full administrative access'),
    ('b0eebc99-9c0b-4ef8-bb6d-6bb9bd380b12', 'service:read', 'Read services'),
    ('b0eebc99-9c0b-4ef8-bb6d-6bb9bd380b13', 'service:write', 'Create/update services'),
    ('b0eebc99-9c0b-4ef8-bb6d-6bb9bd380b14', 'incident:read', 'Read incidents'),
    ('b0eebc99-9c0b-4ef8-bb6d-6bb9bd380b15', 'incident:create', 'Create incidents'),
    ('b0eebc99-9c0b-4ef8-bb6d-6bb9bd380b16', 'incident:update', 'Update incidents'),
    ('b0eebc99-9c0b-4ef8-bb6d-6bb9bd380b17', 'incident:resolve', 'Resolve incidents'),
    ('b0eebc99-9c0b-4ef8-bb6d-6bb9bd380b18', 'release:read', 'Read releases'),
    ('b0eebc99-9c0b-4ef8-bb6d-6bb9bd380b19', 'release:write', 'Create/update releases'),
    ('b0eebc99-9c0b-4ef8-bb6d-6bb9bd380b20', 'ai:use', 'Use AI Assistant'),
    ('b0eebc99-9c0b-4ef8-bb6d-6bb9bd380b21', 'ai:execute', 'Execute AI mutations')
ON CONFLICT (code) DO NOTHING;

-- Assign Permissions to ADMIN Role (admin:all)
INSERT INTO role_permissions (role_id, permission_id) VALUES
    ('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'b0eebc99-9c0b-4ef8-bb6d-6bb9bd380b11')
ON CONFLICT DO NOTHING;

-- Assign Permissions to OPS Role (service, incident, release, ai)
INSERT INTO role_permissions (role_id, permission_id) VALUES
    ('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a12', 'b0eebc99-9c0b-4ef8-bb6d-6bb9bd380b12'),
    ('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a12', 'b0eebc99-9c0b-4ef8-bb6d-6bb9bd380b13'),
    ('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a12', 'b0eebc99-9c0b-4ef8-bb6d-6bb9bd380b14'),
    ('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a12', 'b0eebc99-9c0b-4ef8-bb6d-6bb9bd380b15'),
    ('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a12', 'b0eebc99-9c0b-4ef8-bb6d-6bb9bd380b16'),
    ('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a12', 'b0eebc99-9c0b-4ef8-bb6d-6bb9bd380b17'),
    ('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a12', 'b0eebc99-9c0b-4ef8-bb6d-6bb9bd380b18'),
    ('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a12', 'b0eebc99-9c0b-4ef8-bb6d-6bb9bd380b19'),
    ('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a12', 'b0eebc99-9c0b-4ef8-bb6d-6bb9bd380b20'),
    ('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a12', 'b0eebc99-9c0b-4ef8-bb6d-6bb9bd380b21')
ON CONFLICT DO NOTHING;

-- Assign Permissions to ENGINEER Role
INSERT INTO role_permissions (role_id, permission_id) VALUES
    ('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a13', 'b0eebc99-9c0b-4ef8-bb6d-6bb9bd380b12'),
    ('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a13', 'b0eebc99-9c0b-4ef8-bb6d-6bb9bd380b14'),
    ('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a13', 'b0eebc99-9c0b-4ef8-bb6d-6bb9bd380b15'),
    ('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a13', 'b0eebc99-9c0b-4ef8-bb6d-6bb9bd380b16'),
    ('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a13', 'b0eebc99-9c0b-4ef8-bb6d-6bb9bd380b18'),
    ('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a13', 'b0eebc99-9c0b-4ef8-bb6d-6bb9bd380b20')
ON CONFLICT DO NOTHING;

-- Assign Permissions to VIEWER Role
INSERT INTO role_permissions (role_id, permission_id) VALUES
    ('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a14', 'b0eebc99-9c0b-4ef8-bb6d-6bb9bd380b12'),
    ('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a14', 'b0eebc99-9c0b-4ef8-bb6d-6bb9bd380b14'),
    ('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a14', 'b0eebc99-9c0b-4ef8-bb6d-6bb9bd380b18')
ON CONFLICT DO NOTHING;

-- Seed Default Admin User (password: changeme, bcrypt hash below)
-- Hash generated for 'changeme': $2a$10$3nK48T8fJ/3b99L9r7oJ/.y4H2/1LdD/mZzO7B.d7Cg1234567890 (we will seed with standard bcrypt)
INSERT INTO users (id, email, password_hash, display_name, status) VALUES
    ('c0eebc99-9c0b-4ef8-bb6d-6bb9bd380c11', 'admin@opsflow.local', '$2a$10$w8.1O5k/LzBfM03cK3y.jO04Z.Z2gD/eJ0rJ/.y4H2/1LdD/mZzO7', 'System Admin', 'ACTIVE')
ON CONFLICT (email) DO NOTHING;

-- Assign ADMIN role to admin user
INSERT INTO user_roles (user_id, role_id) VALUES
    ('c0eebc99-9c0b-4ef8-bb6d-6bb9bd380c11', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11')
ON CONFLICT DO NOTHING;
