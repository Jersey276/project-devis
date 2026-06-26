CREATE TABLE quote_line_comments (
    comment_id  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    line_id     TEXT NOT NULL REFERENCES quote_lines(line_id) ON DELETE CASCADE,
    quote_id    TEXT NOT NULL,
    author_id   TEXT NOT NULL,
    author_name TEXT NOT NULL,
    body        TEXT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX ON quote_line_comments(line_id);
