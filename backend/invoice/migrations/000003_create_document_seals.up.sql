-- Technical inalterability (NF203 / anti-fraud): cryptographic chaining of
-- issued documents + DB-level write locks. One chain per issuer (user_id),
-- mixing invoices and credit notes in emission order.

CREATE TABLE document_seals (
    user_id      TEXT        NOT NULL,
    doc_type     TEXT        NOT NULL CHECK (doc_type IN ('INVOICE', 'CREDIT_NOTE')),
    doc_id       TEXT        NOT NULL,
    chain_index  BIGINT      NOT NULL,            -- 0-based, per user_id
    content_hash TEXT        NOT NULL,            -- sha256 hex of the legal content
    prev_hash    TEXT        NOT NULL,            -- predecessor chain_hash, or genesis
    chain_hash   TEXT        NOT NULL,            -- sha256(prev || content || index)
    sealed_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, doc_type, doc_id),
    CONSTRAINT document_seals_chain_unique UNIQUE (user_id, chain_index)
);
CREATE INDEX document_seals_chain_order_idx ON document_seals (user_id, chain_index);

-- Mutable chain head per issuer: the atomic allocation point (locked FOR UPDATE
-- by the sealing transaction).
CREATE TABLE chain_heads (
    user_id    TEXT   PRIMARY KEY,
    last_index BIGINT NOT NULL,
    last_hash  TEXT   NOT NULL
);

-- ─── Immutability triggers ───────────────────────────────────────────────────

-- document_seals is append-only: never updated or deleted.
CREATE FUNCTION trg_document_seals_immutable() RETURNS trigger AS $$
BEGIN
    RAISE EXCEPTION 'document_seals is append-only (inalterability): % forbidden', TG_OP
        USING ERRCODE = 'integrity_constraint_violation';
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER document_seals_no_update_delete
    BEFORE UPDATE OR DELETE ON document_seals
    FOR EACH ROW EXECUTE FUNCTION trg_document_seals_immutable();

-- invoices: a sealed invoice (ISSUED/PAID/CANCELLED) cannot be modified or
-- deleted, with two whitelisted exceptions:
--   (a) UPDATE while OLD.status='DRAFT' — the issue-time sealing UPDATE.
--   (b) UPDATE for the ISSUED->PAID transition touching ONLY status/paid_at/
--       updated_at (every other column compared via IS NOT DISTINCT FROM).
-- MAINTENANCE: any future column added to `invoices` MUST be added to the
-- IS NOT DISTINCT FROM list below, otherwise MarkInvoicePaid will start failing.
CREATE FUNCTION trg_invoices_immutable() RETURNS trigger AS $$
BEGIN
    IF TG_OP = 'DELETE' THEN
        IF OLD.status <> 'DRAFT' THEN
            RAISE EXCEPTION 'invoice % is sealed and cannot be deleted', OLD.invoice_id
                USING ERRCODE = 'integrity_constraint_violation';
        END IF;
        RETURN OLD;
    END IF;

    -- DRAFT rows are not yet sealed: allow (covers DRAFT->ISSUED sealing).
    IF OLD.status = 'DRAFT' THEN
        RETURN NEW;
    END IF;

    -- Sealed row: the only permitted change is ISSUED -> PAID touching
    -- status/paid_at/updated_at exclusively.
    IF OLD.status = 'ISSUED' AND NEW.status = 'PAID'
       AND NEW.invoice_id           IS NOT DISTINCT FROM OLD.invoice_id
       AND NEW.user_id              IS NOT DISTINCT FROM OLD.user_id
       AND NEW.quote_id             IS NOT DISTINCT FROM OLD.quote_id
       AND NEW.schedule_id          IS NOT DISTINCT FROM OLD.schedule_id
       AND NEW.billed_month_indexes IS NOT DISTINCT FROM OLD.billed_month_indexes
       AND NEW.currency             IS NOT DISTINCT FROM OLD.currency
       AND NEW.invoice_number       IS NOT DISTINCT FROM OLD.invoice_number
       AND NEW.number_year          IS NOT DISTINCT FROM OLD.number_year
       AND NEW.number_seq           IS NOT DISTINCT FROM OLD.number_seq
       AND NEW.issued_at            IS NOT DISTINCT FROM OLD.issued_at
       AND NEW.sale_date            IS NOT DISTINCT FROM OLD.sale_date
       AND NEW.due_date             IS NOT DISTINCT FROM OLD.due_date
       AND NEW.cancelled_at         IS NOT DISTINCT FROM OLD.cancelled_at
       AND NEW.total_ht_cents       IS NOT DISTINCT FROM OLD.total_ht_cents
       AND NEW.total_vat_cents      IS NOT DISTINCT FROM OLD.total_vat_cents
       AND NEW.total_ttc_cents      IS NOT DISTINCT FROM OLD.total_ttc_cents
       AND NEW.vat_exempt           IS NOT DISTINCT FROM OLD.vat_exempt
       AND NEW.created_at           IS NOT DISTINCT FROM OLD.created_at
    THEN
        RETURN NEW;
    END IF;

    RAISE EXCEPTION 'invoice % is sealed and cannot be modified', OLD.invoice_id
        USING ERRCODE = 'integrity_constraint_violation';
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER invoices_immutable
    BEFORE UPDATE OR DELETE ON invoices
    FOR EACH ROW EXECUTE FUNCTION trg_invoices_immutable();

-- credit_notes: always sealed (no draft state).
CREATE FUNCTION trg_credit_notes_immutable() RETURNS trigger AS $$
BEGIN
    RAISE EXCEPTION 'credit note % is sealed and cannot be modified or deleted', OLD.credit_note_id
        USING ERRCODE = 'integrity_constraint_violation';
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER credit_notes_immutable
    BEFORE UPDATE OR DELETE ON credit_notes
    FOR EACH ROW EXECUTE FUNCTION trg_credit_notes_immutable();

-- Snapshot rows are written once at issue time and never modified. Block all
-- UPDATE/DELETE so a direct write cannot silently alter hashed content.
CREATE FUNCTION trg_snapshot_immutable() RETURNS trigger AS $$
BEGIN
    RAISE EXCEPTION 'snapshot row in % is sealed and cannot be modified or deleted', TG_TABLE_NAME
        USING ERRCODE = 'integrity_constraint_violation';
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER invoice_party_snapshots_immutable
    BEFORE UPDATE OR DELETE ON invoice_party_snapshots
    FOR EACH ROW EXECUTE FUNCTION trg_snapshot_immutable();
CREATE TRIGGER invoice_line_snapshots_immutable
    BEFORE UPDATE OR DELETE ON invoice_line_snapshots
    FOR EACH ROW EXECUTE FUNCTION trg_snapshot_immutable();
CREATE TRIGGER invoice_vat_breakdown_snapshots_immutable
    BEFORE UPDATE OR DELETE ON invoice_vat_breakdown_snapshots
    FOR EACH ROW EXECUTE FUNCTION trg_snapshot_immutable();
CREATE TRIGGER credit_note_party_snapshots_immutable
    BEFORE UPDATE OR DELETE ON credit_note_party_snapshots
    FOR EACH ROW EXECUTE FUNCTION trg_snapshot_immutable();
CREATE TRIGGER credit_note_lines_immutable
    BEFORE UPDATE OR DELETE ON credit_note_lines
    FOR EACH ROW EXECUTE FUNCTION trg_snapshot_immutable();
CREATE TRIGGER credit_note_vat_breakdown_snapshots_immutable
    BEFORE UPDATE OR DELETE ON credit_note_vat_breakdown_snapshots
    FOR EACH ROW EXECUTE FUNCTION trg_snapshot_immutable();
