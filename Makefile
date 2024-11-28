# Load environment variables from .env file
include .env
export

# Database migration commands
GOOSE := goose -dir internal/infrastructure/db/migrations

# SQLC command
SQLC := sqlc

# Docker Compose files
DOCKER_COMPOSE_FILE := docker-compose.yml
DOCKER_COMPOSE_KAFKA_FILE := docker-compose-kafka.yml

# Go commands
API_CMD := cmd/api/main.go
WORKER_CMD := cmd/worker/main.go

.PHONY: migrate-create migrate-up migrate-down migrate-status db-reset sqlc-generate sqlc-verify build run run-local run-worker-local test up down clean help

# Migration commands
migrate-create:
	@read -p "Enter migration name: " name; \
	$(GOOSE) -s create $$name sql

migrate-up:
	$(GOOSE) up

migrate-down:
	$(GOOSE) down

migrate-status:
	$(GOOSE) status

db-reset:
	$(GOOSE) reset

# SQLC commands
sqlc-generate:
	$(SQLC) generate

sqlc-verify:
	$(SQLC) verify

# Build commands
build:
	go build -o bin/api $(API_CMD)

run: build
	./bin/api

run-local:
	go run $(API_CMD)

run-worker-local:
	go run $(WORKER_CMD)

# Test commands
test:
	go test -v ./...

# Docker commands
up:
	docker-compose -f $(DOCKER_COMPOSE_FILE) up -d
	docker-compose -f $(DOCKER_COMPOSE_KAFKA_FILE) up -d

down:
	docker-compose -f $(DOCKER_COMPOSE_FILE) down
	docker-compose -f $(DOCKER_COMPOSE_KAFKA_FILE) down

clean:
	docker-compose -f $(DOCKER_COMPOSE_FILE) down -v
	docker-compose -f $(DOCKER_COMPOSE_KAFKA_FILE) down -v

# Help command
help:
	@echo "Available commands:"
	@echo "  migrate-create  - Create a new migration file"
	@echo "  migrate-up      - Run all available migrations"
	@echo "  migrate-down    - Rollback the last migration"
	@echo "  migrate-status  - Show current migration status"
	@echo "  db-reset        - Reset the database (rollback all migrations)"
	@echo "  sqlc-generate   - Generate Go code from SQL"
	@echo "  sqlc-verify     - Verify SQL queries against the schema"
	@echo "  build           - Build the application"
	@echo "  run             - Build and run the application"
	@echo "  run-local       - Run the API locally without building"
	@echo "  run-worker-local - Run the worker locally"
	@echo "  test            - Run tests"
	@echo "  up              - Start all Docker services"
	@echo "  down            - Stop all Docker services"
	@echo "  clean           - Stop all Docker services and remove volumes"