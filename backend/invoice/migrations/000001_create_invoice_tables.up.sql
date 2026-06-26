-- Sequential per-(user, year) counter. Continuous, gap-free numbering is a
-- French legal requirement (art. 289 CGI). The number is consumed only inside
-- the issue transaction; a rollback leaves the counter untouched.
CREATE TABLE invoice_number_sequences (
    user_id    TEXT    NOT NULL,
    year       INTEGER NOT NULL,
    last_value INTEGER NOT NULL DEFAULT 0 CHECK (last_value >= 0),
    PRIMARY KEY (user_id, year)
);

CREATE TABLE invoices (
    invoice_id           TEXT PRIMARY KEY,
    user_id              TEXT NOT NULL,
    quote_id             TEXT NOT NULL,
    schedule_id          TEXT,                         -- NULL = invoice for the whole quote
    billed_month_indexes INTEGER[] NOT NULL DEFAULT '{}',
    status               TEXT NOT NULL DEFAULT 'DRAFT'
                            CHECK (status IN ('DRAFT', 'ISSUED', 'PAID', 'CANCELLED')),
    currency             TEXT NOT NULL DEFAULT 'EUR' CHECK (currency = 'EUR'),
    invoice_number       TEXT,                          -- 'YYYY-NNNN', NULL while DRAFT
    number_year          INTEGER,
    number_seq           INTEGER,
    issued_at            TIMESTAMPTZ,
    sale_date            DATE,
    due_date             DATE,
    paid_at              TIMESTAMPTZ,
    cancelled_at         TIMESTAMPTZ,
    total_ht_cents       BIGINT NOT NULL DEFAULT 0,
    total_vat_cents      BIGINT NOT NULL DEFAULT 0,
    total_ttc_cents      BIGINT NOT NULL DEFAULT 0,
    vat_exempt           BOOLEAN NOT NULL DEFAULT FALSE, -- art. 293 B CGI (franchise)
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX invoices_unique_number_per_user_idx
    ON invoices (user_id, invoice_number) WHERE invoice_number IS NOT NULL;
CREATE INDEX invoices_quote_idx    ON invoices (user_id, quote_id);
CREATE INDEX invoices_schedule_idx ON invoices (schedule_id) WHERE schedule_id IS NOT NULL;

-- Frozen snapshot of the parties. users/clients/addresses stay mutable upstream,
-- so the legally-binding mentions must be copied at issue time.
CREATE TABLE invoice_party_snapshots (
    invoice_id        TEXT PRIMARY KEY REFERENCES invoices(invoice_id) ON DELETE CASCADE,
    issuer_company    TEXT NOT NULL DEFAULT '',
    issuer_siren      TEXT NOT NULL DEFAULT '',
    issuer_vat        TEXT NOT NULL DEFAULT '',
    issuer_email      TEXT NOT NULL DEFAULT '',
    issuer_phone      TEXT NOT NULL DEFAULT '',
    issuer_logo_url   TEXT NOT NULL DEFAULT '',
    issuer_street     TEXT NOT NULL DEFAULT '',
    issuer_additional TEXT NOT NULL DEFAULT '',
    issuer_zip        TEXT NOT NULL DEFAULT '',
    issuer_city       TEXT NOT NULL DEFAULT '',
    client_first_name TEXT NOT NULL DEFAULT '',
    client_last_name  TEXT NOT NULL DEFAULT '',
    client_company    TEXT NOT NULL DEFAULT '',
    client_email      TEXT NOT NULL DEFAULT '',
    client_street     TEXT NOT NULL DEFAULT '',
    client_additional TEXT NOT NULL DEFAULT '',
    client_zip        TEXT NOT NULL DEFAULT '',
    client_city       TEXT NOT NULL DEFAULT ''
);

-- Frozen snapshot of the billed lines. The VAT rate is captured here because
-- the upstream taxes table is versioned and its rate may change later.
CREATE TABLE invoice_line_snapshots (
    invoice_id       TEXT NOT NULL REFERENCES invoices(invoice_id) ON DELETE CASCADE,
    position         INTEGER NOT NULL,
    quote_line_id    TEXT NOT NULL,
    name             TEXT NOT NULL DEFAULT '',
    unit             TEXT NOT NULL DEFAULT '',
    quantity         TEXT NOT NULL DEFAULT '',
    unit_price_cents BIGINT NOT NULL DEFAULT 0,
    line_ht_cents    BIGINT NOT NULL DEFAULT 0,
    tax_id           INTEGER NOT NULL DEFAULT 0,
    tax_rate         TEXT NOT NULL DEFAULT '0',
    tax_label        TEXT NOT NULL DEFAULT '',
    PRIMARY KEY (invoice_id, position)
);

-- Frozen VAT breakdown per rate (base HT and VAT amount).
CREATE TABLE invoice_vat_breakdown_snapshots (
    invoice_id    TEXT NOT NULL REFERENCES invoices(invoice_id) ON DELETE CASCADE,
    tax_rate      TEXT NOT NULL,
    base_ht_cents BIGINT NOT NULL DEFAULT 0,
    vat_cents     BIGINT NOT NULL DEFAULT 0,
    PRIMARY KEY (invoice_id, tax_rate)
);
