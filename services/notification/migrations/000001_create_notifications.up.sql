-- 000001_create_notifications.up.sql

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID,
    channel VARCHAR(50) NOT NULL,
    event_type VARCHAR(100) NOT NULL,
    payload JSONB NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'SENT',
    attempts INT NOT NULL DEFAULT 1,
    idempotency_key VARCHAR(255) NOT NULL UNIQUE,
    sent_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_notifications_idempotency ON notifications(idempotency_key);
CREATE INDEX IF NOT EXISTS idx_notifications_user_id ON notifications(user_id);
