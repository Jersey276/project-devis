-- Pre-production schema: NOT NULL backfill is not feasible for existing rows.
DELETE FROM quotes;

ALTER TABLE quotes ADD COLUMN client_id  TEXT    NOT NULL;
ALTER TABLE quotes ADD COLUMN address_id INTEGER NOT NULL;

CREATE INDEX idx_quotes_client_id ON quotes(client_id);
