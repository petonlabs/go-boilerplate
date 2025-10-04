package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/petonlabs/go-boilerplate/internal/lib/job"
	"github.com/petonlabs/go-boilerplate/internal/server"

	"github.com/clerk/clerk-sdk-go/v2"
)

type AuthService struct {
	server *server.Server
}

func NewAuthService(s *server.Server) *AuthService {
	clerk.SetKey(s.Config.Auth.SecretKey)
	return &AuthService{
		server: s,
	}
}

// SyncUser upserts a user record from Clerk webhook data
func (a *AuthService) SyncUser(ctx context.Context, clerkID, externalID, firstName, lastName, imageURL string, rawPayload []byte) error {
	if a.server == nil || a.server.DB == nil || a.server.DB.Pool == nil {
		return fmt.Errorf("database not initialized")
	}

	// Upsert by external_id if provided, otherwise clerk_id
	query := `INSERT INTO users (email, clerk_id, external_id, first_name, last_name, image_url, raw_payload, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, now())
ON CONFLICT (external_id) DO UPDATE SET
  clerk_id = EXCLUDED.clerk_id,
  first_name = EXCLUDED.first_name,
  last_name = EXCLUDED.last_name,
  image_url = EXCLUDED.image_url,
  raw_payload = EXCLUDED.raw_payload;
`

	// email is optional in webhook payload; use external_id as fallback
	email := ""
	if externalID == "" && clerkID == "" {
		return nil
	}

	_, err := a.server.DB.Pool.Exec(ctx, query, email, clerkID, externalID, firstName, lastName, imageURL, rawPayload)
	return err
}

// RegisterUser registers a new user with email and password
func (a *AuthService) RegisterUser(ctx context.Context, email, password string) (string, error) {
	if a.server == nil || a.server.DB == nil || a.server.DB.Pool == nil {
		return "", fmt.Errorf("database not initialized")
	}

	// Hash password
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
		return "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		return "", fmt.Errorf("invalid credentials")
	}

	// update last_login_at
	_, _ = a.server.DB.Pool.Exec(ctx, `UPDATE users SET last_login_at = now() WHERE id = $1`, id)
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

	_, err := a.server.DB.Pool.Exec(ctx, `UPDATE users SET password_reset_token=$1, password_reset_expires=$2 WHERE email=$3`, token, expiry, email)
	if err != nil {
		return "", err
	}
	return token, nil
}

// ResetPassword verifies token and updates password
func (a *AuthService) ResetPassword(ctx context.Context, token, newPassword string) error {
	var id string
	var exp time.Time
	err := a.server.DB.Pool.QueryRow(ctx, `SELECT id::text, password_reset_expires FROM users WHERE password_reset_token=$1`, token).Scan(&id, &exp)
	if err != nil {
		return err
	}
	if time.Now().After(exp) {
		return fmt.Errorf("token expired")
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = a.server.DB.Pool.Exec(ctx, `UPDATE users SET password_hash=$1, password_reset_token=NULL, password_reset_expires=NULL WHERE id=$2`, string(hashed), id)
	return err
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
