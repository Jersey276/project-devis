ALTER TABLE quotes
    ADD COLUMN issued_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ADD COLUMN valid_until   DATE        NULL,
    ADD COLUMN payment_terms TEXT        NULL;
