ALTER TABLE quotes
    ADD COLUMN state TEXT NOT NULL DEFAULT 'draft'
        CHECK (state IN ('draft', 'sent', 'validated', 'drop'));

CREATE INDEX idx_quotes_state ON quotes(state);
