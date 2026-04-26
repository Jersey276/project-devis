#!/bin/bash
set -e

DB_PASSWORD=$(cat /run/secrets/db_password)

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
    CREATE USER "devis-users" WITH PASSWORD '$DB_PASSWORD';
    CREATE DATABASE users OWNER "devis-users";
EOSQL
