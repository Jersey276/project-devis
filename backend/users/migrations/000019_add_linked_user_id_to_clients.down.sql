DROP INDEX IF EXISTS clients_linked_user_id_key;

ALTER TABLE clients DROP COLUMN IF EXISTS linked_user_id;
