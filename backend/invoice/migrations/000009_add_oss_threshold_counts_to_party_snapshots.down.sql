DROP INDEX IF EXISTS invoices_oss_cumulative_idx;

ALTER TABLE credit_note_party_snapshots
    DROP COLUMN IF EXISTS counts_toward_oss_threshold;

ALTER TABLE invoice_party_snapshots
    DROP COLUMN IF EXISTS counts_toward_oss_threshold;
