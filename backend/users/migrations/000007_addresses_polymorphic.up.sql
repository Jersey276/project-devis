ALTER TABLE addresses ADD COLUMN owner_type TEXT;
ALTER TABLE addresses ADD COLUMN owner_id   TEXT;

UPDATE addresses SET owner_type = 'user', owner_id = user_id;

ALTER TABLE addresses ALTER COLUMN owner_type SET NOT NULL;
ALTER TABLE addresses ALTER COLUMN owner_id   SET NOT NULL;

ALTER TABLE addresses ADD CONSTRAINT addresses_owner_type_check
    CHECK (owner_type IN ('user', 'client'));

ALTER TABLE addresses DROP CONSTRAINT addresses_user_id_fkey;
DROP INDEX idx_addresses_user_id;
ALTER TABLE addresses DROP COLUMN user_id;

CREATE INDEX idx_addresses_owner ON addresses(owner_type, owner_id);
