CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    service_name TEXT NOT NULL CHECK (btrim(service_name) <> ''),
    price INTEGER NOT NULL CHECK (price > 0),
    user_id UUID NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CHECK (end_date IS NULL OR end_date >= start_date)
);

CREATE INDEX idx_subscriptions_user_id ON subscriptions(user_id);
CREATE INDEX idx_subscriptions_service_name ON subscriptions(service_name);
CREATE INDEX idx_subscriptions_period ON subscriptions(start_date, end_date);

CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER subscriptions_set_updated_at
BEFORE UPDATE ON subscriptions
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();
