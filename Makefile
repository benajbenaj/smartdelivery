COMPOSE_FILE := docker/docker-compose.yml
API_DIR := apps/api-go
DATABASE_URL ?= postgres://smart_delivery:smart_delivery@localhost:5432/smart_delivery?sslmode=disable

.PHONY: db-up db-down db-logs test test-integration migrate-up migrate-down migrate-status

db-up:
	docker compose -f $(COMPOSE_FILE) up -d postgres

db-down:
	docker compose -f $(COMPOSE_FILE) down

db-logs:
	docker compose -f $(COMPOSE_FILE) logs -f postgres

test:
	cd $(API_DIR) && go test ./...

test-integration: db-up
	cd $(API_DIR) && DATABASE_URL="$(DATABASE_URL)" go test -tags=integration ./...

migrate-up: db-up
	cd $(API_DIR) && DATABASE_URL="$(DATABASE_URL)" go run ./cmd/migrate up

migrate-down: db-up
	cd $(API_DIR) && DATABASE_URL="$(DATABASE_URL)" go run ./cmd/migrate down

migrate-status: db-up
	cd $(API_DIR) && DATABASE_URL="$(DATABASE_URL)" go run ./cmd/migrate status
