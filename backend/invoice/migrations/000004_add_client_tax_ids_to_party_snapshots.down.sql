ALTER TABLE invoice_party_snapshots
    DROP COLUMN IF EXISTS client_siren,
    DROP COLUMN IF EXISTS client_vat;
