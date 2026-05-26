# Backend

## Migrations workflow

Postgres data lives in the named volume `postgres:` (declared in `docker-compose.yml`). The volume **persists across `docker compose down` / `up`** — do not pass `-v` unless you intentionally want a full reset.

### Picking up new Go code or a new SQL migration

Edit your code (or add a migration under the service's `migrations/` folder, e.g. `quote/migrations/000007_xxx.up.sql` + `.down.sql`), then rebuild that service:

```
make rebuild-quote
```

This rebuilds the service image (binary + embedded migration files), restarts the container, and on startup `golang-migrate` applies any pending migrations against the live database. Existing data is preserved. `make rebuild` does all services at once.

### Adding a new database (a new service)

Edit `postgres/init.sh` to add a `create_user_and_db "devis-foo" "foo"` line, then:

```
make db-init
```

This re-runs `init.sh` against the live Postgres cluster via the `postgres-init` sidecar. The script is idempotent (`IF NOT EXISTS`), so existing databases are untouched.

### Recovering from a failed migration

`golang-migrate` marks the migration as `dirty` if it fails. To recover after fixing the SQL:

```
docker compose exec postgres psql -U devis-auth -d <db> -c \
  "UPDATE schema_migrations SET dirty=false WHERE version=<N>;"
make rebuild-<service>
```

No volume wipe needed.
