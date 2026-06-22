ALTER TABLE credit_note_party_snapshots
    DROP COLUMN IF EXISTS client_country_id;

ALTER TABLE invoice_party_snapshots
    DROP COLUMN IF EXISTS client_country_id;
