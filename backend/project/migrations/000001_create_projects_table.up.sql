CREATE TABLE projects (
    id          SERIAL    PRIMARY KEY,
    project_id  TEXT      NOT NULL UNIQUE,
    user_id     TEXT      NOT NULL,
    name        TEXT      NOT NULL,
    client_id   TEXT,
    status      TEXT      NOT NULL DEFAULT 'active',
    archived_at TIMESTAMP,
    created_at  TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_projects_user_id ON projects(user_id);
CREATE INDEX idx_projects_status  ON projects(status);
