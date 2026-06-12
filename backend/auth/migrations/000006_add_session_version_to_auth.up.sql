ALTER TABLE auth
    ADD COLUMN session_version INTEGER NOT NULL DEFAULT 1;

ALTER TABLE auth
    ADD CONSTRAINT auth_session_version_check CHECK (session_version >= 1);
