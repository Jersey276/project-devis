ALTER TABLE clients ADD COLUMN linked_user_id TEXT;

CREATE UNIQUE INDEX clients_linked_user_id_key ON clients (linked_user_id) WHERE linked_user_id IS NOT NULL;
