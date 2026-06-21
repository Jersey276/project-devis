-- Recipient routing handle resolved from the e-invoicing directory (annuaire
-- DGFiP/AIFE) at deposit time (B6). NULL until deposited, or when the no-op
-- directory is in use (which resolves with an empty handle). Frozen once written,
-- like pdp_submission_id. Nullable + additive: no backfill, no constraint.
ALTER TABLE invoices ADD COLUMN recipient_routing_id TEXT;

-- Extend the 000013 immutability trigger to treat recipient_routing_id as
-- operational e-invoicing metadata (mutable while sealed, alongside
-- lifecycle_status / pdp_submission_id), so the deposit flow can freeze it on a
-- sealed invoice. Every legal / financial column still stays frozen.
-- MAINTENANCE: any future column on `invoices` MUST be added to the frozen list
-- below unless it is deliberately mutable-while-sealed.
CREATE OR REPLACE FUNCTION trg_invoices_immutable() RETURNS trigger AS $$
DECLARE
    legal_frozen BOOLEAN;
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

    -- Every legal / financial column unchanged. status/paid_at/updated_at and the
    -- operational columns (lifecycle_status, pdp_submission_id,
    -- recipient_routing_id) are intentionally excluded so the branches below can
    -- vary them.
    legal_frozen :=
           NEW.invoice_id           IS NOT DISTINCT FROM OLD.invoice_id
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
       AND NEW.created_at           IS NOT DISTINCT FROM OLD.created_at;

    -- (a) ISSUED -> PAID, touching only status/paid_at/updated_at.
    IF legal_frozen
       AND OLD.status = 'ISSUED' AND NEW.status = 'PAID'
       AND NEW.lifecycle_status     IS NOT DISTINCT FROM OLD.lifecycle_status
       AND NEW.pdp_submission_id    IS NOT DISTINCT FROM OLD.pdp_submission_id
       AND NEW.recipient_routing_id IS NOT DISTINCT FROM OLD.recipient_routing_id
    THEN
        RETURN NEW;
    END IF;

    -- (b) e-invoicing metadata move on a sealed row: status unchanged, only
    -- lifecycle_status / pdp_submission_id / recipient_routing_id / updated_at
    -- may differ.
    IF legal_frozen
       AND NEW.status IS NOT DISTINCT FROM OLD.status
       AND NEW.paid_at IS NOT DISTINCT FROM OLD.paid_at
    THEN
        RETURN NEW;
    END IF;

    RAISE EXCEPTION 'invoice % is sealed and cannot be modified', OLD.invoice_id
        USING ERRCODE = 'integrity_constraint_violation';
END;
$$ LANGUAGE plpgsql;
