package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	svc "github.com/petonlabs/go-boilerplate/internal/service"
	testhelpers "github.com/petonlabs/go-boilerplate/internal/testing"
)

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
	newPass := "n3wpass"
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

	// Wait longer than original schedule to ensure worker would have run
	time.Sleep(3 * time.Second)

	// Query user to ensure deleted_at is still nil
	var deletedAt *time.Time
	err = testDB.Pool.QueryRow(ctx, `SELECT deleted_at FROM users WHERE id::text=$1`, id).Scan(&deletedAt)
	require.NoError(t, err)
	require.Nil(t, deletedAt)
}
