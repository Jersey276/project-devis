-- Seller payment instructions for EN 16931 invoices (BG-16). Nullable: optional
-- and absent on legacy rows.
ALTER TABLE users
  ADD COLUMN iban TEXT,
  ADD COLUMN bic  TEXT;
