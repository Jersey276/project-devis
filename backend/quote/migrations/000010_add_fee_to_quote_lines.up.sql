-- Link quote lines back to a fee catalog entry so that updating a fee can
-- propagate the new snapshot to non-validated quotes. The column is nullable:
-- only lines created from a fee carry a fee_id. The line type stays
-- 'simple' | 'multiple' (a fee is a `kind` inside data, not a new type).
ALTER TABLE quote_lines ADD COLUMN fee_id TEXT;

CREATE INDEX idx_quote_lines_fee_id ON quote_lines(fee_id) WHERE fee_id IS NOT NULL;
