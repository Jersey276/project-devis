# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Repository structure

```
project-devis/
├── backend/           # Go microservices (each has its own go.mod)
│   ├── auth/          # Authentication service — gRPC :50051
│   ├── users/         # Users/clients/addresses/taxes — gRPC :50052
│   ├── quote/         # Quotes and quote lines — gRPC :50053
│   ├── export/        # PDF generation via Gotenberg — gRPC :50054
│   ├── template/      # Quote templates — gRPC :50055
│   ├── schedule/      # Payment schedules — gRPC :50056
│   ├── gateway/       # HTTP facade (Gin), proxies to gRPC services — :8080
│   ├── docker-compose.yml   # Local dev stack
│   └── Makefile
├── front/             # Next.js 16 app (App Router, TypeScript)
├── docs/              # Architecture docs, runbooks, error codes
└── docker-compose.prod.yml
```

## Commands

### Backend (run from `backend/`)

```bash
make up           # Start the full stack (detached)
make down         # Stop containers (preserves Postgres volume)
make rebuild      # Rebuild all Go services after code/migration changes
make rebuild-quote  # Rebuild a single service (auth|user|quote|schedule|gateway|export)
make migrate      # Rebuild only DB-backed services to apply pending migrations
make db-init      # Re-run init.sh to provision new databases (idempotent)
```

Run Go tests for a single service (from its directory, e.g. `backend/quote`):
```bash
go test ./...
```

### Frontend (run from `front/`)

```bash
npm run dev       # Start Next.js dev server on :3000
npm run build     # Production build
npm run lint      # ESLint
npm run cy:run -- --browser firefox   # Run Cypress e2e tests (always use firefox locally — Electron crashes)
npm run cy:open   # Open Cypress interactive runner
```

## Architecture

### Request flow

`Browser → Next.js (:3000) → [rewrites /api/* to] → Gateway (:8080) → gRPC services`

`next.config.ts` rewrites `/api/*` to `http://localhost:8080` in dev and `http://devis-gateway:8080` in production. All API calls from the front go through `front/lib/api.ts:apiFetch`, which handles automatic token refresh on 401 and session invalidation redirects.

### Backend pattern

Each Go service follows the same structure:
1. `main.go` — bootstraps DB connection, runs embedded migrations (`//go:embed migrations`), starts gRPC server
2. `actions/` — gRPC handler implementations
3. `actions/server.go` — `Server` struct holding DB and any downstream gRPC clients
4. `services/` — env vars, DB setup, migrate helpers, outbound gRPC client protos
5. `migrations/` — SQL migration files (golang-migrate, run at startup)

The gateway (`backend/gateway/`) is the only service that exposes HTTP. It authenticates via JWT middleware and maps gRPC business error codes to HTTP status codes per controller.

### Authorization

`backend/gateway/authz/` defines an `Authorizer` interface with `Can(subject, action, resource)`. The `Subject` carries `Role`, `AccountStatus`, and `SubscriptionTier`. Resources include `admin.countries`, `admin.country_groups`, `admin.taxes`, and `general`.

### Postgres

One shared cluster, one database per service. Provisioned by `backend/postgres/init.sh` (idempotent — safe to re-run with `make db-init`). The Docker volume `postgres:` persists across `compose down`; never pass `-v` unintentionally.

To fix a dirty migration: set `dirty=false` in `schema_migrations`, then `make rebuild-<service>`.

### Frontend structure

- `front/app/(app)/` — authenticated routes (sidebar layout)
- `front/app/(auth)/` — login/register/reset-password
- `front/components/` — feature components grouped by domain (`quote/`, `schedule/`, `template/`, `admin/`, `ui/`)
- `front/lib/api.ts` — central fetch wrapper with refresh/retry logic
- `front/messages/fr.json` — all i18n strings (French only)
- `front/i18n/request.ts` — next-intl locale resolver

The app uses next-intl for i18n with a single locale (`fr`). All user-visible strings must have keys in `front/messages/fr.json`.

### Cypress tests

Tests live in `front/cypress/e2e/`. The `cy.login()` command sets a fake auth cookie and stubs the critical API endpoints. Tests stub API responses with `cy.intercept` — no real backend needed. Always run locally with `--browser firefox`.

## Key conventions

### i18n
When extracting any string with `t("…")`, grep the entire file and confirm **every** key exists in `front/messages/fr.json` — not only newly added ones.

### Migrations
Never use `DELETE`/`TRUNCATE` to work around adding a NOT NULL column. Add the column nullable, backfill data, then add the constraint in a subsequent migration.

### Error codes
Business error codes flow as integers from gRPC services through the gateway to the frontend. See `docs/ERROR_CODES.md` and the per-service `actions/codes/codes.go` for the mapping. Frontend field-validation codes are mapped in `front/lib/api.ts:FIELD_VALIDATION_MESSAGES`.

### gRPC proto files
Generated `.pb.go` files live alongside the `.proto` files (e.g. `backend/quote/services/grpc/`). Regenerate with `protoc`; do not hand-edit generated files.
