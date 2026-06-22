-- Freeze the B2C/B2B nature of the client on the invoice and credit-note party
-- snapshots. Empty string ('') means "legacy snapshot, type unknown", matching
-- the NOT NULL DEFAULT '' convention of the other party-snapshot columns.
ALTER TABLE invoice_party_snapshots
    ADD COLUMN client_type TEXT NOT NULL DEFAULT '';

ALTER TABLE credit_note_party_snapshots
    ADD COLUMN client_type TEXT NOT NULL DEFAULT '';
