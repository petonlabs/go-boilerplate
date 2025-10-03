# Scripts Directory

This directory contains utility scripts for development, CI/CD, and maintenance tasks.

## Setup Scripts

### `setup-hooks.sh`
Installs Git hooks for the repository (including pre-commit linting).

```bash
./scripts/setup-hooks.sh
```

**What it does:**
- Copies `git-hooks/pre-commit` to `.git/hooks/`
- Makes hooks executable
- Enables automatic linting before commits

## Linting Scripts

### `run-lint.sh`
Runs golangci-lint on the backend codebase.

```bash
./scripts/run-lint.sh
```

**Features:**
- Auto-installs golangci-lint if not found
- Runs with project configuration (`.golangci.yml`)
- Colored output for easy reading
- Exit codes: 0 = success, 1 = failure

**Use cases:**
- Manual code quality checks
- CI/CD pipeline integration
- Pre-commit verification (via hook)

## Testing Scripts

### `test-ci-locally.sh`
Simulates the entire GitHub Actions CI pipeline locally.

```bash
cd apps/backend
../../scripts/test-ci-locally.sh
```

**Features:**
- Cleans environment (simulates fresh clone)
- Sets CI environment variables
- Downloads and verifies modules
- Runs all CI checks:
  - Code formatting
  - go vet
  - Build all packages
  - Run tests
  - Build binary
  - Type checking
  - Verify critical imports

**Environment variables:**
- Sets dummy values for missing env vars
- Won't fail if DATABASE_URL, REDIS_URL, etc. are not set
- Safe for local testing without full environment setup

## Git Hooks

### `git-hooks/pre-commit`
Automatically runs golangci-lint before each commit.

**Behavior:**
- Only runs if `.go` files are staged in `apps/backend/`
- Skips if golangci-lint is not installed (with warning)
- Blocks commit if linting fails
- Can be bypassed with `git commit --no-verify`

**Installation:**
```bash
./scripts/setup-hooks.sh
```

**Manual installation:**
```bash
cp scripts/git-hooks/pre-commit .git/hooks/pre-commit
chmod +x .git/hooks/pre-commit
```

## Backup Scripts

### `backup/run-backup.sh`
Database backup script (see backup/README.md for details).

### `backup/restore.sh`
Database restore script (see backup/README.md for details).

## Usage Examples

### Before Committing
```bash
# Option 1: Manual lint check
./scripts/run-lint.sh

# Option 2: Let pre-commit hook handle it
git add .
git commit -m "Your message"  # Hook runs automatically
```

### Bypassing Pre-commit Hook
```bash
# Not recommended, but available for emergencies
git commit --no-verify -m "Your message"
```

### Local CI Testing
```bash
# Test exactly what will run in GitHub Actions
cd apps/backend
../../scripts/test-ci-locally.sh

# If it passes locally, it should pass in CI
```

### Installing Development Tools
```bash
# Install golangci-lint (macOS)
brew install golangci-lint

# Install golangci-lint (Linux/macOS via script)
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.59.0

# Add to PATH if needed
export PATH="$PATH:$(go env GOPATH)/bin"
```

## Script Maintenance

### Adding New Scripts
1. Create script in appropriate subdirectory
2. Add shebang: `#!/usr/bin/env bash`
3. Add `set -euo pipefail` for safety
4. Make executable: `chmod +x scripts/your-script.sh`
5. Document in this README

### Shell Script Best Practices
- Use `set -euo pipefail` for error handling
- Use colors for output clarity
- Provide clear success/failure messages
- Exit with appropriate codes (0 = success, non-zero = failure)
- Handle missing dependencies gracefully

## Troubleshooting

### golangci-lint not found
```bash
# Install via Homebrew (macOS)
brew install golangci-lint

# Or via install script
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.59.0
```

### Pre-commit hook not running
```bash
# Reinstall hooks
./scripts/setup-hooks.sh

# Check if hook exists and is executable
ls -la .git/hooks/pre-commit
```

### CI simulation fails locally
```bash
# Clean everything and retry
cd apps/backend
go clean -cache -modcache -testcache
go mod download
go mod verify
```

## Environment Variables

Scripts that need environment variables will either:
1. Use dummy values for testing (like `test-ci-locally.sh`)
2. Fail with clear error messages if required vars are missing
3. Skip certain checks if optional vars are not set

### Required for Full Functionality
- `DATABASE_URL`: PostgreSQL connection string
- `REDIS_URL`: Redis connection string
- `CLERK_SECRET_KEY`: Clerk authentication key
- `NEW_RELIC_LICENSE_KEY`: New Relic monitoring key
- `RESEND_API_KEY`: Resend email service key

### Test/CI Scripts
Most scripts set dummy values, so full environment setup is not required for basic testing.

## Integration with CI/CD

### GitHub Actions
The CI workflow (`.github/workflows/ci.yml`) uses patterns from these scripts:
- Module download and verification
- Linting with golangci-lint
- Building and testing
- Security scanning

### Local Development
These scripts help ensure your code will pass CI before pushing:
1. `run-lint.sh` - Same linter as CI
2. `test-ci-locally.sh` - Simulates full CI pipeline
3. Pre-commit hook - Prevents bad commits

## Contributing

When adding new scripts:
1. Follow existing patterns and conventions
2. Add appropriate error handling
3. Use colors for output clarity
4. Document in this README
5. Test thoroughly before committing
