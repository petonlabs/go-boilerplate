# Architecture Guide

Overview of the go-boilerplate architecture, design patterns, and project structure.

---

## Overview

Go-boilerplate follows **Clean Architecture** principles with clear separation of concerns:

```
┌─────────────────────────────────────┐
│         External Interfaces          │
│    (HTTP, CLI, gRPC, WebSockets)    │
└──────────────┬──────────────────────┘
               │
┌──────────────▼──────────────────────┐
│           Handlers                   │
│   (HTTP handlers, controllers)      │
└──────────────┬──────────────────────┘
               │
┌──────────────▼──────────────────────┐
│           Services                   │
│     (Business logic layer)          │
└──────────────┬──────────────────────┘
               │
┌──────────────▼──────────────────────┐
│         Repositories                 │
│      (Data access layer)            │
└──────────────┬──────────────────────┘
               │
┌──────────────▼──────────────────────┐
│      External Services               │
│  (Database, Cache, Email, APIs)     │
└─────────────────────────────────────┘
```

---

## Project Structure

```
apps/backend/
├── cmd/
│   └── go-boilerplate/      # Application entry point
│       └── main.go          # Wires up dependencies, starts server
│
├── internal/
│   ├── config/              # Configuration loading and validation
│   │   └── config.go
│   │
│   ├── database/            # Database connection and migrations
│   │   ├── database.go
│   │   └── migrations/
│   │
│   ├── handler/             # HTTP handlers (presentation layer)
│   │   ├── health.go
│   │   ├── user.go
│   │   └── auth_handlers.go
│   │
│   ├── service/             # Business logic
│   │   ├── user_service.go
│   │   └── auth_service.go
│   │
│   ├── repository/          # Data access layer
│   │   ├── user_repository.go
│   │   └── session_repository.go
│   │
│   ├── model/               # Domain models
│   │   ├── user.go
│   │   └── base.go
│   │
│   ├── middleware/          # HTTP middleware
│   │   ├── auth.go
│   │   ├── cors.go
│   │   └── logging.go
│   │
│   ├── router/              # Route definitions
│   │   └── router.go
│   │
│   ├── logger/              # Logging configuration
│   │   └── logger.go
│   │
│   ├── errs/                # Error handling
│   │   └── errors.go
│   │
│   └── server/              # Server setup and lifecycle
│       └── server.go
│
├── Dockerfile               # Multi-stage Docker build
├── Taskfile.yml            # Task automation
├── go.mod                  # Go dependencies
└── .env.example            # Environment template
```

---

## Layers Explained

### 1. Handlers (Presentation Layer)

**Location**: `internal/handler/`

**Responsibility**: Handle HTTP requests and responses

**Characteristics**:
- Parse request data
- Call service layer
- Format responses
- Handle HTTP-specific concerns (status codes, headers)
- No business logic

**Example**:
```go
func (h *UserHandler) GetUser(c echo.Context) error {
    id := c.Param("id")
    
    user, err := h.userService.GetByID(c.Request().Context(), id)
    if err != nil {
        return echo.NewHTTPError(http.StatusNotFound, err.Error())
    }
    
    return c.JSON(http.StatusOK, user)
}
```

### 2. Services (Business Logic Layer)

**Location**: `internal/service/`

**Responsibility**: Implement business logic

**Characteristics**:
- Contains domain logic
- Orchestrates operations across repositories
- Validates business rules
- Independent of HTTP layer

**Example**:
```go
func (s *UserService) GetByID(ctx context.Context, id string) (*model.User, error) {
    // Validate ID format
    if !isValidUUID(id) {
        return nil, errs.ErrInvalidID
    }
    
    // Fetch from repository
    user, err := s.userRepo.GetByID(ctx, id)
    if err != nil {
        return nil, err
    }
    
    // Business logic: check if user is active
    if !user.IsActive {
        return nil, errs.ErrUserInactive
    }
    
    return user, nil
}
```

### 3. Repositories (Data Access Layer)

**Location**: `internal/repository/`

**Responsibility**: Database operations

**Characteristics**:
- CRUD operations
- Query construction
- Transaction management
- Database-specific logic
- No business logic

**Example**:
```go
func (r *UserRepository) GetByID(ctx context.Context, id string) (*model.User, error) {
    query := `SELECT id, email, created_at FROM users WHERE id = $1`
    
    var user model.User
    err := r.db.GetContext(ctx, &user, query, id)
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, errs.ErrNotFound
        }
        return nil, err
    }
    
    return &user, nil
}
```

### 4. Models (Domain Layer)

**Location**: `internal/model/`

**Responsibility**: Define domain entities

**Characteristics**:
- Struct definitions
- No dependencies on other layers
- Validation methods
- Business rules embedded in the model

**Example**:
```go
type User struct {
    Base
    Email     string    `json:"email" db:"email"`
    FirstName string    `json:"firstName" db:"first_name"`
    LastName  string    `json:"lastName" db:"last_name"`
    IsActive  bool      `json:"isActive" db:"is_active"`
}

func (u *User) FullName() string {
    return fmt.Sprintf("%s %s", u.FirstName, u.LastName)
}
```

---

## Design Patterns

### Dependency Injection

Dependencies are injected through constructors:

```go
type UserService struct {
    userRepo  repository.UserRepository
    emailSvc  service.EmailService
    logger    *zerolog.Logger
}

func NewUserService(
    userRepo repository.UserRepository,
    emailSvc service.EmailService,
    logger *zerolog.Logger,
) *UserService {
    return &UserService{
        userRepo: userRepo,
        emailSvc: emailSvc,
        logger:   logger,
    }
}
```

### Interface-Based Design

Define interfaces for abstraction:

```go
type UserRepository interface {
    GetByID(ctx context.Context, id string) (*model.User, error)
    Create(ctx context.Context, user *model.User) error
    Update(ctx context.Context, user *model.User) error
    Delete(ctx context.Context, id string) error
}
```

This allows:
- Easy testing with mocks
- Swappable implementations
- Loose coupling

### Error Handling

Centralized error definitions:

```go
// internal/errs/errors.go
var (
    ErrNotFound      = errors.New("resource not found")
    ErrInvalidInput  = errors.New("invalid input")
    ErrUnauthorized  = errors.New("unauthorized")
)
```

Wrap errors for context:

```go
if err != nil {
    return fmt.Errorf("failed to get user: %w", err)
}
```

---

## Key Components

### Server

**Location**: `internal/server/`

Central server struct that holds all dependencies:

```go
type Server struct {
    Config        *config.Config
    Echo          *echo.Echo
    DB            *sqlx.DB
    Redis         *redis.Client
    Logger        *zerolog.Logger
    LoggerService *logger.Service
}
```

### Middleware

**Location**: `internal/middleware/`

Request/response interceptors:

```go
type Middlewares struct {
    Global          *GlobalMiddlewares
    Auth            *AuthMiddleware
    ContextEnhancer *ContextEnhancer
    Tracing         *TracingMiddleware
    RateLimit       *RateLimitMiddleware
}
```

### Router

**Location**: `internal/router/`

Route registration and grouping:

```go
func SetupRoutes(s *server.Server, m *middleware.Middlewares, h *handler.Handlers) {
    api := s.Echo.Group("/api/v1")
    api.Use(m.Auth.RequireAuth())
    
    api.GET("/users/:id", h.User.GetUser)
    api.POST("/users", h.User.CreateUser)
}
```

---

## Database

### Connection Pooling

Configured via environment variables:
- `DATABASE_MAX_OPEN_CONNS`: Max open connections
- `DATABASE_MAX_IDLE_CONNS`: Max idle connections
- `DATABASE_CONN_MAX_LIFETIME`: Connection lifetime
- `DATABASE_CONN_MAX_IDLE_TIME`: Idle connection timeout

### Migrations

Using `tern` migration tool:

```sql
-- migrations/001_create_users.sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

---- create above / drop below ----

DROP TABLE users;
```

### Transactions

```go
func (r *UserRepository) CreateWithProfile(ctx context.Context, user *model.User, profile *model.Profile) error {
    tx, err := r.db.BeginTxx(ctx, nil)
    if err != nil {
        return err
    }
    defer tx.Rollback()
    
    if err := r.createUser(ctx, tx, user); err != nil {
        return err
    }
    
    if err := r.createProfile(ctx, tx, profile); err != nil {
        return err
    }
    
    return tx.Commit()
}
```

---

## Caching Strategy

### Redis Usage

- Session storage
- Rate limiting
- Background job queue (Asynq)
- Cache frequently accessed data

### Cache Patterns

**Cache-Aside**:
```go
func (s *UserService) GetByID(ctx context.Context, id string) (*model.User, error) {
    // Try cache first
    cached, err := s.cache.Get(ctx, "user:"+id)
    if err == nil {
        return cached, nil
    }
    
    // Fetch from DB
    user, err := s.userRepo.GetByID(ctx, id)
    if err != nil {
        return nil, err
    }
    
    // Store in cache
    s.cache.Set(ctx, "user:"+id, user, 5*time.Minute)
    
    return user, nil
}
```

---

## Background Jobs

Using **Asynq** for async job processing:

```go
// Enqueue job
client := asynq.NewClient(redisOpt)
task := asynq.NewTask("email:send", payload)
client.Enqueue(task)

// Process job
func HandleEmailTask(ctx context.Context, t *asynq.Task) error {
    var payload EmailPayload
    json.Unmarshal(t.Payload(), &payload)
    return sendEmail(payload)
}
```

---

## Observability

### Structured Logging

Using **zerolog**:

```go
logger.Info().
    Str("user_id", userID).
    Str("action", "login").
    Dur("duration", duration).
    Msg("User logged in")
```

### APM Integration

Using **New Relic**:

```go
txn := newrelic.FromContext(ctx)
txn.AddAttribute("user_id", userID)
txn.NoticeError(err)
```

### Health Checks

```go
type HealthResponse struct {
    Status   string            `json:"status"`
    Postgres *ServiceHealth    `json:"postgres"`
    Redis    *ServiceHealth    `json:"redis"`
}
```

---

## Security

### Authentication Flow

1. User sends credentials
2. Handler validates input
3. Service verifies credentials
4. Generate JWT token
5. Return token to client

### Authorization

Middleware checks JWT token:
```go
func (m *AuthMiddleware) RequireAuth() echo.MiddlewareFunc {
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {
            token := extractToken(c)
            claims, err := validateToken(token)
            if err != nil {
                return echo.NewHTTPError(http.StatusUnauthorized)
            }
            c.Set("user_id", claims.UserID)
            return next(c)
        }
    }
}
```

---

## Testing Strategy

- **Unit Tests**: Test individual functions
- **Integration Tests**: Test with real database (testcontainers)
- **E2E Tests**: Test complete workflows
- **Mock Dependencies**: Use interfaces for testing

See [Testing Guide](../development/TESTING.md) for details.

---

## What's Next?

- **Understand dependencies**: [Dependencies Guide](./DEPENDENCIES.md)
- **Configure environment**: [Configuration Reference](./CONFIGURATION.md)
- **Learn authentication**: [Authentication Guide](./AUTHENTICATION.md)
- **Best practices**: [Best Practices](../development/BEST_PRACTICES.md)
