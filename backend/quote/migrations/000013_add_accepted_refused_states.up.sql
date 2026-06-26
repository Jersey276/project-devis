ALTER TABLE quotes DROP CONSTRAINT IF EXISTS quotes_state_check;

ALTER TABLE quotes
    ADD CONSTRAINT quotes_state_check
    CHECK (state IN ('draft', 'negociation', 'validated', 'drop', 'accepted', 'refused'));
