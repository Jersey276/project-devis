-- E-invoicing lifecycle (réforme FR B2B). Orthogonal to the business `status`:
-- it tracks the invoice's journey on the platform (PPF/PDP). 'NONE' = no platform
-- lifecycle yet (every existing/legacy invoice). For now the five real statuses are
-- set manually by the issuer (B3); the platform auto-feed comes in B6.
ALTER TABLE invoices
    ADD COLUMN lifecycle_status TEXT NOT NULL DEFAULT 'NONE'
        CHECK (lifecycle_status IN
            ('NONE','DEPOSITED','RECEIVED','APPROVED','REJECTED','COLLECTED'));

-- Append-only, timestamped, traceable history. Never UPDATE/DELETE rows here.
CREATE TABLE invoice_lifecycle_events (
    id          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    invoice_id  TEXT NOT NULL REFERENCES invoices(invoice_id) ON DELETE CASCADE,
    user_id     TEXT NOT NULL,
    status      TEXT NOT NULL
                  CHECK (status IN ('DEPOSITED','RECEIVED','APPROVED','REJECTED','COLLECTED')),
    note        TEXT NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX invoice_lifecycle_events_invoice_idx
    ON invoice_lifecycle_events (invoice_id, created_at);
