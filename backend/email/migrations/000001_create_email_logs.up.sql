CREATE TABLE IF NOT EXISTS email_logs (
    id             SERIAL PRIMARY KEY,
    user_id        TEXT,
    to_email       TEXT NOT NULL,
    type           TEXT NOT NULL,
    reference_name TEXT,
    subject        TEXT,
    status         TEXT NOT NULL DEFAULT 'sent',
    resend_id      TEXT,
    opened_at      TIMESTAMP,
    clicked_at     TIMESTAMP,
    created_at     TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_email_logs_user_id   ON email_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_email_logs_resend_id ON email_logs(resend_id);
