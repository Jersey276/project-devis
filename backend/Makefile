SERVICES := auth user quote schedule invoice gateway export template subscription email
DB_SERVICES := auth user quote schedule invoice subscription email

.PHONY: db-init migrate rebuild up down $(addprefix rebuild-,$(SERVICES))

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
