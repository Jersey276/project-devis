CREATE TABLE IF NOT EXISTS email_change_tokens (
    id          SERIAL PRIMARY KEY,
    user_id     TEXT      NOT NULL,
    new_email   TEXT      NOT NULL,
    token_hash  TEXT      NOT NULL UNIQUE,
    expires_at  TIMESTAMP NOT NULL,
    used_at     TIMESTAMP NULL,
    created_at  TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_email_change_tokens_user_id ON email_change_tokens (user_id);
