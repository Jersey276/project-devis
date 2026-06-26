DROP INDEX IF EXISTS idx_quote_lines_fee_id;
ALTER TABLE quote_lines DROP COLUMN IF EXISTS fee_id;
