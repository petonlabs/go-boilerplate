# Makefile for common dev tasks (make from repo root)

.PHONY: help docker-up docker-down build backend-run migrations-up migrations-down backup-run test lint

help:
	@echo "Make targets:"
	@echo "  docker-up        - start docker-compose stack (backend, postgres, redis, backup)"
	@echo "  docker-down      - stop docker-compose stack"
	@echo "  build            - build backend binary"
	@echo "  backend-run      - run backend (go run)"
	@echo "  migrations-up    - apply DB migrations (uses DB_DSN env)"
	@echo "  migrations-down  - rollback last migration (uses DB_DSN env)"
	@echo "  backup-run       - run DB backup job"
	@echo "  test             - run backend tests"
	@echo "  lint             - run golangci-lint (backend)"

# Docker shortcuts
docker-up:
	docker compose up -d

docker-down:
	docker compose down

# Build backend binary for local use
build:
	cd apps/backend && CGO_ENABLED=0 go build -o bin/server ./cmd/go-boilerplate

# Run backend locally (uses apps/backend/.env). Useful during development.
backend-run:
	cd apps/backend && go run ./cmd/go-boilerplate

# Migrations (require tern installed). Provide DB_DSN, otherwise falls back to postgres defaults.
migrations-up:
	@echo "Running migrations (DB_DSN=${DB_DSN:-postgresql://app:app@localhost:5432/app?sslmode=disable})"
	cd apps/backend && tern migrate -m ./internal/database/migrations --conn-string "${DB_DSN:-postgresql://app:app@localhost:5432/app?sslmode=disable}"

migrations-down:
	@echo "Rolling back last migration (DB_DSN=${DB_DSN:-postgresql://app:app@localhost:5432/app?sslmode=disable})"
	cd apps/backend && tern migrate -m ./internal/database/migrations --conn-string "${DB_DSN:-postgresql://app:app@localhost:5432/app?sslmode=disable}" down 1

# Run DB backup via docker-compose service
backup-run:
	docker compose run --rm db-backup /bin/sh /backup/run-backup.sh

# Tests
test:
	@echo "Running go tests"
	cd apps/backend && go test ./...

# Lint (requires golangci-lint installed locally)
lint:
	cd apps/backend && golangci-lint run ./...
