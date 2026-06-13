-- Restore the 'sent' state in the constraint. We cannot reliably distinguish
-- which 'negociation' rows were originally 'sent', so rows are left in
-- 'negociation' (a valid value in the wider set) — no data is lost.
ALTER TABLE quotes DROP CONSTRAINT IF EXISTS quotes_state_check;

ALTER TABLE quotes
    ADD CONSTRAINT quotes_state_check
    CHECK (state IN ('draft', 'sent', 'negociation', 'validated', 'drop'));
