CREATE TABLE schedules (
    schedule_id TEXT PRIMARY KEY,
    quote_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    name TEXT NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('DRAFT', 'NEGOCIATE', 'DENIED', 'VALID')),
    start_month DATE NOT NULL,
    duration_months INTEGER NOT NULL CHECK (duration_months > 0),
    currency TEXT NOT NULL DEFAULT 'EUR' CHECK (currency = 'EUR'),
    validated_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE schedule_cells (
    schedule_id TEXT NOT NULL REFERENCES schedules(schedule_id) ON DELETE CASCADE,
    quote_line_id TEXT NOT NULL,
    month_index INTEGER NOT NULL CHECK (month_index >= 1),
    amount_cents BIGINT NOT NULL DEFAULT 0 CHECK (amount_cents >= 0),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (schedule_id, quote_line_id, month_index)
);

CREATE UNIQUE INDEX schedules_single_valid_per_quote_idx
    ON schedules (quote_id)
    WHERE status = 'VALID';