# Quick Start Guide

Get the project running locally in ~5 minutes.

---

## Prerequisites

- Docker & Docker Compose
- Go 1.24+ (for local builds/tests)
- (Optional) bun/npm for frontend development

---

## Steps

### 1. Clone and Configure

```bash
# Clone the repository
git clone https://github.com/petonlabs/go-boilerplate.git
cd go-boilerplate

# Copy environment template
cp apps/backend/.env.example apps/backend/.env

# Or use the helper to append safe dev defaults
make check-env
```

**Important**: Edit `apps/backend/.env` and fill in your credentials. Never commit secrets to git!

### 2. Start the Stack

```bash
make docker-up
```

This will start:
- PostgreSQL 18 database
- Redis 8 cache
- Go backend application
- Automated backup service

### 3. Verify Health

```bash
curl -sS http://localhost:8080/health | python3 -m json.tool
```

Expected response:
```json
{
  "status": "healthy",
  "postgres": { "status": "healthy" },
  "redis": { "status": "healthy" }
}
```

If you see `"status": "healthy"`, you're ready to go! 🎉

---

## What's Next?

- **Develop locally**: See [Local Development](./LOCAL_DEVELOPMENT.md)
- **Access services**: Read [Accessing Services](./ACCESSING_SERVICES.md)
- **Configure environment**: Check [Configuration Reference](../reference/CONFIGURATION.md)
- **Run tests**: Follow the [Testing Guide](../development/TESTING.md)

---

## Troubleshooting

### Services won't start

```bash
# Check logs
make logs

# Restart services
make docker-down
make docker-up
```

### Database connection errors

Ensure your `.env` file has correct database credentials:
```bash
DATABASE_HOST=postgres
DATABASE_PORT=5432
DATABASE_USER=app
DATABASE_PASSWORD=app
DATABASE_NAME=app
```

### Port conflicts

If port 8080 is already in use, change `SERVER_PORT` in `.env`:
```bash
SERVER_PORT=8081
```

---

## Quick Commands Reference

```bash
make docker-up       # Start all services
make docker-down     # Stop all services
make logs           # View backend logs
make psql           # Access PostgreSQL
make redis-cli      # Access Redis
make open-docs      # Open API documentation
```

For more commands, see the [Makefile](../../Makefile) or run `make help`.
