ALTER TABLE credit_note_party_snapshots
    DROP COLUMN IF EXISTS client_type;

ALTER TABLE invoice_party_snapshots
    DROP COLUMN IF EXISTS client_type;
