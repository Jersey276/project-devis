ALTER TABLE credit_note_party_snapshots
    DROP COLUMN IF EXISTS client_country_code,
    DROP COLUMN IF EXISTS issuer_country_code;

ALTER TABLE invoice_party_snapshots
    DROP COLUMN IF EXISTS client_country_code,
    DROP COLUMN IF EXISTS issuer_country_code;
