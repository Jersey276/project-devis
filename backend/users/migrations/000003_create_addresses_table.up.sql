CREATE TABLE addresses (
    id                SERIAL      PRIMARY KEY,
    user_id           TEXT        NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    name              TEXT        NOT NULL,
    street            TEXT        NOT NULL,
    additional_street TEXT,
    city              TEXT        NOT NULL,
    zip_code          TEXT        NOT NULL,
    country_id        INTEGER     NOT NULL REFERENCES countries(id),
    email             TEXT,
    phone             TEXT,
    archived_at       TIMESTAMP,
    created_at        TIMESTAMP   NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMP   NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_addresses_user_id ON addresses(user_id);
