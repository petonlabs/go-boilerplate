# Go Boilerplate

A production-ready monorepo template for building scalable web applications with Go backend and TypeScript frontend. Built with modern best practices, clean architecture, and comprehensive tooling.

## ✨ Features

- 🐳 **Dockerized Go Backend**: Distroless images, non-root user, health checks
- 🗄️ **Database Stack**: PostgreSQL 18 + Redis 8 with connection pooling
- 💾 **Automated Backups**: S3-compatible backups to Cloudflare R2 every 6 hours
- 🔒 **Security First**: Distroless images, restart policies, secrets management
- 🤖 **AI Integration**: dspy-go with Azure AI support
- 📊 **Observability**: New Relic APM and structured logging
- ✉️ **Email Service**: Transactional emails with Resend
- 🔐 **Authentication**: Integrated Clerk SDK
- ⚡ **Background Jobs**: Redis-based async processing with Asynq
- 🧪 **Testing**: Comprehensive test infrastructure with Testcontainers
- 🎯 **Coolify Compatible**: Optional Traefik labels, environment-driven config

---

## 🚀 Quick Start

**Prerequisites**: Docker & Docker Compose, Go 1.24+, (optional) bun/npm for frontend.

### 1. Set up environment

```bash
cp apps/backend/.env.example apps/backend/.env
# Edit apps/backend/.env and fill in your credentials
# Never commit secrets!
```

### 2. Start the stack

```bash
make docker-up
```

### 3. Verify health

```bash
curl -sS http://localhost:8080/health | python3 -m json.tool
```

If you see `{"status":"healthy", ...}` you're ready to go! 🎉

---

## 🛠️ Common Commands

```bash
# Development
make build              # Build backend binary
make backend-run        # Run backend locally
make logs              # Follow container logs

# Database
make psql              # Access PostgreSQL shell
make migrations-up     # Apply migrations
make migrations-down   # Rollback migrations

# Testing
cd apps/backend
go test ./...          # Run unit tests
go test -tags=integration ./...  # Run integration tests

# Utilities
make redis-cli         # Access Redis CLI
make open-docs         # Open API documentation
make list-backups      # List backup files
```

For full command reference, see the [Makefile](./Makefile) or run `make help`.

---

## 📂 Project Structure

```
.
├── apps/
│   ├── backend/          # Go application
│   │   ├── cmd/          # Application entry points
│   │   ├── internal/     # Private application code
│   │   ├── Dockerfile    # Multi-stage build
│   │   └── Taskfile.yml  # Task automation
│   └── frontend/         # React/TypeScript app
├── docs/                 # Comprehensive documentation
│   ├── getting-started/  # Setup and onboarding
│   ├── development/      # Development guides
│   ├── operations/       # Deployment and CI/CD
│   └── reference/        # Technical references
├── packages/             # Shared packages
├── scripts/              # Utility scripts
└── docker-compose.yml    # Stack orchestration
```

---

## 📚 Documentation

Comprehensive documentation is organized by purpose in the [`docs/`](./docs/) folder:

### 🎯 Quick Navigation

| I want to... | Read this |
|--------------|-----------|
| Get started quickly | [Quick Start Guide](./docs/getting-started/QUICK_START.md) |
| Set up local development | [Local Development](./docs/getting-started/LOCAL_DEVELOPMENT.md) |
| Understand the architecture | [Architecture Guide](./docs/reference/ARCHITECTURE.md) |
| Deploy to production | [Production Deployment](./docs/operations/PRODUCTION.md) |
| Debug CI failures | [CI/CD Guide](./docs/operations/CI_CD.md) |
| Write quality code | [Best Practices](./docs/development/BEST_PRACTICES.md) |
| Run tests | [Testing Guide](./docs/development/TESTING.md) |
| Configure environment | [Configuration Reference](./docs/reference/CONFIGURATION.md) |

**Start here**: New to the project? Begin with the [Getting Started](./docs/getting-started/) guides.

---

## 🔧 Configuration

All configuration uses environment variables (no prefix pollution). Key variables:

```bash
# Server
SERVER_PORT=8080
PRIMARY_ENV=production

# Database
DATABASE_HOST=postgres
DATABASE_PORT=5432
DATABASE_USER=your_user
DATABASE_PASSWORD=your_password
DATABASE_NAME=your_db

# Redis
REDIS_ADDRESS=redis:6379

# Authentication
AUTH_SECRET_KEY=your_secret_key

# Email
INTEGRATION_RESEND_API_KEY=your_key

# Observability (optional)
OBSERVABILITY_NEWRELIC_LICENSE_KEY=your_key

# Backups (Cloudflare R2)
S3_ENDPOINT=https://account-id.r2.cloudflarestorage.com
S3_BUCKET=your-bucket
S3_ACCESS_KEY_ID=your_key
S3_SECRET_ACCESS_KEY=your_secret

# AI Integration (optional)
DSPY_ENABLED=true
DSPY_PROVIDER=azure
DSPY_AZURE_ENDPOINT=your_endpoint
DSPY_AZURE_API_KEY=your_key
```

See [`apps/backend/.env.example`](./apps/backend/.env.example) for the complete list.

---

## 🚢 Deployment

### Coolify

1. Create a new application in Coolify
2. Configure build settings:
   - Dockerfile: `apps/backend/Dockerfile`
   - Build context: repository root
   - Exposed port: `8080`
3. Set environment variables via Coolify Secrets
4. Optional: Enable Traefik labels in `docker-compose.yml`

### Docker Compose (Production)

```bash
# Start services
docker compose up -d

# View logs
docker compose logs -f backend

# Stop services
docker compose down
```

**Important**: Use proper secrets management in production (never commit secrets to git).

---

## 🔒 Security

- ✅ Distroless base images (minimal attack surface)
- ✅ Non-root user in containers
- ✅ Health checks and restart policies
- ✅ SSL/TLS for database connections
- ✅ Environment-based secrets management
- ✅ Regular dependency updates and security scans

For detailed security guidelines, see [Security Best Practices](./docs/development/BEST_PRACTICES.md#security).

---

## 🧪 Testing

```bash
cd apps/backend

# Unit tests
go test ./...

# With coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Integration tests (requires Docker)
go test -tags=integration ./...

# With race detection
go test -race ./...
```

See the [Testing Guide](./docs/development/TESTING.md) for more details.

---

## 🤝 Contributing

We welcome contributions! Please:

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests and linters
5. Submit a pull request

See [CONTRIBUTING.md](./docs/CONTRIBUTING.md) for detailed guidelines.

---

## 📝 License

This project is licensed under the MIT License. See [LICENSE](./LICENSE) for details.

---

## 🙏 Acknowledgments

- [dspy-go](https://github.com/XiaoConstantine/dspy-go) — DSPy framework for Go
- [Echo](https://echo.labstack.com/) — High-performance web framework
- [PostgreSQL](https://www.postgresql.org/) — Reliable relational database
- [Redis](https://redis.io/) — In-memory data store
- [Cloudflare R2](https://developers.cloudflare.com/r2/) — S3-compatible storage

---

**Ready to build something amazing? Let's get started! 🚀**
