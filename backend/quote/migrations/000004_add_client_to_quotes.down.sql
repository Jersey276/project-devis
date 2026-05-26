DROP INDEX IF EXISTS idx_quotes_client_id;
ALTER TABLE quotes DROP COLUMN IF EXISTS address_id;
ALTER TABLE quotes DROP COLUMN IF EXISTS client_id;
