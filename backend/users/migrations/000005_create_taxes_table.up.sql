CREATE TABLE taxes (
    id               SERIAL         PRIMARY KEY,
    name             TEXT           NOT NULL,
    rate             DECIMAL(5, 2)  NOT NULL,
    country_group_id INTEGER        NOT NULL REFERENCES country_groups(id) ON DELETE CASCADE,
    created_at       TIMESTAMP      NOT NULL DEFAULT NOW()
);
