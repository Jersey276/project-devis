DROP INDEX IF EXISTS idx_taxes_one_default_per_group;
ALTER TABLE taxes DROP COLUMN IF EXISTS is_default;
