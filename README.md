# Go Boilerplate — Production-Ready, Dockerized, with Backups & dspy-go# Go Boilerplate — Production-Ready, Dockerized, with Backups & dspy-go# Go Boilerplate



This repository is production-ready for **Coolify** deployment with:



- 🐳 **Dockerized Go backend** (distroless, non-root, healthchecks)This repository is production-ready for **Coolify** deployment with:A production-ready monorepo template for building scalable web applications with Go backend and TypeScript frontend. Built with modern best practices, clean architecture, and comprehensive tooling.

- 🗄️ **PostgreSQL 18** + **Redis 8**

- 💾 **Automated S3 backups** to Cloudflare R2 (every 6 hours) with restore scripts

- 🔒 **Security-first**: distroless images, non-root user, healthchecks, restart policies

- 🤖 **dspy-go integration** with Azure AI (connectivity endpoint at `/dspy/health`)- 🐳 **Dockerized Go backend** (distroless, non-root, healthchecks)## Features

- 📦 **No prefix pollution**: Environment variables normalized (no `BOILERPLATE_` prefix)

- 🎯 **Coolify-compatible**: Optional Traefik labels, no hardcoded domains- 🗄️ **PostgreSQL 18** + **Redis 8**

- 🛠️ **Taskfile targets** for Docker, backups, and common workflows

- 💾 **Automated S3 backups** to Cloudflare R2 (every 6 hours) with restore scripts- **Monorepo Structure**: Organized with Turborepo for efficient builds and development

---

- 🔒 **Security-first**: distroless images, non-root user, healthchecks, restart policies- **Go Backend**: High-performance REST API with Echo framework

## 🚀 Quick Start

- 🤖 **dspy-go integration** with Azure AI (connectivity endpoint at `/dspy/health`)- **Authentication**: Integrated Clerk SDK for secure user management

### 1. Copy environment variables

- 📦 **No prefix pollution**: Environment variables normalized (no `BOILERPLATE_` prefix)- **Database**: PostgreSQL with migrations and connection pooling

```bash

cp apps/backend/.env.example apps/backend/.env- 🎯 **Coolify-compatible**: Optional Traefik labels, no hardcoded domains- **Background Jobs**: Redis-based async job processing with Asynq

```

- 🛠️ **Taskfile targets** for Docker, backups, and common workflows- **Observability**: New Relic APM integration and structured logging

Fill in the values (especially `DSPY_*` and `S3_*` credentials). **Never commit secrets.**

- **Email Service**: Transactional emails with Resend and HTML templates

### 2. Local Development

---- **Testing**: Comprehensive test infrastructure with Testcontainers

```bash

    "redis": { "response_time": "496.792µs", "status": "healthy" }
# Go Boilerplate

A production-oriented monorepo template with a Dockerized Go backend, PostgreSQL, Redis, automated backups to S3-compatible storage (Cloudflare R2), and optional DSPy (Azure AI) integration.

This README is organized into chapters to make onboarding easier. Each chapter is short and actionable — follow the Quick Start first, then dig into other sections as needed.

--

## Chapters

1. Quick start
2. Local development
3. Accessing services
4. Environment variables reference
5. Backups & restore
6. Tasks & tooling
7. Testing
8. Production notes
9. Contributing & license

---

## 1) Quick start (get running in ~5 minutes)

Prerequisites: Docker & Docker Compose, Go 1.24+, (optional) bun/npm for frontend.

1. Copy example env into `apps/backend/.env` and edit secrets (do not commit secrets):

```bash
cp apps/backend/.env.example apps/backend/.env
# or run the helper that appends safe dev defaults
make check-env
```

2. Start the stack:

```bash
make docker-up
# or: docker compose up -d --build
```

3. Verify the backend is healthy:

```bash
curl -sS http://localhost:8080/health | python3 -m json.tool
```

If you see `{"status":"healthy", ...}` you're good to go.

---

## 2) Local development

- Build backend binary (local):

```bash
make build
```

- Run backend locally (uses `apps/backend/.env`):

```bash
make backend-run
```

- Bring full dev environment up (DB, Redis, backup runner):

```bash
make docker-up
make logs   # follow backend logs
```

Notes:
- The repo intentionally does not expose Postgres/Redis on the host by default for safety. Use the `docker-compose.override.yml` (opt-in) to map ports to localhost if you need GUI tools.

---

## 3) Accessing services (practical recipes)

HTTP (backend)

```bash
# API base
http://localhost:8080

# Health
curl -sS http://localhost:8080/health | python3 -m json.tool

# Open API UI in macOS default browser
make open-docs
```

Postgres (recommended: use compose network)

```bash
# interactive psql inside container
make psql

# single command

```

Redis

```bash
make redis-cli
```

Backups (list files in the named volume)

```bash
make list-backups
```

Optional: expose DB/Redis on host by enabling the included `docker-compose.override.yml` and starting compose with the override.

---

## 4) Environment variables reference (high level)

All backend environment variables live in `apps/backend/.env` (no `BOILERPLATE_` prefix).

Important keys (see `apps/backend/.env.example` for full list):

- SERVER_PORT — HTTP port (default: 8080)
- PRIMARY_ENV — environment name (development|production)
- DATABASE_HOST, DATABASE_PORT, DATABASE_USER, DATABASE_PASSWORD, DATABASE_NAME
- REDIS_ADDRESS — redis host:port (used internally by the container)
- AUTH_SECRET_KEY — secret for signing tokens
- INTEGRATION_RESEND_API_KEY — resend email service key
- OBSERVABILITY_NEWRELIC_LICENSE_KEY — New Relic license key
- DSPY_* — DSPy (Azure AI) configuration (if enabled)
- S3_* — credentials for backup destination (Cloudflare R2)

Always keep secrets out of the repository and use environment-specific management for production (Coolify, Kubernetes Secrets, etc.).

---

## 5) Backups & restore (Cloudflare R2)

Backups are implemented via scripts in `scripts/backup` and run by `db-backup` service. By default backup cron is `0 */6 * * *` and retention is 14 days.

Manual backup (from repo root):

```bash
make backup-run
```

Restore example (uses `task` automation in backend Taskfile):

```bash
# Example (adapt URI):
---```
```

Backups are stored in a Docker named volume `backups` — use `make list-backups` to inspect.

---

## 6) Tasks & tooling

Root Makefile provides common developer targets:

- `make docker-up` / `make docker-down` — start/stop compose stack
- `make check-env` — populate safe dev defaults in `apps/backend/.env`
- `make build`, `make backend-run` — local build/run
- `make migrations-up` / `migrations-down` — run tern migrations
- `make psql`, `make redis-cli`, `make logs`, `make open-docs`, `make list-backups` — convenience helpers

Backend also has a Taskfile for more advanced automation; see `apps/backend/Taskfile.yml`.

---

## 7) Testing

Run unit tests:

```bash
cd apps/backend

```

Integration tests requiring Docker:

```bash
go test -tags=integration ./...
```

---

## 8) Production notes (concise)

- Use a secrets manager for S3 and DSPy credentials.
- Enable SSL for DB connections in production via `DATABASE_SSL_MODE=require`.
- Configure monitoring and alerting (New Relic, logs).
- Use a reverse proxy (Traefik/Caddy/nginx) for TLS termination and routing.

---

## 9) Contributing & license

- Fork, create a feature branch, run tests and linters, open a PR.
- This project is MIT licensed — see `LICENSE`.

---

This repository now exposes a set of chaptered documents under `docs/book/` which are suitable for exporting into a book-like tool (for example, WriteBook by ONCE).

See the full chapters in `docs/book/`:

- `docs/book/01-quick-start.md`
- `docs/book/02-local-development.md`
- `docs/book/03-accessing-services.md`
- `docs/book/04-env-reference.md`
- `docs/book/05-backups-restore.md`
- `docs/book/06-tasks-tooling.md`
- `docs/book/07-testing.md`
- `docs/book/08-production-notes.md`
- `docs/book/09-contributing-license.md`

Quick start (summary):

```bash
# prepare env
make check-env

# start stack
make docker-up

# confirm health
curl -sS http://localhost:8080/health | python3 -m json.tool
```

If you'd like, I can also produce a single export (e.g., a WriteBook-compatible bundle) or refine chapter frontmatter for a specific tool — tell me which target format you prefer.
# S3 Backups (Cloudflare R2)

S3_ENDPOINT=https://<account-id>.r2.cloudflarestorage.com

S3_BUCKET=boilerplate-og

S3_ACCESS_KEY_ID=<your-r2-access-key>## 📦 Coolify Deployment2. Install dependencies:

S3_SECRET_ACCESS_KEY=<your-r2-secret-key>

S3_REGION=auto```bash

S3_FORCE_PATH_STYLE=true

### Application Setup# Install frontend dependencies

# DSPY (Azure AI)

DSPY_ENABLED=truebun install

DSPY_PROVIDER=azure

DSPY_AZURE_ENDPOINT=https://<your-resource>.services.ai.azure.com/api/projects/<project>1. **Create a new Application** in Coolify

DSPY_AZURE_API_KEY=<your-azure-key>

DSPY_MODEL=gpt-4o-mini2. **Configure Build**:# Install backend dependencies

```

   - Dockerfile: `apps/backend/Dockerfile`cd apps/backend

4. **Optional Traefik Labels** (in `docker-compose.yml`, currently commented out for flexibility)

   - Build context: repository rootgo mod download

---

   - Exposed port: `8080````

## 💾 Automated Backups



### Overview

3. **Set Environment Variables** (via Coolify Secrets):3. Set up environment variables:

- **Cron schedule**: Every 6 hours (configurable via `BACKUP_CRON`)

- **Destination**: Cloudflare R2 (S3-compatible)```bash

- **Compression**: zstd

- **Retention**: 14 days locally (configurable via `BACKUP_RETENTION_DAYS`)```bashcp apps/backend/.env.example apps/backend/.env



### Manual Backup# Server# Edit apps/backend/.env with your configuration



```bashSERVER_PORT=8080```

task -C apps/backend backup:run

```PRIMARY_ENV=production



### Restore from Backup4. Start the database and Redis.



```bash# Database (use Coolify's managed Postgres or external)

task -C apps/backend backup:restore URI=s3://boilerplate-og/postgres/app/pg_app_YYYYMMDDTHHMMSSZ.sql.zst

```DATABASE_HOST=<your-postgres-host>5. Run database migrations:



**Important**: Ensure Cloudflare R2 credentials (`S3_*` vars) are set in your environment or Coolify Secrets.DATABASE_PORT=5432```bash



---DATABASE_USER=<your-db-user>cd apps/backend



## 🌍 Environment Variables ReferenceDATABASE_PASSWORD=<your-db-password>task migrations:up



All environment variables are **unprefixed** (no `BOILERPLATE_` prefix):DATABASE_NAME=<your-db-name>```



### ServerDATABASE_SSL_MODE=require

- `SERVER_PORT` — HTTP port (default: `8080`)

- `SERVER_READ_TIMEOUT` — Request read timeout in secondsDATABASE_MAX_OPEN_CONNS=256. Start the development server:

- `SERVER_WRITE_TIMEOUT` — Response write timeout in seconds

- `SERVER_IDLE_TIMEOUT` — Idle connection timeout in secondsDATABASE_MAX_IDLE_CONNS=5```bash

- `SERVER_CORS_ALLOWED_ORIGINS` — Comma-separated CORS origins

DATABASE_CONN_MAX_LIFETIME=300# From root directory

### Primary

- `PRIMARY_ENV` — Environment name (`development`, `production`, etc.)DATABASE_CONN_MAX_IDLE_TIME=60bun dev



### Database

- `DATABASE_HOST` — Postgres host

- `DATABASE_PORT` — Postgres port (default: `5432`)# Redis# Or just the backend

- `DATABASE_USER` — Database user

- `DATABASE_PASSWORD` — Database passwordREDIS_ADDRESS=<your-redis-host>:6379cd apps/backend

- `DATABASE_NAME` — Database name

- `DATABASE_SSL_MODE` — SSL mode (`disable`, `require`, etc.)task run

- `DATABASE_MAX_OPEN_CONNS` — Max open connections

- `DATABASE_MAX_IDLE_CONNS` — Max idle connections# Auth```

- `DATABASE_CONN_MAX_LIFETIME` — Connection max lifetime (seconds)

- `DATABASE_CONN_MAX_IDLE_TIME` — Connection max idle time (seconds)AUTH_SECRET_KEY=<generate-strong-key>



### RedisThe API will be available at `http://localhost:8080`

- `REDIS_ADDRESS` — Redis address (e.g., `localhost:6379`)

# Email (Resend)

### Auth

- `AUTH_SECRET_KEY` — Secret key for authenticationINTEGRATION_RESEND_API_KEY=<your-resend-key>## Development



### Email

- `INTEGRATION_RESEND_API_KEY` — Resend API key

# Observability (optional)### Available Commands

### Observability

- `OBSERVABILITY_NEWRELIC_LICENSE_KEY` — New Relic license keyOBSERVABILITY_NEWRELIC_LICENSE_KEY=<your-nr-key>

- `OBSERVABILITY_NEWRELIC_APP_NAME` — Application name in New Relic

OBSERVABILITY_NEWRELIC_APP_NAME=go-boilerplate```bash

### S3 (Cloudflare R2)

- `S3_ENDPOINT` — S3-compatible endpoint URL# Backend commands (from backend/ directory)

- `S3_BUCKET` — Bucket name

- `S3_ACCESS_KEY_ID` — Access key ID# S3 Backups (Cloudflare R2)task help              # Show all available tasks

- `S3_SECRET_ACCESS_KEY` — Secret access key

- `S3_REGION` — Region (use `auto` for R2)S3_ENDPOINT=https://<account-id>.r2.cloudflarestorage.comtask run               # Run the application

- `S3_FORCE_PATH_STYLE` — Use path-style URLs (`true` for R2)

S3_BUCKET=boilerplate-ogtask migrations:new    # Create a new migration

### Backups

- `BACKUP_CRON` — Cron schedule (default: `0 */6 * * *`)S3_ACCESS_KEY_ID=<your-r2-access-key>task migrations:up     # Apply migrations

- `BACKUP_RETENTION_DAYS` — Local retention period (default: `14`)

S3_SECRET_ACCESS_KEY=<your-r2-secret-key>task test              # Run tests

### DSPY (Azure AI)

- `DSPY_ENABLED` — Enable dspy-go (`true`/`false`)S3_REGION=autotask tidy              # Format code and manage dependencies

- `DSPY_PROVIDER` — Provider name (`azure`)

- `DSPY_AZURE_ENDPOINT` — Azure AI endpoint URLS3_FORCE_PATH_STYLE=true

- `DSPY_AZURE_API_KEY` — Azure AI API key

- `DSPY_MODEL` — Model name (e.g., `gpt-4o-mini`)# Frontend commands (from root directory)



---# DSPY (Azure AI)bun dev                # Start development servers



## 🤖 dspy-go IntegrationDSPY_ENABLED=truebun build              # Build all packages



This project integrates **[dspy-go](https://github.com/XiaoConstantine/dspy-go)** for Azure AI connectivity.DSPY_PROVIDER=azurebun lint               # Lint all packages



### Health EndpointDSPY_AZURE_ENDPOINT=https://go-boilerplate-resource.services.ai.azure.com/api/projects/go-boilerplate```



- **URL**: `GET /dspy/health`DSPY_AZURE_API_KEY=<your-azure-key>

- **Response** (healthy):

  ```jsonDSPY_MODEL=gpt-4o-mini### Environment Variables

  {

    "status": "ok"```

  }

  ```The backend uses environment variables prefixed with `BOILERPLATE_`. Key variables include:

- **Response** (unhealthy):

  ```json4. **Optional Traefik Labels** (in `docker-compose.yml`, currently commented):

  {

    "status": "unhealthy",   ```yaml- `BOILERPLATE_DATABASE_*` - PostgreSQL connection settings

    "error": "connection timeout"

  }   labels:- `BOILERPLATE_SERVER_*` - Server configuration

  ```

     - "traefik.enable=true"- `BOILERPLATE_AUTH_*` - Authentication settings

### Testing

     - "traefik.http.routers.backend.entrypoints=web"- `BOILERPLATE_REDIS_*` - Redis connection

```bash

cd apps/backend     - "traefik.http.services.backend.loadbalancer.server.port=8080"- `BOILERPLATE_EMAIL_*` - Email service configuration

DSPY_ENABLED=true DSPY_PROVIDER=azure DSPY_AZURE_ENDPOINT=<endpoint> DSPY_AZURE_API_KEY=<key> DSPY_MODEL=gpt-4o-mini go test ./internal/dspy

```   ```- `BOILERPLATE_OBSERVABILITY_*` - Monitoring settings



---



## 🛠️ Taskfile Commands---See `apps/backend/.env.example` for a complete list.



All commands are run from the `apps/backend` directory:



```bash## 💾 Automated Backups## Architecture

task help              # Show all available tasks

task run               # Run the application locally

task tidy              # Format code and tidy dependencies

task docker:build      # Build Docker image### OverviewThis boilerplate follows clean architecture principles:

task docker:up         # Start Docker stack

task docker:down       # Stop Docker stack

task backup:run        # Run manual backup

task backup:restore    # Restore from S3 (URI=s3://...)- **Cron schedule**: Every 6 hours (configurable via `BACKUP_CRON`)- **Handlers**: HTTP request/response handling

task migrations:new    # Create new migration (name=...)

task migrations:up     # Apply migrations- **Destination**: Cloudflare R2 (S3-compatible)- **Services**: Business logic implementation

```

- **Compression**: zstd- **Repositories**: Data access layer

---

- **Retention**: 14 days locally (configurable via `BACKUP_RETENTION_DAYS`)- **Models**: Domain entities

## 🔒 Security & Resilience

- **Infrastructure**: External services (database, cache, email)

- **Distroless base image** (`gcr.io/distroless/static:nonroot`)

- **Non-root user** (`nonroot:nonroot`)### Manual Backup

- **Health checks** built into Docker container and Compose

- **Restart policy**: `unless-stopped`## Testing

- **Secrets management**: Environment variables only (never committed)

- **SSL/TLS** for database connections in production```bash

- **Data volumes** for Postgres with automated backups to R2

task -C apps/backend backup:run```bash

---

```# Run backend tests

## 📂 Project Structure

cd apps/backend

```

.### Restore from Backupgo test ./...

├── apps/

│   ├── backend/          # Go backend application

│   │   ├── cmd/

│   │   │   └── go-boilerplate/  # Main entry point```bash# Run with coverage

│   │   ├── internal/

│   │   │   ├── config/   # Configuration loadingtask -C apps/backend backup:restore URI=s3://boilerplate-og/postgres/app/pg_app_YYYYMMDDTHHMMSSZ.sql.zstgo test -cover ./...

│   │   │   ├── database/ # Database & migrations

│   │   │   ├── dspy/     # dspy-go client```

│   │   │   ├── handler/  # HTTP handlers

│   │   │   ├── router/   # Route registration# Run integration tests (requires Docker)

│   │   │   └── ...

│   │   ├── Dockerfile    # Multi-stage Docker build**Important**: Ensure Cloudflare R2 credentials (`S3_*` vars) are set in your environment or Coolify Secrets.go test -tags=integration ./...

│   │   ├── Taskfile.yml  # Task automation

│   │   └── .env.example  # Environment template```

│   └── frontend/         # React frontend (Vite)

├── packages/             # Shared packages---

│   ├── emails/

│   ├── openapi/### Production Considerations

│   └── zod/

├── scripts/## 🌍 Environment Variables Reference

│   └── backup/           # Backup automation scripts

│       ├── entrypoint.sh1. Use environment-specific configuration

│       ├── run-backup.sh

│       └── restore.shAll environment variables are **unprefixed** (no `BOILERPLATE_` prefix):2. Enable production logging levels

├── docker-compose.yml    # Full stack orchestration

├── .dockerignore3. Configure proper database connection pooling

└── README.md             # This file

```### Server4. Set up monitoring and alerting



---- `SERVER_PORT` — HTTP port (default: `8080`)5. Use a reverse proxy (nginx, Caddy)



## 🧪 Testing- `SERVER_READ_TIMEOUT` — Request read timeout in seconds6. Enable rate limiting and security headers



### Run Go Tests- `SERVER_WRITE_TIMEOUT` — Response write timeout in seconds7. Configure CORS for your domains



```bash- `SERVER_IDLE_TIMEOUT` — Idle connection timeout in seconds

cd apps/backend

go test ./...- `SERVER_CORS_ALLOWED_ORIGINS` — Comma-separated CORS origins## Contributing

```



### Run with Coverage

### Primary1. Fork the repository

```bash

go test -coverprofile=coverage.out ./...- `PRIMARY_ENV` — Environment name (`development`, `production`, etc.)2. Create your feature branch (`git checkout -b feature/amazing-feature`)

go tool cover -html=coverage.out

```3. Commit your changes (`git commit -m 'Add amazing feature'`)



---### Database4. Push to the branch (`git push origin feature/amazing-feature`)



## 📝 License- `DATABASE_HOST` — Postgres host5. Open a Pull Request



See [LICENSE](LICENSE) for details.- `DATABASE_PORT` — Postgres port (default: `5432`)



---- `DATABASE_USER` — Database user## License



## 🙏 Acknowledgments- `DATABASE_PASSWORD` — Database password



- [dspy-go](https://github.com/XiaoConstantine/dspy-go) — DSPy framework for Go- `DATABASE_NAME` — Database nameThis project is licensed under the MIT License - see the LICENSE file for details.

- [Echo](https://echo.labstack.com/) — High-performance Go web framework

- [PostgreSQL](https://www.postgresql.org/) — Reliable relational database- `DATABASE_SSL_MODE` — SSL mode (`disable`, `require`, etc.)

- [Redis](https://redis.io/) — In-memory data structure store- `DATABASE_MAX_OPEN_CONNS` — Max open connections

- [Cloudflare R2](https://developers.cloudflare.com/r2/) — S3-compatible object storage- `DATABASE_MAX_IDLE_CONNS` — Max idle connections

- `DATABASE_CONN_MAX_LIFETIME` — Connection max lifetime (seconds)

---- `DATABASE_CONN_MAX_IDLE_TIME` — Connection max idle time (seconds)



**Ready to deploy? Just push to your Coolify instance and let the magic happen! 🚀**### Redis

- `REDIS_ADDRESS` — Redis address (e.g., `localhost:6379`)

### Auth
- `AUTH_SECRET_KEY` — Secret key for authentication

### Email
- `INTEGRATION_RESEND_API_KEY` — Resend API key

### Observability
- `OBSERVABILITY_NEWRELIC_LICENSE_KEY` — New Relic license key
- `OBSERVABILITY_NEWRELIC_APP_NAME` — Application name in New Relic

### S3 (Cloudflare R2)
- `S3_ENDPOINT` — S3-compatible endpoint URL
- `S3_BUCKET` — Bucket name
- `S3_ACCESS_KEY_ID` — Access key ID
- `S3_SECRET_ACCESS_KEY` — Secret access key
- `S3_REGION` — Region (use `auto` for R2)
- `S3_FORCE_PATH_STYLE` — Use path-style URLs (`true` for R2)

### Backups
- `BACKUP_CRON` — Cron schedule (default: `0 */6 * * *`)
- `BACKUP_RETENTION_DAYS` — Local retention period (default: `14`)

### DSPY (Azure AI)
- `DSPY_ENABLED` — Enable dspy-go (`true`/`false`)
- `DSPY_PROVIDER` — Provider name (`azure`)
- `DSPY_AZURE_ENDPOINT` — Azure AI endpoint URL
- `DSPY_AZURE_API_KEY` — Azure AI API key
- `DSPY_MODEL` — Model name (e.g., `gpt-4o-mini`)

---

## 🤖 dspy-go Integration

This project integrates **[dspy-go](https://github.com/XiaoConstantine/dspy-go)** for Azure AI connectivity.

### Health Endpoint

- **URL**: `GET /dspy/health`
- **Response** (healthy):
  ```json
  {
    "status": "ok"
  }
  ```
- **Response** (unhealthy):
  ```json
  {
    "status": "unhealthy",
    "error": "connection timeout"
  }
  ```

### Testing

```bash
cd apps/backend
DSPY_ENABLED=true DSPY_PROVIDER=azure DSPY_AZURE_ENDPOINT=<endpoint> DSPY_AZURE_API_KEY=<key> DSPY_MODEL=gpt-4o-mini go test ./internal/dspy
```

---

## 🛠️ Taskfile Commands

All commands are run from the `apps/backend` directory:

```bash
task help              # Show all available tasks
task run               # Run the application locally
task tidy              # Format code and tidy dependencies
task docker:build      # Build Docker image
task docker:up         # Start Docker stack
task docker:down       # Stop Docker stack
task backup:run        # Run manual backup
task backup:restore    # Restore from S3 (URI=s3://...)
task migrations:new    # Create new migration (name=...)
task migrations:up     # Apply migrations
```

---

## 🔒 Security & Resilience

- **Distroless base image** (`gcr.io/distroless/static:nonroot`)
- **Non-root user** (`nonroot:nonroot`)
- **Health checks** built into Docker container and Compose
- **Restart policy**: `unless-stopped`
- **Secrets management**: Environment variables only (never committed)
- **SSL/TLS** for database connections in production
- **Data volumes** for Postgres with automated backups to R2

---

## 📂 Project Structure

```
.
├── apps/
│   ├── backend/          # Go backend application
│   │   ├── cmd/
│   │   │   └── go-boilerplate/  # Main entry point
│   │   ├── internal/
│   │   │   ├── config/   # Configuration loading
│   │   │   ├── database/ # Database & migrations
│   │   │   ├── dspy/     # dspy-go client
│   │   │   ├── handler/  # HTTP handlers
│   │   │   ├── router/   # Route registration
│   │   │   └── ...
│   │   ├── Dockerfile    # Multi-stage Docker build
│   │   ├── Taskfile.yml  # Task automation
│   │   └── .env.example  # Environment template
│   └── frontend/         # React frontend (Vite)
├── packages/             # Shared packages
│   ├── emails/
│   ├── openapi/
│   └── zod/
├── scripts/
│   └── backup/           # Backup automation scripts
│       ├── entrypoint.sh
│       ├── run-backup.sh
│       └── restore.sh
├── docker-compose.yml    # Full stack orchestration
├── .dockerignore
└── README.md             # This file
```

---

## Local development

Use the provided helpers to start the full stack and developer servers.

- Makefile (root): common targets for Docker, migrations, tests and linting.
  - Start stack: `make docker-up`
  - Stop stack: `make docker-down`
  - Apply migrations: `make migrations-up`
  - Run backend locally: `make backend-run`
  - Run tests: `make test`

- Or use the orchestration script which brings up docker-compose, waits for services, runs migrations and starts backend and frontend dev servers:

```bash
# start everything (backend + DB + redis + frontend)
scripts/dev-start.sh

# skip the frontend
scripts/dev-start.sh --no-frontend

# run backend in foreground for debugging
scripts/dev-start.sh --foreground
```

Notes:

- The passwords included in `docker-compose.yml` are development defaults only. Replace them with proper secrets in production or CI.
- The script expects Docker to be installed. `tern` is used for migrations when available. The frontend path uses `bun` or `npm` if present.

---

## 🧪 Testing

### Run Go Tests

```bash
cd apps/backend
go test ./...
```

### Run with Coverage

```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

---

## � Documentation

Comprehensive documentation is available in the [`docs/`](./docs/) folder:

### 📖 Core Guides

| Document | Description | For |
|----------|-------------|-----|
| [**Best Practices**](./docs/BEST_PRACTICES.md) | Development workflow, debugging, and coding standards | All developers |
| [**CI Improvements**](./docs/CI_IMPROVEMENTS.md) | CI/CD pipeline setup and troubleshooting | DevOps, CI debugging |
| [**Module Configuration**](./docs/MODULE_CONFIGURATION.md) | Detailed dependency documentation | Architecture, updates |
| [**Dependency Audit**](./docs/DEPENDENCY_AUDIT.md) | Security audits and verification | Security, compliance |
| [**Linting Issues**](./docs/LINTING_ISSUES.md) | Historical fixes and lessons learned | Code review, learning |

### 🎯 Quick Links

- **New to the project?** Start with [Best Practices](./docs/BEST_PRACTICES.md)
- **CI failing?** Check [CI Improvements](./docs/CI_IMPROVEMENTS.md)
- **Need to debug?** See [Debugging section](./docs/BEST_PRACTICES.md#debugging-techniques)
- **Updating dependencies?** Read [Module Configuration](./docs/MODULE_CONFIGURATION.md)

### 📊 What's Documented

- ✅ Development workflow and pre-commit checks
- ✅ Code quality standards with examples
- ✅ Testing best practices (unit, integration, testcontainers)
- ✅ Security guidelines and vulnerability management
- ✅ Debugging techniques (local and CI)
- ✅ CI/CD pipeline architecture and optimization
- ✅ Complete dependency rationale and configuration
- ✅ Common issues and solutions
- ✅ Quick reference commands

---

## �📝 License

See [LICENSE](LICENSE) for details.

---

## 🙏 Acknowledgments

- [dspy-go](https://github.com/XiaoConstantine/dspy-go) — DSPy framework for Go
- [Echo](https://echo.labstack.com/) — High-performance Go web framework
- [PostgreSQL](https://www.postgresql.org/) — Reliable relational database
- [Redis](https://redis.io/) — In-memory data structure store
- [Cloudflare R2](https://developers.cloudflare.com/r2/) — S3-compatible object storage

---

**Ready to deploy? Just push to your Coolify instance and let the magic happen! 🚀**
