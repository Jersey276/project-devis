CREATE TABLE consents (
    id          SERIAL PRIMARY KEY,
    user_id     TEXT        NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    type        TEXT        NOT NULL CHECK (type IN ('cgv', 'privacy_policy', 'cookies')),
    version     TEXT        NOT NULL,
    accepted_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ip          INET,
    mechanism   TEXT        NOT NULL DEFAULT 'checkbox'
);

CREATE INDEX ON consents (user_id, type);
