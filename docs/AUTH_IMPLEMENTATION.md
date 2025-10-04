# Authentication System Implementation

## Overview
This document describes the authentication system implemented in the go-boilerplate project, including webhook signature verification, user authentication flows, and configurable deletion scheduling.

## Features Implemented

### 1. Webhook Signature Verification (Svix/Clerk)
- **Location**: `internal/handler/webhook.go`
- **Algorithm**: HMAC SHA256
- **Format**: Parses `Svix-Signature` header with `v1=<hex>` format
- **Process**:
  1. Reads request body bytes
  2. Parses signature from header
  3. Computes HMAC SHA256 of body using webhook signing secret
  4. Compares signatures in constant time
- **Configuration**: `config.Auth.WebhookSigningSecret`

### 2. Authentication HTTP Handlers
- **Location**: `internal/handler/auth_handlers.go`

#### Endpoints:
1. **POST /auth/register**
   - Registers new user with email and password
   - Returns user ID
   
2. **POST /auth/login**
   - Authenticates user with email and password
   - Updates last login timestamp
   - Returns user ID

3. **POST /auth/password/request**
   - Generates password reset token
   - Sets expiry based on `config.Auth.PasswordResetTTL`
   - Returns reset token

4. **POST /auth/password/reset**
   - Validates reset token and expiry
   - Updates password with bcrypt hash
   - Clears reset token

5. **POST /auth/schedule_deletion**
   - Schedules user account deletion
   - Accepts custom TTL in seconds (overrides default)
   - Enqueues deletion job via Asynq
   - Default TTL: `config.Auth.DeletionDefaultTTL`

6. **POST /auth/cancel_deletion**
   - Cancels scheduled deletion
   - Clears `deletion_scheduled_at` timestamp
   - Allows user to keep account

### 3. Authentication Service
- **Location**: `internal/service/auth.go`

#### Methods:
- `RegisterUser(email, password)`: Creates user with bcrypt-hashed password
- `Login(email, password)`: Verifies credentials with bcrypt comparison
- `RequestPasswordReset(email, ttl)`: Generates 16-byte hex token with expiry
- `ResetPassword(token, newPassword)`: Validates token and updates password
- `ScheduleDeletion(userID, ttl)`: Sets scheduled time and enqueues job
- `SyncUser(data)`: Upserts user from Clerk webhook (existing functionality)

### 4. Deletion Worker System
- **Location**: `internal/lib/job/handlers.go`, `internal/lib/job/user_tasks.go`
- **Queue**: Asynq (Redis-backed)
- **Features**:
  - Checks `deletion_scheduled_at` timestamp before executing
  - Only deletes if current time is after scheduled time
  - Supports cancellation (if timestamp cleared, job is skipped)
  - Soft-delete: Sets `deleted_at`, clears `email` and `password_hash`
  - Configurable TTL per deletion request

### 5. Configuration Extensions
- **Location**: `internal/config/config.go`

#### New AuthConfig Fields:
```go
type AuthConfig struct {
    PasswordResetTTL      int    // Seconds until reset token expires
    DeletionDefaultTTL    int    // Default seconds until account deletion
    WebhookSigningSecret  string // Svix/Clerk webhook HMAC secret
    // ... existing fields
}
```

### 6. Test Coverage
- **Location**: `internal/handler/webhook_test.go`, `internal/service/auth_test.go`, `internal/service/auth_webhook_test.go`

#### Tests:
1. **TestClerkWebhookSignatureValid** (3.13s)
   - Unit test for webhook signature verification
   - Creates HMAC signature, validates handler accepts valid signature

2. **TestSyncUserUpserts** (2.84s)
   - Integration test with testcontainers
   - Verifies webhook user sync functionality
   - Tests database upsert by external_id

3. **TestRegisterLoginResetAndScheduleDeletion** (6.84s)
   - End-to-end integration test
   - Tests: register → login → password reset → schedule deletion → cancel deletion
   - Verifies account not deleted after cancellation

**Total Test Time**: ~13 seconds  
**Test Status**: ✅ All tests passing

## Architecture Patterns

### Soft Delete
Instead of hard-deleting records:
```sql
UPDATE users 
SET deleted_at = NOW(), 
    email = NULL, 
    password_hash = NULL 
WHERE id = $1
```

### Job Scheduling
- Uses Asynq for background task queue
- Redis connection: `localhost:6379` (configurable)
- Tasks can be cancelled before execution by clearing `deletion_scheduled_at`
- Worker checks scheduled time: `time.Now().Before(*scheduledAt)` → skip execution

### Security
- **Password Hashing**: bcrypt with default cost (10)
- **Reset Tokens**: 16-byte random hex (crypto/rand)
- **Webhook Signatures**: HMAC SHA256, constant-time comparison
- **Token Expiry**: Configurable TTL for reset tokens

## Database Schema

### Users Table Fields Used:
- `id`: Primary key (UUID)
- `external_id`: Clerk user ID
- `email`: User email (cleared on deletion)
- `password_hash`: bcrypt hash (cleared on deletion)
- `password_reset_token`: Temporary reset token
- `password_reset_expires_at`: Token expiry timestamp
- `deletion_scheduled_at`: Scheduled deletion time (nullable)
- `deleted_at`: Soft delete timestamp
- `last_login_at`: Last successful login
- `oauth_provider`, `oauth_provider_id`: For OAuth (not yet implemented)

## Future Enhancements

### Not Yet Implemented:
1. **OAuth/Google Authentication**
   - Schema ready (`oauth_provider`, `oauth_provider_id` fields exist)
   - Need handler endpoint and service method

2. **Email Notifications**
   - Password reset token delivery
   - Deletion reminder emails
   - Job tasks exist but not wired to email client

3. **Webhook Replay Protection**
   - Could add timestamp verification (`t=` parsing)
   - Configurable time window (e.g., 5 minutes)

4. **Additional Unit Tests**
   - Expired token handling
   - Invalid signature rejection
   - Edge cases for deletion cancellation

## Configuration Example

```env
# Auth Configuration
AUTH_PASSWORD_RESET_TTL=3600        # 1 hour
AUTH_DELETION_DEFAULT_TTL=604800    # 1 week
AUTH_WEBHOOK_SIGNING_SECRET=whsec_xxxxxxxxxxxxx
```

## Running Tests

```bash
# Run all auth tests
cd apps/backend
go test ./internal/service ./internal/handler -v -timeout 5m

# Run with verbose test output
VERBOSE_TEST=1 go test ./internal/service ./internal/handler -v

# Run specific test
go test ./internal/service -run TestRegisterLoginResetAndScheduleDeletion -v
```

## Deployment Checklist

- [ ] Set `AUTH_WEBHOOK_SIGNING_SECRET` from Clerk dashboard
- [ ] Configure `AUTH_PASSWORD_RESET_TTL` (default: 1 hour)
- [ ] Configure `AUTH_DELETION_DEFAULT_TTL` (default: 1 week)
- [ ] Ensure Redis is running for Asynq job queue
- [ ] Run database migrations (0 → 3)
- [ ] Test webhook endpoint with Clerk test events
- [ ] Verify deletion worker processes jobs correctly
