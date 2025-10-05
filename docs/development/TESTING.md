# Testing Guide

Comprehensive guide for testing the go-boilerplate application.

---

## Overview

Our testing strategy includes:
- **Unit Tests**: Test individual functions and methods
- **Integration Tests**: Test database and external service interactions
- **End-to-End Tests**: Test complete workflows
- **Test Containers**: Docker-based integration testing

---

## Running Tests

### Unit Tests

```bash
cd apps/backend

# Run all unit tests
go test ./...

# Run tests with verbose output
go test ./... -v

# Run specific package
go test ./internal/handler

# Run specific test
go test ./internal/handler -run TestHealthCheck
```

### Integration Tests

Integration tests require Docker and use Testcontainers:

```bash
cd apps/backend

# Run integration tests
go test -tags=integration ./...

# Run specific integration test
go test -tags=integration ./internal/repository -run TestUserRepository
```

### With Coverage

```bash
# Generate coverage report
go test -coverprofile=coverage.out ./...

# View coverage in terminal
go tool cover -func=coverage.out

# View coverage in browser
go tool cover -html=coverage.out
```

### With Race Detection

```bash
# Detect race conditions
go test -race ./...

# Combines well with integration tests
go test -race -tags=integration ./...
```

---

## Test Structure

### Unit Test Example

```go
package handler_test

import (
    "net/http"
    "net/http/httptest"
    "testing"
    
    "github.com/labstack/echo/v4"
    "github.com/stretchr/testify/assert"
)

func TestHealthCheck(t *testing.T) {
    // Setup
    e := echo.New()
    req := httptest.NewRequest(http.MethodGet, "/health", nil)
    rec := httptest.NewRecorder()
    c := e.NewContext(req, rec)
    
    // Execute
    err := HealthCheck(c)
    
    // Assert
    assert.NoError(t, err)
    assert.Equal(t, http.StatusOK, rec.Code)
    assert.Contains(t, rec.Body.String(), "healthy")
}
```

### Integration Test Example

```go
//go:build integration
// +build integration

package repository_test

import (
    "context"
    "testing"
    
    "github.com/stretchr/testify/require"
    "github.com/testcontainers/testcontainers-go"
    "github.com/testcontainers/testcontainers-go/wait"
)

func TestUserRepository(t *testing.T) {
    // Start test container
    ctx := context.Background()
    postgres, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
        ContainerRequest: testcontainers.ContainerRequest{
            Image:        "postgres:18-alpine",
            ExposedPorts: []string{"5432/tcp"},
            WaitingFor:   wait.ForListeningPort("5432/tcp"),
            Env: map[string]string{
                "POSTGRES_PASSWORD": "test",
                "POSTGRES_DB":       "test",
            },
        },
        Started: true,
    })
    require.NoError(t, err)
    defer postgres.Terminate(ctx)
    
    // Get connection details
    host, _ := postgres.Host(ctx)
    port, _ := postgres.MappedPort(ctx, "5432")
    
    // Run tests with real database
    repo := NewUserRepository(host, port.Int())
    user, err := repo.Create(ctx, "test@example.com")
    require.NoError(t, err)
    require.NotEmpty(t, user.ID)
}
```

---

## Test Organization

### File Naming

- Unit tests: `*_test.go` in same package
- Integration tests: `*_integration_test.go` with build tag
- Test helpers: `testing_helper.go` or `testutil/` package

### Package Structure

```
internal/
├── handler/
│   ├── handler.go
│   ├── handler_test.go           # Unit tests
│   └── handler_integration_test.go  # Integration tests
├── repository/
│   ├── user.go
│   ├── user_test.go
│   └── user_integration_test.go
└── testutil/
    ├── fixtures.go               # Test data
    ├── mocks.go                 # Mock implementations
    └── containers.go            # Testcontainer helpers
```

---

## Testing Patterns

### Table-Driven Tests

```go
func TestValidateEmail(t *testing.T) {
    tests := []struct {
        name    string
        email   string
        wantErr bool
    }{
        {"valid email", "user@example.com", false},
        {"missing @", "userexample.com", true},
        {"empty string", "", true},
        {"no domain", "user@", true},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateEmail(tt.email)
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

### Using Mocks

```go
type MockUserRepository struct {
    mock.Mock
}

func (m *MockUserRepository) GetByID(ctx context.Context, id string) (*User, error) {
    args := m.Called(ctx, id)
    return args.Get(0).(*User), args.Error(1)
}

func TestUserService(t *testing.T) {
    mockRepo := new(MockUserRepository)
    mockRepo.On("GetByID", mock.Anything, "123").Return(&User{ID: "123"}, nil)
    
    service := NewUserService(mockRepo)
    user, err := service.GetUser(context.Background(), "123")
    
    assert.NoError(t, err)
    assert.Equal(t, "123", user.ID)
    mockRepo.AssertExpectations(t)
}
```

### Setup and Teardown

```go
func TestMain(m *testing.M) {
    // Global setup
    setup()
    
    // Run tests
    code := m.Run()
    
    // Global teardown
    teardown()
    
    os.Exit(code)
}

func TestWithCleanup(t *testing.T) {
    // Test-specific setup
    db := setupTestDB(t)
    t.Cleanup(func() {
        db.Close()
    })
    
    // Test code
}
```

---

## Testcontainers

### PostgreSQL Container

```go
func StartPostgresContainer(t *testing.T) *testcontainers.Container {
    ctx := context.Background()
    
    req := testcontainers.ContainerRequest{
        Image:        "postgres:18-alpine",
        ExposedPorts: []string{"5432/tcp"},
        WaitingFor:   wait.ForListeningPort("5432/tcp"),
        Env: map[string]string{
            "POSTGRES_USER":     "test",
            "POSTGRES_PASSWORD": "test",
            "POSTGRES_DB":       "test",
        },
    }
    
    container, err := testcontainers.GenericContainer(ctx, 
        testcontainers.GenericContainerRequest{
            ContainerRequest: req,
            Started:          true,
        })
    require.NoError(t, err)
    
    t.Cleanup(func() {
        container.Terminate(ctx)
    })
    
    return &container
}
```

### Redis Container

```go
func StartRedisContainer(t *testing.T) *testcontainers.Container {
    ctx := context.Background()
    
    req := testcontainers.ContainerRequest{
        Image:        "redis:8-alpine",
        ExposedPorts: []string{"6379/tcp"},
        WaitingFor:   wait.ForListeningPort("6379/tcp"),
    }
    
    container, err := testcontainers.GenericContainer(ctx,
        testcontainers.GenericContainerRequest{
            ContainerRequest: req,
            Started:          true,
        })
    require.NoError(t, err)
    
    t.Cleanup(func() {
        container.Terminate(ctx)
    })
    
    return &container
}
```

---

## CI Testing

### Local CI Simulation

```bash
# Simulate GitHub Actions locally
cd apps/backend
../../scripts/test-ci-locally.sh
```

This script:
- Cleans environment
- Downloads dependencies
- Runs formatters
- Runs linters
- Builds code
- Runs tests

### GitHub Actions

Tests run automatically on:
- Push to main
- Pull requests
- Manual workflow dispatch

See [CI/CD Guide](../operations/CI_CD.md) for details.

---

## Best Practices

### 1. Test Naming

```go
// Good
func TestUserRepository_CreateUser_Success(t *testing.T) {}
func TestUserRepository_CreateUser_DuplicateEmail(t *testing.T) {}

// Bad
func TestUser(t *testing.T) {}
func TestCreateUser(t *testing.T) {}
```

### 2. Use Subtests

```go
func TestUserService(t *testing.T) {
    t.Run("GetUser", func(t *testing.T) {
        t.Run("Success", func(t *testing.T) {
            // Test success case
        })
        t.Run("NotFound", func(t *testing.T) {
            // Test not found case
        })
    })
}
```

### 3. Test Edge Cases

```go
func TestDivide(t *testing.T) {
    tests := []struct {
        name      string
        a, b      int
        want      float64
        wantPanic bool
    }{
        {"normal", 10, 2, 5.0, false},
        {"zero divisor", 10, 0, 0, true},
        {"negative", -10, 2, -5.0, false},
    }
    // ...
}
```

### 4. Keep Tests Fast

- Mock external dependencies
- Use in-memory databases when possible
- Run expensive tests separately with build tags
- Parallelize independent tests: `t.Parallel()`

### 5. Test Coverage Goals

- Aim for 80%+ coverage
- 100% coverage for critical paths
- Don't test vendor code
- Focus on business logic

---

## Debugging Tests

### Run Single Test

```bash
go test -v -run TestUserService
```

### With Debugging

```bash
# Using delve
dlv test ./internal/handler -- -test.run TestHealthCheck

# In VS Code, use launch.json configuration
```

### Verbose Output

```bash
# Show all logs
go test -v ./...

# With additional debugging
go test -v -count=1 ./... # Disable cache
```

---

## Common Issues

### Tests Pass Locally But Fail in CI

- Check for timing issues (use explicit waits)
- Ensure deterministic test data
- Check for file system dependencies
- Verify environment variables

### Flaky Tests

- Look for race conditions: `go test -race`
- Check for shared state between tests
- Add proper cleanup in `t.Cleanup()`
- Use deterministic random seeds in tests

### Slow Tests

- Use `t.Parallel()` for independent tests
- Mock external services
- Use build tags for slow integration tests
- Profile tests: `go test -cpuprofile cpu.prof`

---

## Resources

- [Go Testing Package](https://pkg.go.dev/testing)
- [Testify Documentation](https://github.com/stretchr/testify)
- [Testcontainers Go](https://golang.testcontainers.org/)
- [Table Driven Tests](https://dave.cheney.net/2019/05/07/prefer-table-driven-tests)

---

## What's Next?

- **Write quality code**: [Best Practices](./BEST_PRACTICES.md)
- **Debug issues**: [Best Practices - Debugging](./BEST_PRACTICES.md#debugging-techniques)
- **Set up CI**: [CI/CD Guide](../operations/CI_CD.md)
