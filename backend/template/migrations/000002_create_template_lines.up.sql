CREATE TABLE IF NOT EXISTS template_lines (
    id          BIGSERIAL PRIMARY KEY,
    line_id     UUID        NOT NULL DEFAULT gen_random_uuid() UNIQUE,
    template_id UUID        NOT NULL REFERENCES templates(template_id) ON DELETE CASCADE,
    type        TEXT        NOT NULL,
    name        TEXT        NOT NULL DEFAULT '',
    quantity    DECIMAL(12,3) NOT NULL DEFAULT 1,
    unit        TEXT,
    unit_price  BIGINT      NOT NULL DEFAULT 0,
    data        JSONB       NOT NULL DEFAULT '{}',
    position    INT         NOT NULL DEFAULT 0,
    tax_id      INT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_template_lines_template_id ON template_lines (template_id);
