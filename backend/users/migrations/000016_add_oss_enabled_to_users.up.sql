-- oss_enabled is the per-seller opt-in for the OSS distance-selling regime. When
-- true, B2C invoices to EU (non-FR) clients apply the destination country VAT
-- rate. Defaults to FALSE: sellers opt in once over the 10 000 € threshold (the
-- threshold itself is tracked manually for now — automatic accrual is a later lot).
ALTER TABLE users
  ADD COLUMN oss_enabled BOOLEAN NOT NULL DEFAULT FALSE;
