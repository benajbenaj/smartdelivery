COMPOSE_FILE := docker/docker-compose.yml
API_DIR := apps/api-go
DATABASE_URL ?= postgres://smart_delivery:smart_delivery@localhost:5432/smart_delivery?sslmode=disable

.PHONY: db-up db-down db-logs migrate-up migrate-down migrate-status

db-up:
	docker compose -f $(COMPOSE_FILE) up -d postgres

db-down:
	docker compose -f $(COMPOSE_FILE) down

db-logs:
	docker compose -f $(COMPOSE_FILE) logs -f postgres

migrate-up: db-up
	cd $(API_DIR) && DATABASE_URL="$(DATABASE_URL)" go run ./cmd/migrate up

migrate-down: db-up
	cd $(API_DIR) && DATABASE_URL="$(DATABASE_URL)" go run ./cmd/migrate down

migrate-status: db-up
	cd $(API_DIR) && DATABASE_URL="$(DATABASE_URL)" go run ./cmd/migrate status
