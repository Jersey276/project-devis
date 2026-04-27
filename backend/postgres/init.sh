#!/bin/bash
set -e

DB_PASSWORD=$(cat /run/secrets/db_password)

create_user_and_db() {
    local user="$1"
    local db="$2"
    psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
        DO \$\$
        BEGIN
            IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = '$user') THEN
                CREATE ROLE "$user" LOGIN PASSWORD '$DB_PASSWORD';
            END IF;
        END
        \$\$;
        SELECT 'CREATE DATABASE "$db" OWNER "$user"'
        WHERE NOT EXISTS (SELECT 1 FROM pg_database WHERE datname = '$db')\gexec
EOSQL
}

create_user_and_db "devis-users" "users"
create_user_and_db "devis-quote" "quote"
