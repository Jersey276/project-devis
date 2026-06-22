-- Add client_type to distinguish individuals (B2C) from businesses (B2B).
-- Nullable for now: existing rows are backfilled in 000013, and the NOT NULL
-- constraint is deferred to a later migration once the front always sends it.
ALTER TABLE clients
  ADD COLUMN client_type TEXT;

ALTER TABLE clients
  ADD CONSTRAINT clients_client_type_check
  CHECK (client_type IS NULL OR client_type IN ('individual', 'business'));
