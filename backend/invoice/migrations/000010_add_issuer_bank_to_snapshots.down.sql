ALTER TABLE invoice_party_snapshots
    DROP COLUMN IF EXISTS issuer_iban,
    DROP COLUMN IF EXISTS issuer_bic;

ALTER TABLE credit_note_party_snapshots
    DROP COLUMN IF EXISTS issuer_iban,
    DROP COLUMN IF EXISTS issuer_bic;
