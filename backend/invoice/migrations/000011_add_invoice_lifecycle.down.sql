DROP TABLE IF EXISTS invoice_lifecycle_events;
ALTER TABLE invoices DROP COLUMN IF EXISTS lifecycle_status;
