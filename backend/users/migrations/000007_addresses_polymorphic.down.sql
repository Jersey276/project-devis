-- Lossy: client-owned addresses cannot be reconstructed from user_id and are dropped.
DELETE FROM addresses WHERE owner_type = 'client';

ALTER TABLE addresses ADD COLUMN user_id TEXT;
UPDATE addresses SET user_id = owner_id WHERE owner_type = 'user';
ALTER TABLE addresses ALTER COLUMN user_id SET NOT NULL;
ALTER TABLE addresses ADD CONSTRAINT addresses_user_id_fkey
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE;

DROP INDEX idx_addresses_owner;
ALTER TABLE addresses DROP CONSTRAINT addresses_owner_type_check;
ALTER TABLE addresses DROP COLUMN owner_type;
ALTER TABLE addresses DROP COLUMN owner_id;

CREATE INDEX idx_addresses_user_id ON addresses(user_id);
