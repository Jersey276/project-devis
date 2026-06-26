-- E-reporting des transactions (B5) et transfrontalier B2C (C5): périodique
-- aggregate transmission to the platform, distinct from the per-invoice deposit
-- (B6). One row per (issuer, kind, civil month); not legal documents (no seal,
-- no inalterability), so this is a plain mutable status table — the immutability
-- trigger on `invoices` is untouched.
-- TRANSACTION = domestic B2C sales; CROSS_BORDER_B2C = intra-EU distance sales
-- (the OSS scope). status reuses the B3 lifecycle vocabulary for a homogeneous
-- view alongside invoices.
CREATE TABLE invoice_reports (
    id              BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id         TEXT NOT NULL,
    kind            TEXT NOT NULL CHECK (kind IN ('TRANSACTION','CROSS_BORDER_B2C')),
    period_year     INT  NOT NULL,
    period_month    INT  NOT NULL CHECK (period_month BETWEEN 1 AND 12),
    status          TEXT NOT NULL DEFAULT 'NONE'
                      CHECK (status IN ('NONE','DEPOSITED','RECEIVED','APPROVED','REJECTED','COLLECTED')),
    total_ht_cents  BIGINT NOT NULL,
    total_vat_cents BIGINT NOT NULL,
    report_id       TEXT,            -- platform handle; NULL under the no-op adapter
    submitted_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    -- One report per period/kind: prevents double submission, enables a controlled
    -- re-submission via ON CONFLICT.
    UNIQUE (user_id, kind, period_year, period_month)
);

-- The poller scans submitted, non-terminal reports.
CREATE INDEX invoice_reports_poll_idx ON invoice_reports (status) WHERE report_id IS NOT NULL;
