COMPOSE_FILE := docker/docker-compose.yml

.PHONY: db-up db-down db-logs

db-up:
	docker compose -f $(COMPOSE_FILE) up -d postgres

db-down:
	docker compose -f $(COMPOSE_FILE) down

db-logs:
	docker compose -f $(COMPOSE_FILE) logs -f postgres
