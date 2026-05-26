ALTER TABLE taxes ADD COLUMN is_default BOOLEAN NOT NULL DEFAULT FALSE;

CREATE UNIQUE INDEX idx_taxes_one_default_per_group
    ON taxes(country_group_id) WHERE is_default = TRUE;
