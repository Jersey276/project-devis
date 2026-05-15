-- Versioning columns. NULL original_tax_id means "this row IS the original"
-- (avoids the chicken-and-egg of a self-referencing NOT NULL FK on insert).
ALTER TABLE taxes ADD COLUMN original_tax_id INTEGER NULL REFERENCES taxes(id);
ALTER TABLE taxes ADD COLUMN version         INTEGER NOT NULL DEFAULT 1;
ALTER TABLE taxes ADD COLUMN superseded_at   TIMESTAMP NULL;
ALTER TABLE taxes ADD COLUMN superseded_by   INTEGER NULL REFERENCES taxes(id);

CREATE INDEX idx_taxes_original_tax_id
    ON taxes(original_tax_id) WHERE original_tax_id IS NOT NULL;

-- Serialise version numbers within a family so two concurrent updates
-- can't both compute MAX(version)+1 and collide. COALESCE(original_tax_id, id)
-- groups originals with their descendants under a single family key.
CREATE UNIQUE INDEX idx_taxes_unique_version_per_family
    ON taxes ((COALESCE(original_tax_id, id)), version);

CREATE INDEX idx_taxes_current
    ON taxes(country_group_id) WHERE superseded_at IS NULL;

DROP INDEX IF EXISTS idx_taxes_one_default_per_group;
CREATE UNIQUE INDEX idx_taxes_one_default_per_group
    ON taxes(country_group_id)
    WHERE is_default = TRUE AND superseded_at IS NULL;
