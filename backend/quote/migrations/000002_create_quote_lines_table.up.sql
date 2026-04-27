CREATE TABLE quote_lines (
    id          SERIAL          PRIMARY KEY,
    line_id     TEXT            NOT NULL UNIQUE,
    quote_id    TEXT            NOT NULL REFERENCES quotes(quote_id) ON DELETE CASCADE,
    type        TEXT            NOT NULL CHECK (type IN ('simple', 'multiple')),
    name        TEXT            NOT NULL,
    quantity    DECIMAL(12, 3)  NOT NULL DEFAULT 1,
    unit        TEXT,
    unit_price  BIGINT          NOT NULL DEFAULT 0,
    data        JSONB           NOT NULL DEFAULT '{}'::jsonb,
    position    INTEGER         NOT NULL DEFAULT 0,
    created_at  TIMESTAMP       NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMP       NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_quote_lines_quote_id ON quote_lines(quote_id);
