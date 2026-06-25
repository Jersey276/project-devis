ALTER TABLE schedules ADD COLUMN client_id TEXT;
CREATE INDEX schedules_client_id_idx ON schedules (client_id) WHERE client_id IS NOT NULL;
