# Configuration Reference

Complete reference for all environment variables and configuration options.

---

## Overview

All configuration is managed through environment variables in `apps/backend/.env`. No prefixes are used for cleaner configuration management.

---

## Server Configuration

### `SERVER_PORT`
- **Type**: Integer
- **Default**: `8080`
- **Description**: HTTP port the server listens on
- **Example**: `SERVER_PORT=8080`

### `SERVER_READ_TIMEOUT`
- **Type**: Integer (seconds)
- **Default**: `30`
- **Description**: Maximum duration for reading the entire request
- **Example**: `SERVER_READ_TIMEOUT=30`

### `SERVER_WRITE_TIMEOUT`
- **Type**: Integer (seconds)
- **Default**: `30`
- **Description**: Maximum duration before timing out writes of the response
- **Example**: `SERVER_WRITE_TIMEOUT=30`

### `SERVER_IDLE_TIMEOUT`
- **Type**: Integer (seconds)
- **Default**: `120`
- **Description**: Maximum amount of time to wait for the next request
- **Example**: `SERVER_IDLE_TIMEOUT=120`

### `SERVER_CORS_ALLOWED_ORIGINS`
- **Type**: Comma-separated string
- **Default**: `*`
- **Description**: CORS allowed origins
- **Example**: `SERVER_CORS_ALLOWED_ORIGINS=https://app.example.com,https://admin.example.com`
- **Production**: Always set specific origins, never use `*`

---

## Environment Configuration

### `PRIMARY_ENV`
- **Type**: String
- **Default**: `development`
- **Values**: `development`, `staging`, `production`
- **Description**: Current environment name
- **Example**: `PRIMARY_ENV=production`

---

## Database Configuration

### `DATABASE_HOST`
- **Type**: String
- **Default**: `postgres`
- **Description**: PostgreSQL host (use service name in Docker)
- **Example**: `DATABASE_HOST=postgres`

### `DATABASE_PORT`
- **Type**: Integer
- **Default**: `5432`
- **Description**: PostgreSQL port
- **Example**: `DATABASE_PORT=5432`

### `DATABASE_USER`
- **Type**: String
- **Required**: Yes
- **Description**: PostgreSQL username
- **Example**: `DATABASE_USER=app`

### `DATABASE_PASSWORD`
- **Type**: String
- **Required**: Yes
- **Description**: PostgreSQL password
- **Security**: Never commit this value
- **Example**: `DATABASE_PASSWORD=your_secure_password`

### `DATABASE_NAME`
- **Type**: String
- **Required**: Yes
- **Description**: PostgreSQL database name
- **Example**: `DATABASE_NAME=app`

### `DATABASE_SSL_MODE`
- **Type**: String
- **Default**: `disable`
- **Values**: `disable`, `require`, `verify-ca`, `verify-full`
- **Description**: SSL mode for database connections
- **Example**: `DATABASE_SSL_MODE=require`
- **Production**: Always use `require` or higher

### `DATABASE_MAX_OPEN_CONNS`
- **Type**: Integer
- **Default**: `25`
- **Description**: Maximum number of open connections to the database
- **Example**: `DATABASE_MAX_OPEN_CONNS=25`

### `DATABASE_MAX_IDLE_CONNS`
- **Type**: Integer
- **Default**: `5`
- **Description**: Maximum number of idle connections
- **Example**: `DATABASE_MAX_IDLE_CONNS=5`

### `DATABASE_CONN_MAX_LIFETIME`
- **Type**: Integer (seconds)
- **Default**: `300`
- **Description**: Maximum amount of time a connection may be reused
- **Example**: `DATABASE_CONN_MAX_LIFETIME=300`

### `DATABASE_CONN_MAX_IDLE_TIME`
- **Type**: Integer (seconds)
- **Default**: `60`
- **Description**: Maximum amount of time a connection may be idle
- **Example**: `DATABASE_CONN_MAX_IDLE_TIME=60`

---

## Redis Configuration

### `REDIS_ADDRESS`
- **Type**: String
- **Default**: `redis:6379`
- **Description**: Redis server address (host:port)
- **Example**: `REDIS_ADDRESS=redis:6379`

### `REDIS_PASSWORD`
- **Type**: String
- **Default**: (empty)
- **Description**: Redis password (if authentication enabled)
- **Example**: `REDIS_PASSWORD=your_redis_password`

### `REDIS_DB`
- **Type**: Integer
- **Default**: `0`
- **Description**: Redis database number
- **Example**: `REDIS_DB=0`

---

## Authentication Configuration

### `AUTH_SECRET_KEY`
- **Type**: String
- **Required**: Yes
- **Description**: Secret key for signing authentication tokens
- **Security**: Use a strong random string, never commit
- **Example**: `AUTH_SECRET_KEY=your_32_char_secret_key_here`
- **Generation**: `openssl rand -base64 32`

### `AUTH_TOKEN_EXPIRY`
- **Type**: Integer (hours)
- **Default**: `24`
- **Description**: JWT token expiry time
- **Example**: `AUTH_TOKEN_EXPIRY=24`

### `AUTH_PASSWORD_RESET_TTL`
- **Type**: Integer (seconds)
- **Default**: `3600`
- **Description**: Password reset token time-to-live
- **Example**: `AUTH_PASSWORD_RESET_TTL=3600`

### `AUTH_DELETION_DEFAULT_TTL`
- **Type**: Integer (seconds)
- **Default**: `2592000` (30 days)
- **Description**: Default time before account deletion
- **Example**: `AUTH_DELETION_DEFAULT_TTL=2592000`

### `AUTH_WEBHOOK_SIGNING_SECRET`
- **Type**: String
- **Description**: Secret for verifying Clerk webhook signatures
- **Example**: `AUTH_WEBHOOK_SIGNING_SECRET=whsec_...`

---

## Email Configuration (Resend)

### `INTEGRATION_RESEND_API_KEY`
- **Type**: String
- **Required**: Yes (if email features enabled)
- **Description**: Resend API key for sending emails
- **Example**: `INTEGRATION_RESEND_API_KEY=re_...`

### `INTEGRATION_RESEND_FROM_EMAIL`
- **Type**: String
- **Default**: `noreply@example.com`
- **Description**: Default sender email address
- **Example**: `INTEGRATION_RESEND_FROM_EMAIL=noreply@yourapp.com`

---

## Observability Configuration

### `OBSERVABILITY_NEWRELIC_LICENSE_KEY`
- **Type**: String
- **Description**: New Relic license key for APM
- **Example**: `OBSERVABILITY_NEWRELIC_LICENSE_KEY=...`

### `OBSERVABILITY_NEWRELIC_APP_NAME`
- **Type**: String
- **Default**: `go-boilerplate`
- **Description**: Application name in New Relic
- **Example**: `OBSERVABILITY_NEWRELIC_APP_NAME=my-app-production`

### `LOG_LEVEL`
- **Type**: String
- **Default**: `info`
- **Values**: `debug`, `info`, `warn`, `error`
- **Description**: Logging level
- **Example**: `LOG_LEVEL=debug`

---

## S3/Backup Configuration (Cloudflare R2)

### `S3_ENDPOINT`
- **Type**: String
- **Description**: S3-compatible endpoint URL
- **Example**: `S3_ENDPOINT=https://account-id.r2.cloudflarestorage.com`

### `S3_BUCKET`
- **Type**: String
- **Description**: S3 bucket name
- **Example**: `S3_BUCKET=my-app-backups`

### `S3_ACCESS_KEY_ID`
- **Type**: String
- **Description**: S3 access key ID
- **Example**: `S3_ACCESS_KEY_ID=your_access_key`

### `S3_SECRET_ACCESS_KEY`
- **Type**: String
- **Description**: S3 secret access key
- **Security**: Never commit this value
- **Example**: `S3_SECRET_ACCESS_KEY=your_secret_key`

### `S3_REGION`
- **Type**: String
- **Default**: `auto`
- **Description**: S3 region (use 'auto' for Cloudflare R2)
- **Example**: `S3_REGION=auto`

### `S3_FORCE_PATH_STYLE`
- **Type**: Boolean
- **Default**: `true`
- **Description**: Use path-style URLs (required for R2)
- **Example**: `S3_FORCE_PATH_STYLE=true`

### `BACKUP_CRON`
- **Type**: Cron expression
- **Default**: `0 */6 * * *` (every 6 hours)
- **Description**: Backup schedule
- **Example**: `BACKUP_CRON=0 0 * * *` (daily at midnight)

### `BACKUP_RETENTION_DAYS`
- **Type**: Integer
- **Default**: `14`
- **Description**: Number of days to retain local backups
- **Example**: `BACKUP_RETENTION_DAYS=30`

---

## DSPy/AI Configuration (Optional)

### `DSPY_ENABLED`
- **Type**: Boolean
- **Default**: `false`
- **Description**: Enable DSPy AI integration
- **Example**: `DSPY_ENABLED=true`

### `DSPY_PROVIDER`
- **Type**: String
- **Default**: `azure`
- **Values**: `azure`, `openai`
- **Description**: AI provider
- **Example**: `DSPY_PROVIDER=azure`

### `DSPY_AZURE_ENDPOINT`
- **Type**: String
- **Description**: Azure AI endpoint URL
- **Example**: `DSPY_AZURE_ENDPOINT=https://your-resource.services.ai.azure.com/api/projects/your-project`

### `DSPY_AZURE_API_KEY`
- **Type**: String
- **Description**: Azure AI API key
- **Security**: Never commit this value
- **Example**: `DSPY_AZURE_API_KEY=your_api_key`

### `DSPY_MODEL`
- **Type**: String
- **Default**: `gpt-4o-mini`
- **Description**: AI model to use
- **Example**: `DSPY_MODEL=gpt-4o`

---

## Example Configurations

### Development

```bash
# apps/backend/.env (development)
PRIMARY_ENV=development
SERVER_PORT=8080
DATABASE_HOST=postgres
DATABASE_PORT=5432
DATABASE_USER=app
DATABASE_PASSWORD=app
DATABASE_NAME=app
DATABASE_SSL_MODE=disable
REDIS_ADDRESS=redis:6379
AUTH_SECRET_KEY=dev_secret_key_32_characters_long
LOG_LEVEL=debug
```

### Production

```bash
# Production environment (managed via secrets manager)
PRIMARY_ENV=production
SERVER_PORT=8080
SERVER_CORS_ALLOWED_ORIGINS=https://app.example.com

DATABASE_HOST=prod-postgres.example.com
DATABASE_PORT=5432
DATABASE_USER=app_prod
DATABASE_PASSWORD=<from-secrets-manager>
DATABASE_NAME=app_production
DATABASE_SSL_MODE=require
DATABASE_MAX_OPEN_CONNS=50

REDIS_ADDRESS=prod-redis.example.com:6379
REDIS_PASSWORD=<from-secrets-manager>

AUTH_SECRET_KEY=<from-secrets-manager>

INTEGRATION_RESEND_API_KEY=<from-secrets-manager>

OBSERVABILITY_NEWRELIC_LICENSE_KEY=<from-secrets-manager>
OBSERVABILITY_NEWRELIC_APP_NAME=myapp-production

S3_ENDPOINT=https://account-id.r2.cloudflarestorage.com
S3_BUCKET=myapp-prod-backups
S3_ACCESS_KEY_ID=<from-secrets-manager>
S3_SECRET_ACCESS_KEY=<from-secrets-manager>
S3_REGION=auto
S3_FORCE_PATH_STYLE=true

LOG_LEVEL=info
```

---

## Security Best Practices

1. **Never commit secrets** to git
2. **Use secrets manager** in production (AWS Secrets Manager, HashiCorp Vault, etc.)
3. **Rotate credentials** regularly
4. **Use SSL/TLS** for all database connections in production
5. **Set specific CORS origins** (never use `*` in production)
6. **Use strong random values** for `AUTH_SECRET_KEY`
7. **Enable proper logging** but avoid logging sensitive data

---

## Validation

The application validates configuration on startup and will fail fast if:
- Required variables are missing
- Values are invalid (e.g., negative port numbers)
- Database connection cannot be established
- Redis connection cannot be established

Check logs for validation errors:
```bash
make logs
```

---

## What's Next?

- **Set up dependencies**: [Dependencies Guide](./DEPENDENCIES.md)
- **Deploy to production**: [Production Deployment](../operations/PRODUCTION.md)
- **Security guidelines**: [Best Practices](../development/BEST_PRACTICES.md#security)
