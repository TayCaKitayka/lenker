MIGRATIONS_DIR ?= migrations
MIGRATE ?= migrate

.PHONY: migrate-up migrate-down migrate-force

migrate-up:
	@if [ -z "$$LENKER_DATABASE_URL" ]; then echo "LENKER_DATABASE_URL is required"; exit 1; fi
	$(MIGRATE) -path $(MIGRATIONS_DIR) -database "$$LENKER_DATABASE_URL" up

migrate-down:
	@if [ -z "$$LENKER_DATABASE_URL" ]; then echo "LENKER_DATABASE_URL is required"; exit 1; fi
	$(MIGRATE) -path $(MIGRATIONS_DIR) -database "$$LENKER_DATABASE_URL" down 1

migrate-force:
	@if [ -z "$$LENKER_DATABASE_URL" ]; then echo "LENKER_DATABASE_URL is required"; exit 1; fi
	@if [ -z "$$VERSION" ]; then echo "VERSION is required"; exit 1; fi
	$(MIGRATE) -path $(MIGRATIONS_DIR) -database "$$LENKER_DATABASE_URL" force "$$VERSION"
