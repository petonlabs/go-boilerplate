# Makefile for common dev tasks (make from repo root)

.PHONY: help docker-up docker-down build backend-run migrations-up migrations-down backup-run test lint
.PHONY: check-env psql redis-cli logs open-docs list-backups psql-runner

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

# Prepare development environment file with safe defaults
check-env:
	@echo "Checking and populating developer env (apps/backend/.env) with safe defaults"
	./scripts/check-dev-env.sh --fix

# Convenience helpers for local development (safe defaults: run inside compose network)
psql:
	docker compose exec postgres psql -U app -d app

redis-cli:
	docker compose exec redis redis-cli ping

logs:
	docker compose logs --follow backend

open-docs:
	@echo "Opening OpenAPI UI at http://localhost:8080/docs"
	open "http://localhost:8080/docs"

list-backups:
	docker compose run --rm db-backup ls -la /var/backups || true

# Start a temporary psql client container that connects to the compose network
psql-runner:
	docker run --rm -it --network $(shell docker compose ps -q | xargs -I{} docker inspect --format '{{range $k,$v := .NetworkSettings.Networks}}{{printf "%s" $k}}{{end}}' {}) postgres:18-alpine psql -h postgres -U app -d app

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
