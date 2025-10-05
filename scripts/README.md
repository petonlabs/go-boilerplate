# Scripts Directory

Utility scripts for development, CI/CD, and maintenance tasks.

---

## 🛠️ Available Scripts

### Development

#### `run-lint.sh`
Run golangci-lint on the backend codebase.

```bash
./scripts/run-lint.sh
```

**Features**:
- Auto-installs golangci-lint if not found
- Uses project configuration (`.golangci.yml`)
- Colored output for readability

#### `test-ci-locally.sh`
Simulate the GitHub Actions CI pipeline locally.

```bash
cd apps/backend
../../scripts/test-ci-locally.sh
```

**What it does**:
- Cleans environment (simulates fresh clone)
- Downloads and verifies Go modules
- Runs formatters, linters, and tests
- Builds the binary
- Sets dummy environment variables for testing

**Use this** before pushing to catch CI failures early.

### Git Hooks

#### `setup-hooks.sh`
Install Git hooks for automatic linting.

```bash
./scripts/setup-hooks.sh
```

**Installs**:
- Pre-commit hook that runs golangci-lint
- Prevents commits with linting errors
- Can be bypassed with `--no-verify` if needed

#### `git-hooks/pre-commit`
Pre-commit hook that runs golangci-lint automatically.

**Behavior**:
- Runs only if `.go` files are staged
- Blocks commit if linting fails
- Shows clear error messages

### Development Environment

#### `dev-start.sh`
Start the full development stack (backend, database, Redis, frontend).

```bash
# Start everything
./scripts/dev-start.sh

# Skip frontend
./scripts/dev-start.sh --no-frontend

# Run backend in foreground for debugging
./scripts/dev-start.sh --foreground
```

**Features**:
- Starts Docker Compose services
- Waits for services to be healthy
- Runs database migrations
- Starts backend and frontend dev servers

### Backup Scripts

Located in `scripts/backup/`:

#### `entrypoint.sh`
Backup service entrypoint for Docker container.

#### `run-backup.sh`
Create a database backup and upload to S3.

#### `restore.sh`
Restore database from S3 backup.

See [Production Deployment](../docs/operations/PRODUCTION.md#backups) for usage.

---

## 📋 Usage Examples

### Before Committing

```bash
# Option 1: Manual check
./scripts/run-lint.sh

# Option 2: Install pre-commit hook
./scripts/setup-hooks.sh
git commit -m "Your message"  # Hook runs automatically
```

### Testing CI Locally

```bash
cd apps/backend
../../scripts/test-ci-locally.sh
```

If this passes, your code should pass in GitHub Actions.

### Starting Development

```bash
# Full stack
./scripts/dev-start.sh

# Backend only
make docker-up
make backend-run
```

---

## 🔧 Script Requirements

### System Dependencies

- **bash**: Shell for running scripts
- **Docker**: For containers
- **Docker Compose**: For orchestration
- **Go 1.24+**: For building backend
- **make**: For Makefile targets

### Go Tools (Auto-installed)

- **golangci-lint**: Code linting
- **tern**: Database migrations

---

## 🚀 CI/CD Integration

### GitHub Actions

Scripts used in CI pipeline:

1. **Linting**: `run-lint.sh`
2. **Testing**: Built into CI workflow
3. **Building**: Go build commands

See [CI/CD Guide](../docs/operations/CI_CD.md) for CI configuration.

### Local CI Testing

The `test-ci-locally.sh` script exactly mirrors the GitHub Actions workflow:

```bash
cd apps/backend
../../scripts/test-ci-locally.sh
```

**Environment variables**:
- Sets dummy values for missing vars
- Safe for local testing
- Won't fail without full environment setup

---

## 📝 Environment Variables

Scripts that need environment variables either:
1. Use dummy values for testing (like `test-ci-locally.sh`)
2. Fail with clear error messages if required vars are missing
3. Skip certain checks if optional vars are not set

### Required for Full Functionality

See [Configuration Reference](../docs/reference/CONFIGURATION.md) for complete list.

**For local development**, most scripts work without full environment setup.

---

## 🔍 Troubleshooting

### Script Permission Denied

```bash
chmod +x ./scripts/script-name.sh
```

### Linter Not Found

The `run-lint.sh` script auto-installs golangci-lint:
```bash
./scripts/run-lint.sh
```

### Docker Not Running

```bash
# Check Docker status
docker ps

# Start Docker Desktop or daemon
```

### Tests Failing Locally

```bash
# Check environment
cat apps/backend/.env

# Restart services
make docker-down
make docker-up

# Run tests again
cd apps/backend
go test ./...
```

---

## 🤝 Contributing Scripts

When adding new scripts:

1. **Follow naming conventions**: Use kebab-case
2. **Add shebang**: `#!/usr/bin/env bash`
3. **Add error handling**: `set -euo pipefail`
4. **Document in this README**: Explain purpose and usage
5. **Make executable**: `chmod +x script.sh`
6. **Test thoroughly**: Ensure works on different systems

---

## 📚 Related Documentation

- [Local Development](../docs/getting-started/LOCAL_DEVELOPMENT.md)
- [CI/CD Guide](../docs/operations/CI_CD.md)
- [Best Practices](../docs/development/BEST_PRACTICES.md)
- [Testing Guide](../docs/development/TESTING.md)

---

**Need help?** See the [main documentation](../docs/) or open an issue.
