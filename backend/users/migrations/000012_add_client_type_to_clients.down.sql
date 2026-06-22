ALTER TABLE clients
  DROP CONSTRAINT IF EXISTS clients_client_type_check;

ALTER TABLE clients
  DROP COLUMN IF EXISTS client_type;
