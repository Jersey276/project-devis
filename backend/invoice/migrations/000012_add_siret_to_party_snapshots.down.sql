ALTER TABLE credit_note_party_snapshots
    DROP COLUMN issuer_siret,
    DROP COLUMN client_siret;

ALTER TABLE invoice_party_snapshots
    DROP COLUMN issuer_siret,
    DROP COLUMN client_siret;
