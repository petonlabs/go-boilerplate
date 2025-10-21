package service

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/rs/zerolog"

	"golang.org/x/crypto/bcrypt"

	"github.com/petonlabs/go-boilerplate/internal/lib/job"
	"github.com/petonlabs/go-boilerplate/internal/server"

	"github.com/clerk/clerk-sdk-go/v2"
)

type AuthService struct {
	server *server.Server
	// tokenSecrets holds the active and rotated HMAC secrets.
	// Access must be done under secretsMu.
	secretsMu    sync.RWMutex
	tokenSecrets []string
}

// ErrInvalidCredentials is returned when login fails due to invalid email/password
var ErrInvalidCredentials = errors.New("invalid credentials")
var (
	ErrInvalidPasswordResetToken = errors.New("invalid password reset token")
	ErrExpiredPasswordResetToken = errors.New("password reset token expired")
	ErrUserNotFound              = errors.New("user not found or already deleted")
	ErrPasswordValidation        = errors.New("password validation failed")
)

func NewAuthService(s *server.Server) *AuthService {
	a := &AuthService{server: s}
	if s != nil {
		if cfg := s.GetConfig(); cfg != nil {
			clerk.SetKey(cfg.Auth.SecretKey)
		}
	}
	// Initialize token secrets from config so reads can use the in-memory slice.
	if s != nil {
		if cfg := s.GetConfig(); cfg != nil {
			initial := parseTokenSecrets(cfg.Auth.TokenHMACSecret, cfg.Auth.SecretKey)
			if len(initial) == 0 {
				initial = []string{}
			}
			a.secretsMu.Lock()
			a.tokenSecrets = initial
			a.secretsMu.Unlock()
		} else {
			a.secretsMu.Lock()
			a.tokenSecrets = []string{}
			a.secretsMu.Unlock()
		}
	}
	return a
}

// SyncUser upserts a user record from Clerk webhook data
// email parameter should be provided from the webhook payload when available.
func (a *AuthService) SyncUser(ctx context.Context, clerkID, externalID, email, firstName, lastName, imageURL, role string, rawPayload []byte) error {
	if a.server == nil || a.server.DB == nil || a.server.DB.Pool == nil {
		return fmt.Errorf("database not initialized")
	}

	// If both identifiers are missing, nothing to do.
	if externalID == "" && clerkID == "" {
		return nil
	}

	// Upsert user: some databases in tests or older deployments may not have
	// the composite constraint the original implementation expected. To be
	// resilient we perform an insert that does nothing on conflict and then
	// an update by matching either external_id or clerk_id (case-insensitive
	// where appropriate). This avoids relying on a specific constraint
	// name existing in the database schema.

	// Try insert; ON CONFLICT DO NOTHING prevents unique-violation errors
	// from bubbling up if one of the single-column unique indexes exists.
	insertQuery := `INSERT INTO users (email, clerk_id, external_id, first_name, last_name, image_url, role, raw_payload, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, now()) ON CONFLICT DO NOTHING;`
	if _, err := a.server.DB.Pool.Exec(ctx, insertQuery, email, clerkID, externalID, firstName, lastName, imageURL, role, rawPayload); err != nil {
		return err
	}

	// Update any existing row that matches by external_id or clerk_id. Use
	// lower(...) comparisons to match the behavior of the unique indexes
	// created by migrations which use lower(...).
	updateQuery := `UPDATE users SET
		email = $1,
		clerk_id = COALESCE(NULLIF($2, ''), clerk_id),
		external_id = COALESCE(NULLIF($3, ''), external_id),
		first_name = $4,
		last_name = $5,
		image_url = $6,
		role = $7,
		raw_payload = $8
	  WHERE (external_id IS NOT NULL AND external_id <> '' AND lower(external_id) = lower($3))
		 OR (clerk_id IS NOT NULL AND clerk_id <> '' AND lower(clerk_id) = lower($2));`

	if _, err := a.server.DB.Pool.Exec(ctx, updateQuery, email, clerkID, externalID, firstName, lastName, imageURL, role, rawPayload); err != nil {
		return err
	}

	return nil
}

// RegisterUser registers a new user with email and password
func (a *AuthService) RegisterUser(ctx context.Context, email, password string) (string, error) {
	if a.server == nil || a.server.DB == nil || a.server.DB.Pool == nil {
		return "", fmt.Errorf("database not initialized")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	var id string
	query := `INSERT INTO users (email, password_hash, created_at) VALUES ($1, $2, now()) RETURNING id::text`
	err = a.server.DB.Pool.QueryRow(ctx, query, email, string(hashed)).Scan(&id)
	if err != nil {
		return "", err
	}
	return id, nil
}

// Login verifies email and password and updates last_login_at
func (a *AuthService) Login(ctx context.Context, email, password string) (string, error) {
	if a.server == nil || a.server.DB == nil || a.server.DB.Pool == nil {
		return "", fmt.Errorf("database not initialized")
	}

	var id string
	var hash string
	err := a.server.DB.Pool.QueryRow(ctx, `SELECT id::text, password_hash FROM users WHERE email=$1 AND deleted_at IS NULL`, email).Scan(&id, &hash)
	if err != nil {
		// avoid revealing whether the user exists
		return "", ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		return "", ErrInvalidCredentials
	}

	// update last_login_at
	if _, err := a.server.DB.Pool.Exec(ctx, `UPDATE users SET last_login_at = now() WHERE id = $1`, id); err != nil {
		// Log the error but don't fail login to avoid impacting UX
		if a.server != nil && a.server.Logger != nil {
			a.server.Logger.Error().Err(err).Str("user_id", id).Msg("failed to update last_login_at")
		} else {
			// Fallback: create a temporary logger and log structured message
			tmp := zerolog.New(zerolog.NewConsoleWriter()).With().Timestamp().Logger()
			tmp.Error().Err(err).Str("user_id", id).Msg("failed to update last_login_at")
		}
	}
	return id, nil
}

// RequestPasswordReset creates a reset token and sets expiry
func (a *AuthService) RequestPasswordReset(ctx context.Context, email string, ttl time.Duration) (string, error) {
	tokenBytes := make([]byte, 16)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", err
	}
	token := hex.EncodeToString(tokenBytes)
	expiry := time.Now().Add(ttl)
	if a.server == nil || a.server.DB == nil || a.server.DB.Pool == nil {
		return "", fmt.Errorf("database not initialized")
	}

	// Compute HMAC-SHA256 of the token using the current configured secret to avoid storing raw tokens.
	// Load the current secrets under a read lock; fall back to config parsing if not initialized.
	var currentSecret string
	a.secretsMu.RLock()
	if len(a.tokenSecrets) > 0 {
		currentSecret = a.tokenSecrets[0]
	}
	a.secretsMu.RUnlock()
	if currentSecret == "" {
		// fallback: parse from config (should be rare)
		if cfg := a.server.GetConfig(); cfg != nil {
			parsed := parseTokenSecrets(cfg.Auth.TokenHMACSecret, cfg.Auth.SecretKey)
			if len(parsed) == 0 {
				return "", fmt.Errorf("no token HMAC secret configured")
			}
			currentSecret = parsed[0]
		} else {
			return "", fmt.Errorf("no token HMAC secret configured")
		}

	}
	mac := hmac.New(sha256.New, []byte(currentSecret))
	mac.Write([]byte(token))
	hashedToken := hex.EncodeToString(mac.Sum(nil))

	ct, err := a.server.DB.Pool.Exec(ctx, `UPDATE users SET password_reset_token=$1, password_reset_expires=$2 WHERE email=$3`, hashedToken, expiry, email)
	if err != nil {
		return "", fmt.Errorf("failed to set password reset token for email %s: %w", email, err)
	}
	if ct.RowsAffected() == 0 {
		// No rows updated means no user with that email (or user deleted)
		return "", sql.ErrNoRows
	}
	return token, nil
}

// ResetPassword verifies token and updates password
func (a *AuthService) ResetPassword(ctx context.Context, token, newPassword string) error {
	// Ensure DB is initialized
	if a.server == nil || a.server.DB == nil || a.server.DB.Pool == nil {
		return fmt.Errorf("database not initialized")
	}

	// Validate token is non-empty
	if token == "" {
		return ErrInvalidPasswordResetToken
	}

	var id string
	var exp sql.NullTime
	// Compute HMAC-SHA256 digests for the provided token using all configured secrets (supports rotation).
	a.secretsMu.RLock()
	localSecrets := make([]string, len(a.tokenSecrets))
	copy(localSecrets, a.tokenSecrets)
	a.secretsMu.RUnlock()
	digests := computeTokenDigests(token, localSecrets)
	// If no secrets were loaded from the in-memory store (unlikely), fallback to parsing from config.
	if len(digests) == 0 {
		if a.server == nil {
			return ErrInvalidPasswordResetToken
		}
		if cfg := a.server.GetConfig(); cfg != nil {
			secrets := parseTokenSecrets(cfg.Auth.TokenHMACSecret, cfg.Auth.SecretKey)
			if len(secrets) == 0 {
				return ErrInvalidPasswordResetToken
			}
			digests = computeTokenDigests(token, secrets)
		} else {
			return ErrInvalidPasswordResetToken
		}
	}

	// Build a parameterized IN clause to find the user by any of the digests
	placeholders := make([]string, len(digests))
	args := make([]interface{}, len(digests))
	for i, d := range digests {
		placeholders[i] = "$" + fmt.Sprint(i+1)
		args[i] = d
	}
	query := `SELECT id::text, password_reset_expires FROM users WHERE password_reset_token IN (` + strings.Join(placeholders, ",") + `) AND deleted_at IS NULL`
	// Only consider tokens for non-deleted users
	err := a.server.DB.Pool.QueryRow(ctx, query, args...).Scan(&id, &exp)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrInvalidPasswordResetToken
		}
		return err
	}
	if !exp.Valid {
		return ErrInvalidPasswordResetToken
	}
	if time.Now().After(exp.Time) {
		return ErrExpiredPasswordResetToken
	}

	// Validate password with shared helper (min/max length and character classes)
	if err := validatePassword(newPassword); err != nil {
		return err
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	// Ensure we only update non-deleted users and return an error if nothing updated
	ct, err := a.server.DB.Pool.Exec(ctx, `UPDATE users SET password_hash=$1, password_reset_token=NULL, password_reset_expires=NULL WHERE id=$2 AND deleted_at IS NULL`, string(hashed), id)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrUserNotFound
	}
	return nil
}

// ScheduleDeletion marks a user for deletion at now + ttl and enqueues a deletion job
func (a *AuthService) ScheduleDeletion(ctx context.Context, userID string, ttl time.Duration) error {
	when := time.Now().Add(ttl)
	_, err := a.server.DB.Pool.Exec(ctx, `UPDATE users SET deletion_scheduled_at=$1 WHERE id::text=$2`, when, userID)
	if err != nil {
		return err
	}
	// Enqueue job to delete (the job worker should check deletion_scheduled_at)
	if a.server.Job != nil && a.server.Job.Client != nil {
		task, err := job.NewUserDeleteTask(userID)
		if err == nil {
			_, _ = a.server.Job.Client.Enqueue(task)
		}
	}
	return nil
}

// CancelDeletion clears deletion_scheduled_at for a user, interrupting a scheduled deletion
func (a *AuthService) CancelDeletion(ctx context.Context, userID string) error {
	if a.server == nil || a.server.DB == nil || a.server.DB.Pool == nil {
		return fmt.Errorf("database not initialized")
	}
	_, err := a.server.DB.Pool.Exec(ctx, `UPDATE users SET deletion_scheduled_at = NULL WHERE id::text = $1`, userID)
	return err
}

// validatePassword enforces min/max length and character class requirements.
func validatePassword(pw string) error {
	if len(pw) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}
	if len(pw) > 128 {
		return fmt.Errorf("password must not exceed 128 characters")
	}
	var hasUpper, hasLower, hasDigit bool
	for _, r := range pw {
		if unicode.IsUpper(r) {
			hasUpper = true
		}
		if unicode.IsLower(r) {
			hasLower = true
		}
		if unicode.IsDigit(r) {
			hasDigit = true
		}
	}
	if !hasUpper || !hasLower || !hasDigit {
		return fmt.Errorf("password must include upper and lower case letters and a digit")
	}
	return nil
}

// computeTokenDigests computes HMAC-SHA256 hex-encoded digests for the provided token
// using the provided secrets slice. Returns an empty slice if secrets is empty.
func computeTokenDigests(token string, secrets []string) []string {
	if len(secrets) == 0 {
		return []string{}
	}
	digests := make([]string, 0, len(secrets))
	for _, s := range secrets {
		mac := hmac.New(sha256.New, []byte(s))
		mac.Write([]byte(token))
		digests = append(digests, hex.EncodeToString(mac.Sum(nil)))
	}
	return digests
}

// parseTokenSecrets returns a slice of secrets to try for HMAC. Accepts an explicit
// tokenHMACSecret string which may include multiple secrets separated by ',' or '|'.
// If tokenHMACSecret is empty, fall back to the mainSecret as the single value.
func parseTokenSecrets(tokenHMACSecret, mainSecret string) []string {
	// If no explicit tokenHMACSecret is provided, fall back to mainSecret only if it is non-empty.
	if strings.TrimSpace(tokenHMACSecret) == "" {
		if strings.TrimSpace(mainSecret) == "" {
			return nil
		}
		return []string{mainSecret}
	}
	normalized := strings.ReplaceAll(tokenHMACSecret, "|", ",")
	parts := strings.Split(normalized, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		out = append(out, p)
	}
	// If after filtering there are no non-empty parts, don't silently fall back to mainSecret;
	// return nil to indicate configuration is effectively empty.
	if len(out) == 0 {
		return nil
	}
	return out
}

// RotateTokenHMACSecrets atomically replaces the configured token HMAC secrets.
// Accepts a comma or pipe separated list where the first value is the active
// secret used for newly created tokens. This is a lightweight in-memory helper
// intended for admin tooling or tests. In production you should persist and
// distribute secrets securely (vault, env, k8s secret, etc.).
// RotateTokenHMACSecrets atomically replaces the configured token HMAC secrets.
// Accepts an actor string to include in the audit log (e.g. 'admin_api' or user id).
func (a *AuthService) RotateTokenHMACSecrets(newSecrets string, actor string) error {
	if a == nil || a.server == nil || a.server.GetConfig() == nil {
		return fmt.Errorf("server or config not available")
	}
	// sanitize input by trimming spaces
	normalized := strings.TrimSpace(newSecrets)
	if normalized == "" {
		return fmt.Errorf("newSecrets must not be empty")
	}
	// Parse the new secrets without falling back to mainSecret.
	parsed := parseTokenSecrets(normalized, "")
	if len(parsed) == 0 {
		return fmt.Errorf("parsed secrets are empty or invalid")
	}
	// Atomically replace the in-memory secrets under write lock.
	a.secretsMu.Lock()
	a.tokenSecrets = parsed
	a.secretsMu.Unlock()
	// Persist the raw secrets string into the server config under a synchronized
	// setter so other in-process components can observe the new configuration.
	// We intentionally do not log raw secrets; log only a masked preview.
	a.server.SetTokenHMACSecret(normalized)

	if a.server.Logger != nil {
		// Build a non-sensitive summary: count and masked preview (only last 4 chars visible)
		masked := make([]string, 0, len(parsed))
		for _, s := range parsed {
			if len(s) <= 4 {
				masked = append(masked, "****")
			} else {
				tail := s[len(s)-4:]
				masked = append(masked, "****"+tail)
			}
		}
		a.server.Logger.Info().Int("secrets_count", len(parsed)).Strs("secrets_preview_masked", masked).Msg("rotated token HMAC secrets (preview)")

		// Audit entry for persistence action (actor info is best-effort; expand if available)
		a.server.Logger.Info().Str("actor", actor).Msg("persisted token HMAC secrets to server config (masked preview logged above)")
	}
	return nil
}

// GetTokenSecrets returns a copy of the currently configured token HMAC secrets.
// The returned slice is a shallow copy to avoid exposing internal state for modification.
func (a *AuthService) GetTokenSecrets() []string {
	a.secretsMu.RLock()
	defer a.secretsMu.RUnlock()
	out := make([]string, len(a.tokenSecrets))
	copy(out, a.tokenSecrets)
	return out
}
