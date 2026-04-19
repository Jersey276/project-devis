CREATE TABLE users (
    id         SERIAL PRIMARY KEY,
    user_id    TEXT        NOT NULL UNIQUE,
    email      TEXT        NOT NULL UNIQUE,
    phone      TEXT,
    company    TEXT,
    siren      TEXT,
    vat        TEXT,
    created_at TIMESTAMP   NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP   NOT NULL DEFAULT NOW()
);
