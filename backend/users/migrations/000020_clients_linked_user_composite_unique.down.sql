DROP INDEX IF EXISTS clients_linked_user_provider_key;

CREATE UNIQUE INDEX clients_linked_user_id_key
    ON clients (linked_user_id)
    WHERE linked_user_id IS NOT NULL;
