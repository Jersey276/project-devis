ALTER TABLE credit_note_party_snapshots
    DROP COLUMN IF EXISTS oss_applied;

ALTER TABLE invoice_party_snapshots
    DROP COLUMN IF EXISTS oss_applied;
