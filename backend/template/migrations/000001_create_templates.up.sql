CREATE TABLE IF NOT EXISTS templates (
    id          BIGSERIAL PRIMARY KEY,
    template_id UUID        NOT NULL DEFAULT gen_random_uuid() UNIQUE,
    user_id     TEXT        NOT NULL,
    template_type TEXT      NOT NULL,
    target_resource TEXT    NOT NULL,
    name        TEXT        NOT NULL,
    archived_at TIMESTAMPTZ,
    payload_version INT     NOT NULL DEFAULT 1,
    payload     JSONB       NOT NULL DEFAULT '{}',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_templates_user_id        ON templates (user_id);
CREATE INDEX IF NOT EXISTS idx_templates_template_type  ON templates (template_type);
CREATE INDEX IF NOT EXISTS idx_templates_target_resource ON templates (target_resource);
