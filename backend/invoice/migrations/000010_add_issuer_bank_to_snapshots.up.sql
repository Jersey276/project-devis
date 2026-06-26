-- Issuer payment instructions (BG-16) frozen at issue time. '' default is safe
-- for legacy rows and blank entries (NOT NULL).
ALTER TABLE invoice_party_snapshots
    ADD COLUMN issuer_iban TEXT NOT NULL DEFAULT '',
    ADD COLUMN issuer_bic  TEXT NOT NULL DEFAULT '';

-- Mirrored on credit notes, which inherit the source invoice's instructions.
ALTER TABLE credit_note_party_snapshots
    ADD COLUMN issuer_iban TEXT NOT NULL DEFAULT '',
    ADD COLUMN issuer_bic  TEXT NOT NULL DEFAULT '';
