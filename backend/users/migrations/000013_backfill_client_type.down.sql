-- Irreversible data backfill: we cannot reliably tell backfilled rows from rows
-- set deliberately, so the down migration is a no-op. Dropping the column (down
-- of 000012) removes the data entirely if a full rollback is required.
SELECT 1;
