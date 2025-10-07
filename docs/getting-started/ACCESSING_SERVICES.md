# Accessing Services

Practical commands for accessing HTTP API, PostgreSQL, Redis, and backups when the compose stack is running.

---

## HTTP API (Backend)

### Base URL

```bash
http://localhost:8080
```

### Health Check

```bash
# Basic check
curl http://localhost:8080/health

# Pretty JSON output
curl -sS http://localhost:8080/health | python3 -m json.tool

# Or use jq
curl -sS http://localhost:8080/health | jq
```

Expected response:
```json
{
  "status": "healthy",
  "postgres": {
    "status": "healthy",
    "response_time": "1.234ms"
  },
  "redis": {
    "status": "healthy",
    "response_time": "0.456ms"
  }
}
```

### API Documentation

```bash
# Open OpenAPI/Swagger UI (macOS)
make open-docs

# Or navigate manually
open http://localhost:8080/docs
```

### DSPy Health Endpoint (if enabled)

```bash
curl http://localhost:8080/dspy/health
```

---

## PostgreSQL

### Interactive Shell

```bash
# Quick access via Makefile
make psql

# This runs:
# docker compose exec postgres psql -U app -d app
```

### Run Single Queries

```bash
# Using docker compose
docker compose exec postgres psql -U app -d app -c "SELECT now();"

# List tables
docker compose exec postgres psql -U app -d app -c "\dt"

# Describe table
docker compose exec postgres psql -U app -d app -c "\d users"
```

### Common psql Commands

Once in the psql shell:

```sql
-- List databases
\l

-- List tables
\dt

-- Describe table structure
\d table_name

-- List indexes
\di

-- Show table size
\dt+

-- Execute SQL file
\i /path/to/file.sql

-- Quit
\q
```

### GUI Access (Optional)

To connect with tools like pgAdmin, DBeaver, or TablePlus:

1. Enable port mapping (see [Local Development](./LOCAL_DEVELOPMENT.md#using-a-gui-tool))
2. Connection details:
   - **Host**: localhost
   - **Port**: 5432
   - **Database**: app
   - **Username**: app
   - **Password**: app (from `.env`)

---

## Redis

### CLI Access

```bash
# Quick access via Makefile
make redis-cli

# This runs:
# docker compose exec redis redis-cli
```

### Common Redis Commands

Once in redis-cli:

```bash
# Test connection
PING
# Response: PONG

# Set a key
SET mykey "Hello World"

# Get a key
GET mykey

# List all keys (use carefully in production)
KEYS *

# Get key info
TYPE mykey
TTL mykey

# Delete a key
DEL mykey

# Flush all data (DANGER!)
FLUSHALL

# Exit
exit
```

### Run Single Commands

```bash
# Using docker compose
docker compose exec redis redis-cli PING

# Get a specific key
docker compose exec redis redis-cli GET session:abc123

# Check memory usage
docker compose exec redis redis-cli INFO memory
```

---

## Backups

### List Backups

```bash
# Quick access via Makefile
make list-backups

# This lists files in the backups volume
```

### Manual Backup

```bash
# Trigger a manual backup
make backup-run

# Or using task
cd apps/backend
task backup:run
```

### Restore from Backup

```bash
cd apps/backend

# Restore from S3 URI
task backup:restore URI=s3://boilerplate-og/postgres/app/pg_app_20250103T120000Z.sql.zst

# The script will:
# 1. Download from S3
# 2. Decompress zstd archive
# 3. Restore to database
```

**Important**: Ensure S3 credentials are configured in `.env` before restoring.

---

## Docker Services

### View Logs

```bash
# All services
docker compose logs -f

# Specific service
docker compose logs -f backend
docker compose logs -f postgres
docker compose logs -f redis

# Via Makefile
make logs  # Backend logs
```

### Service Status

```bash
# List running containers
docker compose ps

# Detailed status
docker compose ps -a
```

### Restart Services

```bash
# Restart all
docker compose restart

# Restart specific service
docker compose restart backend
docker compose restart postgres
```

### Execute Commands in Containers

```bash
# Backend container
docker compose exec backend /bin/sh

# Postgres container
docker compose exec postgres bash

# Redis container
docker compose exec redis sh
```

---

## Network Access

### From Host Machine

All services are accessible on `localhost`:
- Backend: `http://localhost:8080`
- PostgreSQL: `localhost:5432` (if exposed)
- Redis: `localhost:6379` (if exposed)

### From Container to Container

Services communicate via Docker network using service names:
- Backend → Postgres: `postgres:5432`
- Backend → Redis: `redis:6379`

Example connection string in `.env`:
```bash
DATABASE_HOST=postgres  # Not 'localhost'
REDIS_ADDRESS=redis:6379  # Not 'localhost:6379'
```

---

## Makefile Quick Reference

All convenience commands from the root Makefile:

```bash
make docker-up       # Start all services
make docker-down     # Stop all services
make logs           # View backend logs
make psql           # PostgreSQL shell
make redis-cli      # Redis CLI
make open-docs      # API documentation
make list-backups   # List backup files
make backup-run     # Manual backup
make build          # Build backend
make backend-run    # Run backend locally
make migrations-up  # Apply migrations
make migrations-down # Rollback migrations
make test           # Run tests
```

---

## Troubleshooting

### Can't connect to database

1. Check if containers are running:
   ```bash
   docker compose ps
   ```

2. Check database logs:
   ```bash
   docker compose logs postgres
   ```

3. Verify credentials in `.env`

### Redis connection issues

1. Test Redis is running:
   ```bash
   docker compose exec redis redis-cli PING
   ```

2. Check Redis address in `.env`:
   ```bash
   REDIS_ADDRESS=redis:6379  # Inside Docker network
   ```

### Backend not responding

1. Check health endpoint:
   ```bash
   curl http://localhost:8080/health
   ```

2. View logs:
   ```bash
   make logs
   ```

3. Restart backend:
   ```bash
   docker compose restart backend
   ```

---

## What's Next?

- **Local development workflow**: [Local Development](./LOCAL_DEVELOPMENT.md)
- **Testing your changes**: [Testing Guide](../development/TESTING.md)
- **Configuration reference**: [Configuration](../reference/CONFIGURATION.md)
