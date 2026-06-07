ALTER TABLE plans ADD COLUMN stripe_price_id TEXT;

ALTER TABLE subscriptions
    ADD COLUMN stripe_customer_id     TEXT,
    ADD COLUMN stripe_subscription_id TEXT,
    ADD COLUMN cancel_at_period_end   BOOLEAN NOT NULL DEFAULT FALSE;

CREATE TABLE subscription_events (
    event_id        TEXT PRIMARY KEY,
    user_id         TEXT NOT NULL,
    stripe_event_id TEXT UNIQUE,
    event_type      TEXT NOT NULL,
    payload         JSONB NOT NULL DEFAULT '{}',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
