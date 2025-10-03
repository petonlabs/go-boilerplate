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

# Install dependencies- **API Documentation**: OpenAPI/Swagger specification

bun install

cd apps/backend && go mod download## 🚀 Quick Start- **Security**: Rate limiting, CORS, secure headers, and JWT validation



# Build and start the stack

cd ../..

task -C apps/backend docker:build### 1. Copy environment variables## Project Structure

task -C apps/backend docker:up

```



### 3. Verify Health```bash```



- **Backend**: <http://localhost:8080/health>cp apps/backend/.env.example apps/backend/.envgo-boilerplate/

- **dspy-go health**: <http://localhost:8080/dspy/health>

```├── apps/backend/          # Go backend application

---

├── packages/         # Frontend packages (React, Vue, etc.)

## 📦 Coolify Deployment

Fill in the values (especially `DSPY_*` and `S3_*` credentials). **Never commit secrets.**├── package.json      # Monorepo configuration

### Application Setup

├── turbo.json        # Turborepo configuration

1. **Create a new Application** in Coolify

2. **Configure Build**:### 2. Local Development└── README.md         # This file

   - Dockerfile: `apps/backend/Dockerfile`

   - Build context: repository root```

   - Exposed port: `8080`

```bash

3. **Set Environment Variables** (via Coolify Secrets):

# Install dependencies## Quick Start

```bash

# Serverbun install

SERVER_PORT=8080

PRIMARY_ENV=productioncd apps/backend && go mod download### Prerequisites



# Database (use Coolify's managed Postgres or external)

DATABASE_HOST=<your-postgres-host>

DATABASE_PORT=5432# Build and start the stack- Go 1.24 or higher

DATABASE_USER=<your-db-user>

DATABASE_PASSWORD=<your-db-password>cd ../..- Node.js 22+ and Bun

DATABASE_NAME=<your-db-name>

DATABASE_SSL_MODE=requiretask -C apps/backend docker:build- PostgreSQL 16+

DATABASE_MAX_OPEN_CONNS=25

DATABASE_MAX_IDLE_CONNS=5task -C apps/backend docker:up- Redis 8+

DATABASE_CONN_MAX_LIFETIME=300

DATABASE_CONN_MAX_IDLE_TIME=60```



# Redis### Installation

REDIS_ADDRESS=<your-redis-host>:6379

### 3. Verify Health

# Auth

AUTH_SECRET_KEY=<generate-strong-key>1. Clone the repository:



# Email (Resend)- **Backend**: http://localhost:8080/health```bash

INTEGRATION_RESEND_API_KEY=<your-resend-key>

- **dspy-go health**: http://localhost:8080/dspy/healthgit clone https://github.com/petonlabs/go-boilerplate.git

# Observability (optional)

OBSERVABILITY_NEWRELIC_LICENSE_KEY=<your-nr-key>cd go-boilerplate

OBSERVABILITY_NEWRELIC_APP_NAME=go-boilerplate

---```

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
