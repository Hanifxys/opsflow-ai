-- =============================================================================
-- OpsFlow — Database Initialization
-- =============================================================================
-- This script runs on first PostgreSQL container start.
-- It ensures the opsflow database and extensions exist.

-- Enable commonly needed extensions.
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Log initialization.
DO $$
BEGIN
  RAISE NOTICE 'OpsFlow database initialized successfully';
END
$$;
