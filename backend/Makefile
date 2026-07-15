SERVICES := auth user quote schedule invoice gateway export template subscription email audit project
DB_SERVICES := auth user quote schedule invoice subscription email audit

# Maps short service names (used by cover-%) to their directory on disk, for
# the one case where they differ (user -> users/).
service-dir = $(if $(filter user,$(1)),users,$(1))

.PHONY: db-init migrate rebuild up down cover test-db-up test-db-down test-invoice $(addprefix rebuild-,$(SERVICES)) $(addprefix cover-,$(SERVICES))

# Re-run init.sh against the running cluster. Idempotent. Use after editing
# init.sh (e.g. adding a new database/role) — does NOT require wiping the volume.
db-init:
	docker compose run --rm postgres-init

# Rebuild every Go service so new code (including new //go:embed migration
# files) is baked into the binary, then restart. golang-migrate's m.Up() picks
# up any pending migrations on startup.
rebuild: $(addprefix rebuild-,$(SERVICES))

rebuild-%:
	docker compose up -d --build devis-$*

# Apply pending migrations: rebuild only the DB-backed services.
migrate: $(addprefix rebuild-,$(DB_SERVICES))

up:
	docker compose up -d

# Stops containers but PRESERVES the volume (and your data).
# Do NOT add -v to this target — that would delete the database.
down:
	docker compose down

# Disposable Postgres for integration tests that need a real DB (currently
# invoice/tests/*_integration_test.go, gated on INVOICE_TEST_DATABASE_URL).
# Each test creates its own schema and applies migrations into it, so this
# container is shared scratch space, not a per-service database.
TEST_DB_DSN := postgres://testadmin:testpass@localhost:55432/testdb?sslmode=disable

test-db-up:
	docker run -d --name devis-test-db \
		-e POSTGRES_USER=testadmin -e POSTGRES_PASSWORD=testpass -e POSTGRES_DB=testdb \
		-p 55432:5432 postgres:16-alpine
	@echo "waiting for postgres..."
	@until docker exec devis-test-db pg_isready -U testadmin >/dev/null 2>&1; do sleep 1; done
	@echo "test DB ready at $(TEST_DB_DSN)"

test-db-down:
	docker rm -f devis-test-db

# Run invoice's integration tests against test-db-up.
test-invoice:
	cd invoice && INVOICE_TEST_DATABASE_URL="$(TEST_DB_DSN)" go test ./...

# True test coverage per service, excluding generated protoc output
# (*.pb.go, *_grpc.pb.go) which is never hand-tested. Prints the overall
# percentage; use `make cover-<service> COVER_HTML=1` for an HTML report.
# For invoice, run test-db-up first (or export INVOICE_TEST_DATABASE_URL
# yourself) so the integration tests aren't silently skipped.
cover: $(addprefix cover-,$(SERVICES))

cover-%:
	@dir=$(call service-dir,$*); \
	cd $$dir && go test -coverpkg=./... -coverprofile=/tmp/devis-cover-$*.out ./... > /tmp/devis-cover-$*.log 2>&1; \
	status=$$?; \
	grep -v -E '(\.pb\.go|_grpc\.pb\.go):' /tmp/devis-cover-$*.out > /tmp/devis-cover-$*.filtered.out; \
	total=$$(go tool cover -func=/tmp/devis-cover-$*.filtered.out 2>/dev/null | tail -1 | awk '{print $$NF}'); \
	if [ $$status -ne 0 ]; then \
		echo "$* : BUILD/TEST FAILED (see /tmp/devis-cover-$*.log)"; \
	else \
		echo "$* : $$total"; \
	fi; \
	if [ -n "$$COVER_HTML" ]; then \
		go tool cover -html=/tmp/devis-cover-$*.filtered.out -o /tmp/devis-cover-$*.html; \
		echo "  -> /tmp/devis-cover-$*.html"; \
	fi
