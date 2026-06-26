-- Replace the global unique constraint on linked_user_id with a composite unique
-- on (linked_user_id, user_id) so a customer can be linked to multiple providers.
DROP INDEX IF EXISTS clients_linked_user_id_key;

CREATE UNIQUE INDEX clients_linked_user_provider_key
    ON clients (linked_user_id, user_id)
    WHERE linked_user_id IS NOT NULL;
