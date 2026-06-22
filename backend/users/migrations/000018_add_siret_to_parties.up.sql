-- SIRET (14 digits = SIREN + NIC establishment) routes B2B e-invoices to the
-- recipient establishment in the DGFiP directory. Nullable: optional and absent
-- on legacy rows.
ALTER TABLE users   ADD COLUMN siret TEXT;
ALTER TABLE clients ADD COLUMN siret TEXT;
