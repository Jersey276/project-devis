-- OAuth-only accounts have no password. Relax the NOT NULL constraint so they
-- can exist. The "password OR oauth identity" invariant is enforced in the
-- application (Postgres CHECK constraints cannot reference other tables).
ALTER TABLE auth ALTER COLUMN password DROP NOT NULL;
