# Local Development Guide

How to build, run, and iterate on the backend locally.

---

## Development Workflow

### Option 1: Local Binary (Recommended for Development)

Build and run the backend binary directly on your machine:

```bash
# Build the binary
make build

# Run the backend (uses apps/backend/.env)
make backend-run
```

**Advantages**:
- Faster iteration
- Direct debugging support
- Native performance

### Option 2: Docker-based Development

Run everything in containers:

```bash
# Start all services (DB, Redis, backend, backups)
make docker-up

# Follow backend logs
make logs

# Stop all services
make docker-down
```

**Advantages**:
- Consistent environment
- Mirrors production setup
- Isolated from host system

---

## Common Development Tasks

### Making Code Changes

1. Edit code in `apps/backend/internal/`
2. For local binary: `make backend-run` (auto-reloads)
3. For Docker: `make docker-up --build`

### Database Migrations

```bash
# Create a new migration
cd apps/backend
task migrations:new name=add_users_table

# Apply migrations
make migrations-up

# Rollback migrations
make migrations-down
```

### Running Tests

```bash
cd apps/backend

# Unit tests
go test ./...

# Integration tests (requires Docker)
go test -tags=integration ./...

# With coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Code Formatting and Linting

```bash
# Format code
cd apps/backend
go fmt ./...

# Run linter
golangci-lint run ./...

# Or use the script
../../scripts/run-lint.sh
```

### Dependency Management

```bash
cd apps/backend

# Add a new dependency
go get github.com/some/package

# Update dependencies
go get -u ./...

# Tidy and verify
go mod tidy
go mod verify
```

---

## Frontend Development

```bash
# Install dependencies
bun install

# Start dev server (from root)
bun dev

# Or just backend
cd apps/backend
task run

# Build for production
bun build
```

---

## Database Access

### PostgreSQL Shell

```bash
# Interactive psql
make psql

# Run a query
docker compose exec postgres psql -U app -d app -c "SELECT * FROM users;"
```

### Using a GUI Tool

By default, PostgreSQL and Redis are not exposed on host ports for security. To enable:

1. Copy `docker-compose.override.yml.example` to `docker-compose.override.yml`
2. Uncomment port mappings:
   ```yaml
   services:
     postgres:
       ports:
         - "5432:5432"
     redis:
       ports:
         - "6379:6379"
   ```
3. Restart: `make docker-down && make docker-up`

Connect with your favorite tool:
- **Host**: localhost
- **Port**: 5432
- **User**: app
- **Password**: app (from `.env`)
- **Database**: app

---

## Redis Access

```bash
# Redis CLI
make redis-cli

# Or direct access
docker compose exec redis redis-cli

# Test connection
> PING
PONG
```

---

## Environment Variables

All configuration is in `apps/backend/.env`. Key variables for development:

```bash
# Server
SERVER_PORT=8080
PRIMARY_ENV=development

# Database
DATABASE_HOST=postgres
DATABASE_PORT=5432
DATABASE_USER=app
DATABASE_PASSWORD=app
DATABASE_NAME=app

# Redis
REDIS_ADDRESS=redis:6379

# Enable debug logging
LOG_LEVEL=debug
```

See [Configuration Reference](../reference/CONFIGURATION.md) for complete list.

---

## Debugging

### Using Delve (Go Debugger)

```bash
cd apps/backend

# Debug the application
dlv debug ./cmd/go-boilerplate

# Debug specific test
dlv test ./internal/handler -- -test.run TestHealthCheck
```

### Debugging with VS Code

Add to `.vscode/launch.json`:
```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Debug Backend",
      "type": "go",
      "request": "launch",
      "mode": "debug",
      "program": "${workspaceFolder}/apps/backend/cmd/go-boilerplate",
      "env": {},
      "envFile": "${workspaceFolder}/apps/backend/.env"
    }
  ]
}
```

---

## Hot Reload

For automatic recompilation on file changes:

```bash
# Install air
go install github.com/cosmtrek/air@latest

# Run with hot reload
cd apps/backend
air
```

---

## Development Scripts

Helpful scripts in `scripts/`:

```bash
# Test CI locally (simulates GitHub Actions)
./scripts/test-ci-locally.sh

# Run linter
./scripts/run-lint.sh

# Start full dev environment
./scripts/dev-start.sh

# Skip frontend
./scripts/dev-start.sh --no-frontend

# Run in foreground for debugging
./scripts/dev-start.sh --foreground
```

---

## Tips & Best Practices

1. **Keep .env updated**: When adding new config, update `.env.example`
2. **Run tests before committing**: Use pre-commit hooks
3. **Check logs regularly**: `make logs` helps catch issues early
4. **Use Makefile**: Standardized commands across team
5. **Read error messages**: Go's errors are descriptive

---

## What's Next?

- **Learn best practices**: [Best Practices](../development/BEST_PRACTICES.md)
- **Write tests**: [Testing Guide](../development/TESTING.md)
- **Access services**: [Accessing Services](./ACCESSING_SERVICES.md)
- **Deploy**: [Production Deployment](../operations/PRODUCTION.md)
