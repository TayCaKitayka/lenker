MIGRATIONS_DIR ?= migrations
MIGRATE ?= migrate
PANEL_API_DIR ?= services/panel-api
NODE_AGENT_DIR ?= services/node-agent
OPENAPI_SPEC ?= docs/openapi/panel-api.v1.yaml

.PHONY: migrate-up migrate-down migrate-force bootstrap-admin run-panel-api run-node-agent test-panel-api test-node-agent openapi-lint validate-openapi test

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

bootstrap-admin:
	@if [ -z "$$LENKER_DATABASE_URL" ]; then echo "LENKER_DATABASE_URL is required"; exit 1; fi
	@if [ -z "$$ADMIN_EMAIL" ]; then echo "ADMIN_EMAIL is required"; exit 1; fi
	@if [ -z "$$ADMIN_PASSWORD" ]; then echo "ADMIN_PASSWORD is required"; exit 1; fi
	cd $(PANEL_API_DIR) && go run ./cmd/bootstrap-admin -email "$$ADMIN_EMAIL" -password "$$ADMIN_PASSWORD"

run-panel-api:
	cd $(PANEL_API_DIR) && go run ./cmd/panel-api

run-node-agent:
	cd $(NODE_AGENT_DIR) && go run ./cmd/node-agent

test-panel-api:
	cd $(PANEL_API_DIR) && go test ./...

test-node-agent:
	cd $(NODE_AGENT_DIR) && go test ./...

openapi-lint validate-openapi:
	@if command -v ruby >/dev/null 2>&1; then ruby scripts/validate-openapi.rb $(OPENAPI_SPEC); else echo "ruby not found; skipping OpenAPI validation"; fi

test: test-panel-api test-node-agent openapi-lint
