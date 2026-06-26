-- Dedicated AV- sequence (separate from invoices; purely additive). Continuous,
-- gap-free numbering is a French legal requirement, same as for invoices.
CREATE TABLE credit_note_number_sequences (
    user_id    TEXT    NOT NULL,
    year       INTEGER NOT NULL,
    last_value INTEGER NOT NULL DEFAULT 0 CHECK (last_value >= 0),
    PRIMARY KEY (user_id, year)
);

-- A credit note (avoir) neutralises part or all of an issued invoice. The
-- invoice itself keeps its status (a credit note never cancels it).
CREATE TABLE credit_notes (
    credit_note_id     TEXT PRIMARY KEY,
    user_id            TEXT NOT NULL,
    invoice_id         TEXT NOT NULL REFERENCES invoices(invoice_id) ON DELETE RESTRICT,
    credit_note_number TEXT NOT NULL,                  -- 'AV-YYYY-NNNN'
    number_year        INTEGER NOT NULL,
    number_seq         INTEGER NOT NULL,
    is_total           BOOLEAN NOT NULL DEFAULT FALSE,
    reason             TEXT NOT NULL DEFAULT '',
    currency           TEXT NOT NULL DEFAULT 'EUR' CHECK (currency = 'EUR'),
    issued_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    -- Amounts stored POSITIVE (semantics "credited amount"); negated at render time.
    total_ht_cents     BIGINT NOT NULL DEFAULT 0,
    total_vat_cents    BIGINT NOT NULL DEFAULT 0,
    total_ttc_cents    BIGINT NOT NULL DEFAULT 0,
    vat_exempt         BOOLEAN NOT NULL DEFAULT FALSE,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX credit_notes_unique_number_per_user_idx
    ON credit_notes (user_id, credit_note_number);
CREATE INDEX credit_notes_invoice_idx ON credit_notes (invoice_id);
CREATE INDEX credit_notes_user_idx    ON credit_notes (user_id);

-- Frozen party snapshot, re-captured for independent immutability (mirrors
-- invoice_party_snapshots).
CREATE TABLE credit_note_party_snapshots (
    credit_note_id    TEXT PRIMARY KEY REFERENCES credit_notes(credit_note_id) ON DELETE CASCADE,
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

-- Credited lines = frozen snapshot + over-crediting registry. The UNIQUE
-- constraint on (origin_invoice_id, origin_position) guarantees a given invoice
-- line can be credited at most once across all credit notes.
CREATE TABLE credit_note_lines (
    credit_note_id    TEXT    NOT NULL REFERENCES credit_notes(credit_note_id) ON DELETE CASCADE,
    position          INTEGER NOT NULL,               -- position within the credit note
    origin_invoice_id TEXT    NOT NULL,                -- = credit_notes.invoice_id (denormalised)
    origin_position   INTEGER NOT NULL,                -- position in invoice_line_snapshots
    quote_line_id     TEXT    NOT NULL,
    name              TEXT    NOT NULL DEFAULT '',
    unit              TEXT    NOT NULL DEFAULT '',
    quantity          TEXT    NOT NULL DEFAULT '',
    unit_price_cents  BIGINT  NOT NULL DEFAULT 0,
    line_ht_cents     BIGINT  NOT NULL DEFAULT 0,
    tax_id            INTEGER NOT NULL DEFAULT 0,
    tax_rate          TEXT    NOT NULL DEFAULT '0',
    tax_label         TEXT    NOT NULL DEFAULT '',
    PRIMARY KEY (credit_note_id, position),
    CONSTRAINT credit_note_lines_origin_unique UNIQUE (origin_invoice_id, origin_position)
);

CREATE INDEX credit_note_lines_origin_idx ON credit_note_lines (origin_invoice_id);

CREATE TABLE credit_note_vat_breakdown_snapshots (
    credit_note_id TEXT    NOT NULL REFERENCES credit_notes(credit_note_id) ON DELETE CASCADE,
    tax_rate       TEXT    NOT NULL,
    base_ht_cents  BIGINT  NOT NULL DEFAULT 0,
    vat_cents      BIGINT  NOT NULL DEFAULT 0,
    PRIMARY KEY (credit_note_id, tax_rate)
);
