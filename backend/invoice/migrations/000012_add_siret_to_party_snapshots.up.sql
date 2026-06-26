-- SIRET (14-digit establishment id) frozen at issue time: the recipient routing
-- key for B2B e-invoicing. '' default is safe for legacy rows (NOT NULL).
ALTER TABLE invoice_party_snapshots
    ADD COLUMN issuer_siret TEXT NOT NULL DEFAULT '',
    ADD COLUMN client_siret TEXT NOT NULL DEFAULT '';

-- Mirrored on credit notes, which inherit the source invoice's parties.
ALTER TABLE credit_note_party_snapshots
    ADD COLUMN issuer_siret TEXT NOT NULL DEFAULT '',
    ADD COLUMN client_siret TEXT NOT NULL DEFAULT '';
