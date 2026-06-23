CREATE TABLE project_quotes (
    project_id TEXT      NOT NULL,
    quote_id   TEXT      NOT NULL UNIQUE,
    added_at   TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (project_id, quote_id)
);

CREATE INDEX idx_project_quotes_project_id ON project_quotes(project_id);
