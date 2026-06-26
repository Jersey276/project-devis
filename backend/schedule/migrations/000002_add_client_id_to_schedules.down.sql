DROP INDEX IF EXISTS schedules_client_id_idx;
ALTER TABLE schedules DROP COLUMN IF EXISTS client_id;
