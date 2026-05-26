DROP INDEX IF EXISTS idx_taxes_one_default_per_group;
DROP INDEX IF EXISTS idx_taxes_current;
DROP INDEX IF EXISTS idx_taxes_unique_version_per_family;
DROP INDEX IF EXISTS idx_taxes_original_tax_id;

ALTER TABLE taxes DROP COLUMN IF EXISTS superseded_by;
ALTER TABLE taxes DROP COLUMN IF EXISTS superseded_at;
ALTER TABLE taxes DROP COLUMN IF EXISTS version;
ALTER TABLE taxes DROP COLUMN IF EXISTS original_tax_id;

CREATE UNIQUE INDEX idx_taxes_one_default_per_group
    ON taxes(country_group_id) WHERE is_default = TRUE;
