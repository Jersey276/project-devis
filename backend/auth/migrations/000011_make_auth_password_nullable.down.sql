-- Best-effort revert. This fails if any OAuth-only account (NULL password)
-- exists, which is expected — downs are only used in development.
ALTER TABLE auth ALTER COLUMN password SET NOT NULL;
