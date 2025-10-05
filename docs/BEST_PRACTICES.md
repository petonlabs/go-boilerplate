# Best Practices & Debugging Guide

**Project:** go-boilerplate  
**Last Updated:** October 3, 2025  
**Status:** Production-ready guidelines based on actual implementation

---

## 📋 Table of Contents

1. [Development Workflow](#development-workflow)
2. [Code Quality Standards](#code-quality-standards)
3. [Testing Best Practices](#testing-best-practices)
4. [Security Guidelines](#security-guidelines)
5. [Debugging Techniques](#debugging-techniques)
6. [CI/CD Best Practices](#cicd-best-practices)
7. [Dependency Management](#dependency-management)
8. [Common Issues & Solutions](#common-issues--solutions)

---

## 🔄 Development Workflow

### Pre-Commit Checklist

Before committing code, ensure:

```bash
# 1. Run linting
cd apps/backend
golangci-lint run ./...

# 2. Run tests
go test ./... -v

# 3. Verify dependencies
go mod verify

# 4. Check for vulnerabilities
govulncheck ./...

# 5. Format code
gofmt -w .

# 6. Build to ensure no compilation errors
go build ./cmd/go-boilerplate
```

### Git Pre-Commit Hook

We have a pre-commit hook installed that runs `golangci-lint` automatically:

```bash
# Location: .git/hooks/pre-commit
# To set up:
./scripts/setup-hooks.sh

# To bypass (not recommended):
git commit --no-verify
```

### Branch Strategy

- `main` - Production-ready code
- `dev` - Development branch
- `feat/*` - Feature branches
- `fix/*` - Bug fix branches

### Commit Message Convention

```bash
# Format: <type>: <subject>

# Types:
feat:     # New feature
fix:      # Bug fix
docs:     # Documentation changes
style:    # Code style changes (formatting)
refactor: # Code refactoring
test:     # Test additions/changes
chore:    # Build/tool changes
perf:     # Performance improvements
ci:       # CI/CD changes

# Example:
git commit -m "feat: add user authentication middleware"
git commit -m "fix: resolve memory leak in job processor"
git commit -m "docs: update API documentation"
```

---

## ✅ Code Quality Standards

### Linting Configuration

We use **golangci-lint v2.5.0** with 15 essential linters:

```yaml
# apps/backend/.golangci.yml
linters:
  enable:
    - errcheck       # Catch unchecked errors
    - govet          # Go vet analysis
    - ineffassign    # Detect ineffectual assignments
    - staticcheck    # Advanced static analysis
    - unused         # Find unused code
    - gosec          # Security issues
    - bodyclose      # HTTP body close checks
    - sqlclosecheck  # SQL connection close
    - rowserrcheck   # SQL rows error check
    - errorlint      # Error wrapping issues
    - gocritic       # Code quality suggestions
    - unconvert      # Unnecessary type conversions
    - wastedassign   # Wasted value assignments
    - misspell       # Spelling mistakes
```

### Code Standards

#### 1. **Error Handling**

✅ **Good:**
```go
func processData() error {
    conn, err := db.Connect()
    if err != nil {
        return fmt.Errorf("failed to connect: %w", err)
    }
    defer func() {
        if closeErr := conn.Close(); closeErr != nil {
            log.Error().Err(closeErr).Msg("failed to close connection")
        }
    }()
    // ... rest of code
}
```

❌ **Bad:**
```go
func processData() error {
    conn, _ := db.Connect()  // Ignoring error
    defer conn.Close()        // Ignoring close error
    // ... rest of code
}
```

#### 2. **Context Keys**

✅ **Good:**
```go
type contextKey string

const (
    UserIDKey contextKey = "user_id"
)

// Use custom type to avoid collisions
ctx = context.WithValue(ctx, UserIDKey, "123")
```

❌ **Bad:**
```go
// Built-in string type can cause collisions
ctx = context.WithValue(ctx, "user_id", "123")
```

#### 3. **Defer with Exit**

✅ **Good:**
```go
func main() {
    if err := run(); err != nil {
        fmt.Fprintf(os.Stderr, "error: %v\n", err)
        os.Exit(1)
    }
}

func run() error {
    service := NewService()
    defer service.Shutdown()  // Will run properly
    
    if err := service.Start(); err != nil {
        return fmt.Errorf("start failed: %w", err)
    }
    return nil
}
```

❌ **Bad:**
```go
func main() {
    service := NewService()
    defer service.Shutdown()  // Won't run with os.Exit
    
    if err := service.Start(); err != nil {
        log.Fatal(err)  // Calls os.Exit, skips defer
    }
}
```

#### 4. **Error Wrapping**

✅ **Good:**
```go
// Use errors.As for wrapped errors
var validationErr *validator.ValidationErrors
if errors.As(err, &validationErr) {
    // Handle validation errors
}

var invalidErr *validator.InvalidValidationError
if errors.As(err, &invalidErr) {
    // Handle invalid validation errors
}
```

❌ **Bad:**
```go
// Type assertion fails on wrapped errors
if validationErr, ok := err.(validator.ValidationErrors); ok {
    // Won't work with wrapped errors
}
```

---

## 🧪 Testing Best Practices

### Test Structure

```go
func TestServiceMethod(t *testing.T) {
    // Setup
    ctx := context.Background()
    service := NewTestService(t)
    defer service.Cleanup()
    
    // Execute
    result, err := service.Method(ctx, input)
    
    // Assert
    require.NoError(t, err)
    assert.Equal(t, expected, result)
}
```

### Test Helpers

```go
// internal/testing/helpers.go
func NewTestService(t *testing.T) *Service {
    t.Helper()
    // Setup test service
}

func CleanupDatabase(t *testing.T, db *DB) {
    t.Helper()
    t.Cleanup(func() {
        // Cleanup logic
    })
}
```

### Running Tests

```bash
# Run all tests
go test ./...

# Run with race detection
go test -race ./...

# Run with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run specific test
go test -v -run TestSpecificFunction ./path/to/package

# Run tests with timeout
go test -timeout 30s ./...
```

### Using Testcontainers

```go
func TestWithDatabase(t *testing.T) {
    ctx := context.Background()
    
    // Start PostgreSQL container
    container, err := postgres.RunContainer(ctx,
        testcontainers.WithImage("postgres:15"),
        postgres.WithDatabase("testdb"),
        postgres.WithUsername("test"),
        postgres.WithPassword("test"),
    )
    require.NoError(t, err)
    defer container.Terminate(ctx)
    
    // Get connection string
    connStr, err := container.ConnectionString(ctx)
    require.NoError(t, err)
    
    // Use in tests
    db, err := sql.Open("pgx", connStr)
    require.NoError(t, err)
    defer db.Close()
}
```

---

## 🔒 Security Guidelines

### 1. **Dependency Management**

```bash
# Check for vulnerabilities
govulncheck ./...

# Update vulnerable dependencies
go get package@latest
go mod tidy

# Verify all modules
go mod verify
```

### 2. **Security Scanning**

```bash
# Run gosec
gosec ./...

# Run gosec with JSON output
gosec -fmt=json -out=report.json ./...

# Check specific rules
gosec -include=G104,G401 ./...
```

### 3. **Secrets Management**

✅ **Good:**
```go
// Use environment variables
apiKey := os.Getenv("API_KEY")
if apiKey == "" {
    return errors.New("API_KEY not set")
}
```

❌ **Bad:**
```go
// Never hardcode secrets
const apiKey = "sk_live_abc123..."  // NEVER DO THIS
```

### 4. **SQL Injection Prevention**

✅ **Good:**
```go
// Use parameterized queries
query := "SELECT * FROM users WHERE id = $1"
row := db.QueryRow(ctx, query, userID)
```

❌ **Bad:**
```go
// String concatenation is dangerous
query := fmt.Sprintf("SELECT * FROM users WHERE id = %s", userID)
row := db.QueryRow(ctx, query)
```

---

## 🐛 Debugging Techniques

### Local Development Debugging

#### 1. **Using Delve (dlv)**

```bash
# Install Delve
go install github.com/go-delve/delve/cmd/dlv@latest

# Debug main package
cd apps/backend
dlv debug ./cmd/go-boilerplate

# Debug tests
dlv test ./internal/handler

# Common commands in dlv:
# b main.main        - Set breakpoint
# c                  - Continue execution
# n                  - Next line
# s                  - Step into
# p variable         - Print variable
# bt                 - Backtrace
# q                  - Quit
```

#### 2. **Logging for Debugging**

```go
// Add temporary debug logging
log.Debug().
    Str("user_id", userID).
    Interface("request", req).
    Msg("debugging request processing")

// Use log levels appropriately
log.Trace().Msg("very detailed trace")  // Most detailed
log.Debug().Msg("debug information")    // Debug info
log.Info().Msg("informational")         // Normal operations
log.Warn().Msg("warning")               // Warnings
log.Error().Msg("error occurred")       // Errors
```

#### 3. **Printf Debugging**

```go
// Quick debugging (remove before commit)
// Prefer structured logging even for temporary debug statements. If a full
// logger isn't available, create a temporary zerolog console logger.
tmp := zerolog.New(zerolog.NewConsoleWriter()).With().Timestamp().Logger()
tmp.Debug().Str("user_id", userID).Interface("status", status).Msg("DEBUG")

// Better: Use application logger
appLog.Debug().Str("user_id", userID).Interface("status", status).Msg("DEBUG")
```

### CI/CD Debugging

#### 1. **Local CI Simulation**

```bash
# Run the full CI suite locally
./scripts/test-ci-locally.sh

# This runs:
# - go mod download
# - go mod verify
# - gofmt check
# - golangci-lint
# - go vet
# - go test
# - go build
```

#### 2. **GitHub Actions Debugging**

```yaml
# Add debug step to workflow
- name: Debug environment
  run: |
    echo "Go version: $(go version)"
    echo "Working directory: $(pwd)"
    echo "Files: $(ls -la)"
    go env
    go list -m all | head -20
```

#### 3. **Reproduce CI Failures Locally**

```bash
# Use same Go version as CI
go version  # Check your version

# If different, install correct version:
# Download from https://go.dev/dl/

# Use same linter version
golangci-lint --version  # Should be v2.5.0

# Run same commands as CI
go mod download
go mod verify
golangci-lint run ./...
go test -race ./...
```

### Common Debugging Scenarios

#### 1. **Memory Leaks**

```bash
# Run with memory profiling
go test -memprofile=mem.prof ./...

# Analyze with pprof
go tool pprof mem.prof
# In pprof: top10, list FunctionName

# Check for goroutine leaks
go test -run TestSpecific -count=1000 ./...
# Monitor goroutine count
```

#### 2. **Race Conditions**

```bash
# Always test with race detector
go test -race ./...

# Run specific test multiple times
go test -race -count=100 -run TestConcurrentAccess ./...

# Example race condition fix:
# Before:
var counter int
func increment() { counter++ }  // Race!

# After:
var counter int64
func increment() { atomic.AddInt64(&counter, 1) }  // Safe
```

#### 3. **Database Issues**

```go
// Enable query logging
import "github.com/jackc/pgx/v5/tracelog"

config.ConnConfig.Tracer = &tracelog.TraceLog{
    Logger:   pgxLogger,
    LogLevel: tracelog.LogLevelDebug,
}

// Check connection pool stats
stats := pool.Stat()
log.Debug().
    Int32("total_conns", stats.TotalConns()).
    Int32("idle_conns", stats.IdleConns()).
    Int32("acquired_conns", stats.AcquiredConns()).
    Msg("connection pool stats")
```

#### 4. **API Request Debugging**

```go
// Log request/response in middleware
func DebugMiddleware() echo.MiddlewareFunc {
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {
            // Log request
            reqBody, _ := io.ReadAll(c.Request().Body)
            log.Debug().
                Str("method", c.Request().Method).
                Str("path", c.Path()).
                Str("body", string(reqBody)).
                Msg("incoming request")
            
            // Restore body for handler
            c.Request().Body = io.NopCloser(bytes.NewBuffer(reqBody))
            
            return next(c)
        }
    }
}
```

---

## 🚀 CI/CD Best Practices

### Workflow Optimization

#### 1. **Caching Strategy**

```yaml
# Use built-in cache in setup-go
- name: Set up Go
  uses: actions/setup-go@v6
  with:
    go-version: '1.25.0'
    cache-dependency-path: apps/backend/go.sum
```

#### 2. **Parallel Jobs**

```yaml
# Separate concerns for faster feedback
jobs:
  build:
    name: Build & Test
    # Quick feedback on code quality
    
  security:
    name: Security Scans
    # Slower scans run in parallel
```

#### 3. **Fail Fast**

```yaml
# Run quick checks first
steps:
  - name: Check formatting
    run: gofmt -l .
    
  - name: Run linter
    uses: golangci/golangci-lint-action@v7
    
  - name: Run tests
    run: go test ./...
```

### Monitoring CI Health

```bash
# Check workflow status
gh workflow view ci

# List recent runs
gh run list --workflow=ci.yml

# View logs for failed run
gh run view <run-id> --log

# Re-run failed jobs
gh run rerun <run-id>
```

---

## 📦 Dependency Management

### Best Practices

#### 1. **Regular Updates**

```bash
# Check for outdated dependencies
go list -u -m all

# Update specific dependency
go get package@latest

# Update all dependencies (careful!)
go get -u ./...
go mod tidy
```

#### 2. **Vulnerability Scanning**

```bash
# Scan for vulnerabilities (do this weekly)
govulncheck ./...

# Fix vulnerabilities
go get package@v1.2.3  # Update to fixed version
go mod tidy
```

#### 3. **Dependency Auditing**

```bash
# List all dependencies
go list -m all > dependencies.txt

# Check dependency licenses
go-licenses check ./...

# Visualize dependency tree
go mod graph | grep package-name
```

### Version Pinning

```go
// go.mod
require (
    github.com/labstack/echo/v4 v4.13.4
    // Pin to specific version for stability
)
```

---

## 🔧 Common Issues & Solutions

### Issue 1: "undefined" Errors in CI but Works Locally

**Symptoms:**
```
error: undefined: echo.New
error: undefined: validator.New
```

**Root Cause:** Version mismatch or cache issues

**Solution:**
```bash
# In CI workflow:
- name: Verify dependencies
  run: |
    go mod download
    go mod verify
    
# Locally:
go clean -modcache
go mod download
go mod verify
```

### Issue 2: golangci-lint Version Mismatch

**Symptoms:**
```
invalid version string 'v2.5.0', golangci-lint v2 is not supported
```

**Root Cause:** Using golangci-lint-action@v6 with v2.x linter

**Solution:**
```yaml
# Update to v7 action
- name: Run golangci-lint
  uses: golangci/golangci-lint-action@v7  # v7 supports v2.x
  with:
    version: v2.5.0
```

### Issue 3: Linting Issues Not Caught Locally

**Symptoms:** CI fails but local linting passes

**Root Cause:** Different linter versions or configurations

**Solution:**
```bash
# Check versions
golangci-lint --version  # Should match CI

# Install correct version
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | \
  sh -s -- -b $(go env GOPATH)/bin v2.5.0

# Verify config
golangci-lint config verify
```

### Issue 4: Test Failures with Testcontainers

**Symptoms:** Tests fail with Docker errors

**Root Cause:** Docker not running or permissions

**Solution:**
```bash
# Check Docker is running
docker ps

# Check Docker socket permissions
ls -la /var/run/docker.sock

# For macOS/Linux:
sudo chmod 666 /var/run/docker.sock

# For CI (GitHub Actions):
services:
  docker:
    image: docker:dind
```

### Issue 5: Memory/Performance Issues

**Symptoms:** Slow tests, high memory usage

**Solution:**
```bash
# Profile memory
go test -memprofile=mem.prof ./...
go tool pprof mem.prof

# Profile CPU
go test -cpuprofile=cpu.prof ./...
go tool pprof cpu.prof

# Find goroutine leaks
go test -trace=trace.out ./...
go tool trace trace.out
```

### Issue 6: Context Deadline Exceeded

**Symptoms:**
```
context deadline exceeded
```

**Solution:**
```go
// Increase timeout
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

// For tests
func TestLongRunning(t *testing.T) {
    ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
    defer cancel()
    
    // ... test code
}
```

---

## 📚 Additional Resources

### Tools

- [golangci-lint](https://golangci-lint.run/) - Comprehensive Go linter
- [govulncheck](https://pkg.go.dev/golang.org/x/vuln/cmd/govulncheck) - Vulnerability scanner
- [gosec](https://github.com/securego/gosec) - Security analyzer
- [Delve](https://github.com/go-delve/delve) - Go debugger
- [testcontainers-go](https://golang.testcontainers.org/) - Docker containers for tests

### Documentation

- [Effective Go](https://go.dev/doc/effective_go)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Go Security Best Practices](https://go.dev/security/best-practices)
- [GitHub Actions Best Practices](https://docs.github.com/en/actions/security-guides/security-hardening-for-github-actions)

### Project Documentation

- [CI Improvements](./CI_IMPROVEMENTS.md) - Complete CI/CD setup guide
- [Module Configuration](./MODULE_CONFIGURATION.md) - Dependency documentation
- [Dependency Audit](./DEPENDENCY_AUDIT.md) - Dependency verification details
- [Linting Issues](./LINTING_ISSUES.md) - Historical linting fixes

---

## 🎯 Quick Reference

### Daily Development

```bash
# Start development
cd apps/backend

# Run linter
golangci-lint run ./...

# Run tests
go test ./... -v

# Run with race detection
go test -race ./...

# Build
go build ./cmd/go-boilerplate
```

### Before Commit

```bash
# Format code
gofmt -w .

# Run linter
golangci-lint run ./...

# Run tests
go test ./...

# Verify dependencies
go mod verify

# Check vulnerabilities
govulncheck ./...
```

### Debugging

```bash
# Local debugging
dlv debug ./cmd/go-boilerplate

# Test debugging
dlv test ./internal/handler -- -test.run TestName

# CI simulation
./scripts/test-ci-locally.sh
```

---

**Last Updated:** October 3, 2025  
**Maintained By:** Development Team  
**Status:** ✅ Production-ready and actively maintained
