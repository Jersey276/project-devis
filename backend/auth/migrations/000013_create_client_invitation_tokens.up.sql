CREATE TABLE client_invitation_tokens (
    id          SERIAL PRIMARY KEY,
    client_id   TEXT        NOT NULL,
    provider_id TEXT        NOT NULL,
    token_hash  TEXT        NOT NULL UNIQUE,
    expires_at  TIMESTAMPTZ NOT NULL,
    used_at     TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX ON client_invitation_tokens (token_hash);
