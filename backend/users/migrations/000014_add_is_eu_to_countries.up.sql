-- is_eu marks a country as an EU member state, used to trigger OSS distance-
-- selling VAT. This is a stable geopolitical fact, independent of any seller's
-- country_group tax configuration — hence a flag on countries, not a group.
ALTER TABLE countries
  ADD COLUMN is_eu BOOLEAN NOT NULL DEFAULT FALSE;
