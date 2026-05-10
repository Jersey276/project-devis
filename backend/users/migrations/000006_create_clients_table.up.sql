CREATE TABLE clients (
    id          SERIAL      PRIMARY KEY,
    client_id   TEXT        NOT NULL UNIQUE,
    user_id     TEXT        NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    first_name  TEXT        NOT NULL,
    last_name   TEXT        NOT NULL,
    email       TEXT,
    phone       TEXT,
    company     TEXT,
    siren       TEXT,
    vat         TEXT,
    archived_at TIMESTAMP,
    created_at  TIMESTAMP   NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMP   NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_clients_user_id ON clients(user_id);
