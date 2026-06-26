-- Links a provider identity (provider + stable provider_user_id) to a user.
-- A single user_id may have several rows (one per provider) when identities are
-- linked by email. user_id is the opaque id owned by the users service (TEXT,
-- no FK — consistent with refresh_tokens / email_verification_tokens).
CREATE TABLE IF NOT EXISTS oauth_identities (
    id               SERIAL PRIMARY KEY,
    provider         TEXT NOT NULL,
    provider_user_id TEXT NOT NULL,
    user_id          TEXT NOT NULL,
    email            TEXT NOT NULL,
    created_at       TIMESTAMP DEFAULT NOW(),
    UNIQUE (provider, provider_user_id)
);

CREATE INDEX IF NOT EXISTS idx_oauth_identities_user_id ON oauth_identities (user_id);
