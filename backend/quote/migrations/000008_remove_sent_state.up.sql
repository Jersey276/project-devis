-- The 'sent' state is absorbed into 'negociation': putting a quote into
-- negotiation is now the act of sending it to the client. Backfill any existing
-- 'sent' quotes to 'negociation', then tighten the CHECK constraint.
UPDATE quotes SET state = 'negociation' WHERE state = 'sent';

ALTER TABLE quotes DROP CONSTRAINT IF EXISTS quotes_state_check;

ALTER TABLE quotes
    ADD CONSTRAINT quotes_state_check
    CHECK (state IN ('draft', 'negociation', 'validated', 'drop'));
