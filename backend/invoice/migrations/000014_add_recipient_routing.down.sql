-- Restore the 000013 immutability trigger (without recipient_routing_id), then
-- drop the column.
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

    IF OLD.status = 'DRAFT' THEN
        RETURN NEW;
    END IF;

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

    IF legal_frozen
       AND OLD.status = 'ISSUED' AND NEW.status = 'PAID'
       AND NEW.lifecycle_status   IS NOT DISTINCT FROM OLD.lifecycle_status
       AND NEW.pdp_submission_id  IS NOT DISTINCT FROM OLD.pdp_submission_id
    THEN
        RETURN NEW;
    END IF;

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

ALTER TABLE invoices DROP COLUMN recipient_routing_id;
