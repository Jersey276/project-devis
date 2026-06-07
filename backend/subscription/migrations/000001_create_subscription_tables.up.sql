CREATE TABLE plans (
    plan_id       SERIAL PRIMARY KEY,
    name          TEXT NOT NULL UNIQUE,
    tier          TEXT NOT NULL UNIQUE CHECK (tier IN ('free', 'pro', 'enterprise')),
    price_cents   INTEGER NOT NULL DEFAULT 0 CHECK (price_cents >= 0),
    billing_cycle TEXT NOT NULL DEFAULT 'none'
                  CHECK (billing_cycle IN ('monthly', 'yearly', 'none')),
    features      JSONB NOT NULL DEFAULT '{}',
    active        BOOLEAN NOT NULL DEFAULT TRUE,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE subscriptions (
    subscription_id      TEXT PRIMARY KEY,
    user_id              TEXT NOT NULL UNIQUE,
    plan_id              INTEGER NOT NULL REFERENCES plans(plan_id),
    status               TEXT NOT NULL DEFAULT 'active'
                         CHECK (status IN ('active', 'cancelled', 'expired')),
    current_period_start TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    current_period_end   TIMESTAMPTZ,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO plans (name, tier, price_cents, billing_cycle, features) VALUES
    ('Free',       'free',       0,    'none',    '{"max_schedules":3,"max_templates":2}'),
    ('Pro',        'pro',        900,  'monthly', '{"max_schedules":-1,"max_templates":-1}'),
    ('Enterprise', 'enterprise', 4900, 'monthly', '{"max_schedules":-1,"max_templates":-1}')
ON CONFLICT (tier) DO NOTHING;
