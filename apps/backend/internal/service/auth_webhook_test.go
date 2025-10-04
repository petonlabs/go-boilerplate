package service_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	svc "github.com/petonlabs/go-boilerplate/internal/service"
	testhelpers "github.com/petonlabs/go-boilerplate/internal/testing"
)

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

	err := authSvc.SyncUser(ctx, clerkID, externalID, firstName, lastName, imageURL, raw)
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
