-- Reset the backfilled flags. Safe because is_eu defaults to FALSE and is only
-- ever set by this backfill in the current schema.
UPDATE countries SET is_eu = FALSE;
