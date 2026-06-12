ALTER TABLE auth
    DROP CONSTRAINT IF EXISTS auth_session_version_check,
    DROP COLUMN IF EXISTS session_version;
