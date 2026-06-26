#!/bin/bash
set -e

DB_PASSWORD=$(cat /run/secrets/db_password)
# PGPASSWORD is consumed by psql when the sidecar connects over TCP. Harmless
# at first-boot (run via /docker-entrypoint-initdb.d/ via local socket).
export PGPASSWORD="$DB_PASSWORD"

create_user_and_db() {
    local user="$1"
    local db="$2"
    # Connect to the default "postgres" maintenance DB so this script works both
    # at first-boot (run via /docker-entrypoint-initdb.d/) and on a live cluster
    # (run via the postgres-init sidecar).
    psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "postgres" <<-EOSQL
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
create_user_and_db "devis-schedule" "schedule"
create_user_and_db "devis-invoice" "invoice"
create_user_and_db "devis-template" "template"
create_user_and_db "devis-subscription" "subscription"
create_user_and_db "devis-email" "email"
create_user_and_db "devis-audit" "audit"
create_user_and_db "devis-project" "project"

# Purge role for audit service: DELETE-only on activity_logs (no own database)
psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "postgres" <<-EOSQL
    DO \$\$
    BEGIN
        IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'devis-audit-purge') THEN
            CREATE ROLE "devis-audit-purge" LOGIN PASSWORD '$DB_PASSWORD';
        END IF;
    END
    \$\$;
EOSQL
