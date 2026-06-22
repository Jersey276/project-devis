CREATE TABLE activity_logs (
    id          BIGSERIAL    PRIMARY KEY,
    user_id     TEXT,
    method      TEXT         NOT NULL,
    url         TEXT         NOT NULL,
    duration_ms INTEGER      NOT NULL,
    req_body    TEXT,
    resp_body   TEXT         NOT NULL,
    resp_status INTEGER      NOT NULL,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE INDEX idx_activity_logs_created_at  ON activity_logs (created_at);
CREATE INDEX idx_activity_logs_user_id     ON activity_logs (user_id);
CREATE INDEX idx_activity_logs_resp_status ON activity_logs (resp_status);

-- Applicative role: read + insert only, no update/delete (French regulatory requirement)
GRANT INSERT, SELECT ON activity_logs TO "devis-audit";
GRANT USAGE, SELECT ON SEQUENCE activity_logs_id_seq TO "devis-audit";
REVOKE UPDATE, DELETE ON activity_logs FROM "devis-audit";

-- Purge role: delete only (role created by init.sh, not here — requires superuser)
GRANT DELETE ON activity_logs TO "devis-audit-purge";
