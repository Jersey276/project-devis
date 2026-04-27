CREATE TABLE quotes (
    id          SERIAL      PRIMARY KEY,
    quote_id    TEXT        NOT NULL UNIQUE,
    user_id     TEXT        NOT NULL,
    name        TEXT        NOT NULL,
    archived_at TIMESTAMP,
    created_at  TIMESTAMP   NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMP   NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_quotes_user_id ON quotes(user_id);
CREATE INDEX idx_quotes_archived_at ON quotes(archived_at);
