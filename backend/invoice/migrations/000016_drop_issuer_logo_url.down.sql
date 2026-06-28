ALTER TABLE invoice_party_snapshots ADD COLUMN issuer_logo_url TEXT NOT NULL DEFAULT '';
ALTER TABLE credit_note_party_snapshots ADD COLUMN issuer_logo_url TEXT NOT NULL DEFAULT '';
