-- Revert to the original state set. Any rows in 'negociation' are moved back to
-- 'sent' so the stricter constraint can be applied without data loss.
UPDATE quotes SET state = 'sent' WHERE state = 'negociation';

ALTER TABLE quotes DROP CONSTRAINT IF EXISTS quotes_state_check;

ALTER TABLE quotes
    ADD CONSTRAINT quotes_state_check
    CHECK (state IN ('draft', 'sent', 'validated', 'drop'));
