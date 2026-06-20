CREATE TABLE fees (
    id          SERIAL      PRIMARY KEY,
    fee_id      TEXT        NOT NULL UNIQUE,
    user_id     TEXT        NOT NULL,
    category    TEXT        NOT NULL CHECK (category IN ('fixed', 'service')),
    name        TEXT        NOT NULL,
    unit        TEXT,
    unit_price  BIGINT      NOT NULL DEFAULT 0,
    tax_id      INTEGER,
    archived_at TIMESTAMP,
    created_at  TIMESTAMP   NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMP   NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_fees_user_id ON fees(user_id);
