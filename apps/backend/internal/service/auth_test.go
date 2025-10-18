//go:build integration
// +build integration

package service_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	svc "github.com/petonlabs/go-boilerplate/internal/service"
	testhelpers "github.com/petonlabs/go-boilerplate/internal/testhelpers"
)

func TestMain(m *testing.M) {
	// Setup shared container once for all tests in this package
	if err := testhelpers.SetupSharedContainer(); err != nil {
		panic(err)
	}

	// Run tests
	code := m.Run()

	// Cleanup
	testhelpers.CleanupSharedContainer()

	os.Exit(code)
}

func TestRegisterLoginResetAndScheduleDeletion(t *testing.T) {
	testDB, testServer, cleanup := testhelpers.SetupTest(t)
	defer cleanup()

	authSvc := svc.NewAuthService(testServer)

	ctx := context.Background()
	email := "bob@example.com"
	password := "s3cret"

	id, err := authSvc.RegisterUser(ctx, email, password)
	require.NoError(t, err)
	require.NotEmpty(t, id)

	// Login should succeed
	gotID, err := authSvc.Login(ctx, email, password)
	require.NoError(t, err)
	require.Equal(t, id, gotID)

	// Request password reset
	token, err := authSvc.RequestPasswordReset(ctx, email, 1*time.Hour)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	// Reset the password
	newPass := "N3wpassw1"
	err = authSvc.ResetPassword(ctx, token, newPass)
	require.NoError(t, err)

	// Login with new password
	gotID2, err := authSvc.Login(ctx, email, newPass)
	require.NoError(t, err)
	require.Equal(t, id, gotID2)

	// Schedule deletion in 1 second and wait
	err = authSvc.ScheduleDeletion(ctx, id, 2*time.Second)
	require.NoError(t, err)

	// Cancel deletion before it executes
	_, err = testDB.Pool.Exec(ctx, `UPDATE users SET deletion_scheduled_at = NULL WHERE id::text = $1`, id)
	require.NoError(t, err)

	// Poll for up to 5s to ensure any background worker would have run.
	// Use short polling intervals and a query timeout for robustness.
	deadline := time.Now().Add(5 * time.Second)
	var deletedAt *time.Time
	for {
		qCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		err = testDB.Pool.QueryRow(qCtx, `SELECT deleted_at FROM users WHERE id::text=$1`, id).Scan(&deletedAt)
		cancel()
		require.NoError(t, err)
		if deletedAt != nil {
			t.Fatalf("user was deleted unexpectedly: deleted_at=%v", deletedAt)
		}
		if time.Now().After(deadline) {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func TestLoginInvalidCredentialsWhenUserMissing(t *testing.T) {
	_, testServer, cleanup := testhelpers.SetupTest(t)
	defer cleanup()

	authSvc := svc.NewAuthService(testServer)

	ctx := context.Background()
	_, err := authSvc.Login(ctx, "missing@example.com", "irrelevant")
	require.ErrorIs(t, err, svc.ErrInvalidCredentials)
}

// This integration-style test runs against the test DB created by testhelpers
// and asserts that RequestPasswordReset stores an HMAC digest in the DB and
// ResetPassword succeeds only when given the original raw token.
func TestRequestPasswordReset_StoresHMACAndResetSucceeds(t *testing.T) {
	testDB, testServer, cleanup := testhelpers.SetupTest(t)
	defer cleanup()

	// ensure config has a token HMAC secret for deterministic behavior in tests
	cfg := testServer.GetConfig()
	if cfg == nil {
		t.Fatal("test server config is nil")
	}
	cfg.Auth.TokenHMACSecret = "test-secret-1"
	testServer.SetConfig(cfg)

	authSvc := svc.NewAuthService(testServer)

	ctx := context.Background()
	// create a user
	email := "hmac-test@example.com"
	id, err := authSvc.RegisterUser(ctx, email, "Password1")
	require.NoError(t, err)
	require.NotEmpty(t, id)

	// request reset
	token, err := authSvc.RequestPasswordReset(ctx, email, 1*time.Hour)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	// read stored digest from DB
	var storedDigest string
	err = testDB.Pool.QueryRow(ctx, `SELECT password_reset_token FROM users WHERE email=$1`, email).Scan(&storedDigest)
	require.NoError(t, err)
	require.NotEmpty(t, storedDigest)

	// ensure storedDigest is not equal to the raw token (i.e., it's a digest)
	require.NotEqual(t, token, storedDigest)

	// Reset with wrong token should fail
	err = authSvc.ResetPassword(ctx, "wrongtoken", "NewPass1")
	require.Error(t, err)

	// Reset with correct raw token should succeed
	err = authSvc.ResetPassword(ctx, token, "NewPass1")
	require.NoError(t, err)
}

func TestRotateTokenHMACSecrets(t *testing.T) {
	_, testServer, cleanup := testhelpers.SetupTest(t)
	defer cleanup()

	// initial secret
	cfg := testServer.GetConfig()
	if cfg == nil {
		t.Fatal("test server config is nil")
	}
	cfg.Auth.TokenHMACSecret = "secret-old"
	testServer.SetConfig(cfg)
	authSvc := svc.NewAuthService(testServer)

	ctx := context.Background()
	email := "rotate-test@example.com"
	_, err := authSvc.RegisterUser(ctx, email, "Password1")
	require.NoError(t, err)

	// create token with old secret
	tokenOld, err := authSvc.RequestPasswordReset(ctx, email, 1*time.Hour)
	require.NoError(t, err)
	require.NotEmpty(t, tokenOld)

	// rotate to new secret
	err = authSvc.RotateTokenHMACSecrets("secret-new,secret-old", "test")
	require.NoError(t, err)

	// Reset using old token must succeed because Reset attempts all secrets.
	// Do this before creating another token (which would overwrite the stored token)
	err = authSvc.ResetPassword(ctx, tokenOld, "NewPass1")
	require.NoError(t, err)

	// create token with new secret (first in list)
	tokenNew, err := authSvc.RequestPasswordReset(ctx, email, 1*time.Hour)
	require.NoError(t, err)
	require.NotEmpty(t, tokenNew)
	require.NotEqual(t, tokenOld, tokenNew)

	// Now the new token should also work
	require.NoError(t, authSvc.ResetPassword(ctx, tokenNew, "NewPass2"))
}

func TestSyncUserUpserts(t *testing.T) {
	testDB, testServer, cleanup := testhelpers.SetupTest(t)
	defer cleanup()

	// Create AuthService
	authSvc := svc.NewAuthService(testServer)

	ctx := context.Background()
	clerkID := "user_123"
	externalID := "ext_123"
	firstName := "Alice"
	lastName := "Tester"
	imageURL := "https://example.com/avatar.jpg"
	raw := []byte(`{"id":"user_123","external_id":"ext_123","first_name":"Alice"}`)

	email := "alice@example.com"
	err := authSvc.SyncUser(ctx, clerkID, externalID, email, firstName, lastName, imageURL, raw)
	require.NoError(t, err)

	// Query DB for the user
	var gotClerkID string
	var gotExternalID string
	var gotFirstName string
	var gotLastName string
	var gotImage string

	row := testDB.Pool.QueryRow(ctx, `SELECT clerk_id, external_id, first_name, last_name, image_url FROM users WHERE external_id=$1`, externalID)
	err = row.Scan(&gotClerkID, &gotExternalID, &gotFirstName, &gotLastName, &gotImage)
	require.NoError(t, err)
	require.Equal(t, clerkID, gotClerkID)
	require.Equal(t, externalID, gotExternalID)
	require.Equal(t, firstName, gotFirstName)
	require.Equal(t, lastName, gotLastName)
	require.Equal(t, imageURL, gotImage)
}

func TestSyncUserUpserts_ClerkIDOnly(t *testing.T) {
	testDB, testServer, cleanup := testhelpers.SetupTest(t)
	defer cleanup()

	authSvc := svc.NewAuthService(testServer)

	ctx := context.Background()
	clerkID := "user_only_123"
	externalID := ""
	email := "bob@example.com"
	firstName := "Bob"
	lastName := "NoExt"
	imageURL := "https://example.com/bob.jpg"
	raw := []byte(`{"id":"user_only_123","first_name":"Bob"}`)

	err := authSvc.SyncUser(ctx, clerkID, externalID, email, firstName, lastName, imageURL, raw)
	require.NoError(t, err)

	var gotClerkID string
	var gotExternalID string
	var gotFirstName string
	var gotLastName string
	var gotImage string

	row := testDB.Pool.QueryRow(ctx, `SELECT clerk_id, external_id, first_name, last_name, image_url FROM users WHERE clerk_id=$1`, clerkID)
	err = row.Scan(&gotClerkID, &gotExternalID, &gotFirstName, &gotLastName, &gotImage)
	require.NoError(t, err)
	require.Equal(t, clerkID, gotClerkID)
	require.Equal(t, externalID, gotExternalID)
	require.Equal(t, firstName, gotFirstName)
	require.Equal(t, lastName, gotLastName)
	require.Equal(t, imageURL, gotImage)
}
