ALTER TABLE quotes
    DROP COLUMN IF EXISTS issued_at,
    DROP COLUMN IF EXISTS valid_until,
    DROP COLUMN IF EXISTS payment_terms;
