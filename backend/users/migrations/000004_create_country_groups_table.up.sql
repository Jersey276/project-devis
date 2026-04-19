CREATE TABLE country_groups (
    id         SERIAL      PRIMARY KEY,
    name       TEXT        NOT NULL,
    created_at TIMESTAMP   NOT NULL DEFAULT NOW()
);

CREATE TABLE country_group_countries (
    country_group_id INTEGER NOT NULL REFERENCES country_groups(id) ON DELETE CASCADE,
    country_id       INTEGER NOT NULL REFERENCES countries(id)       ON DELETE CASCADE,
    PRIMARY KEY (country_group_id, country_id)
);
