package handler

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"

	svc "github.com/petonlabs/go-boilerplate/internal/service"
	testhelpers "github.com/petonlabs/go-boilerplate/internal/testing"
)

func TestAdminRotateSecretsEndpoint(t *testing.T) {
	_, testServer, cleanup := testhelpers.SetupTest(t)
	defer cleanup()

	testServer.Config.Auth.AdminToken = "admintoken"

	services, err := svc.NewServices(testServer, nil)
	require.NoError(t, err)
	h := NewHandlers(testServer, services)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/admin/rotate-secrets", bytes.NewReader([]byte(`{"secrets":"s1,s2"}`)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Admin-Token", "admintoken")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	require.NoError(t, h.Admin.RotateSecrets(c))
	require.Equal(t, http.StatusOK, rec.Code)
}

func TestAdminRotateSecretsEndpoint_Unauthorized(t *testing.T) {
	_, testServer, cleanup := testhelpers.SetupTest(t)
	defer cleanup()

	testServer.Config.Auth.AdminToken = "admintoken"

	services, err := svc.NewServices(testServer, nil)
	require.NoError(t, err)
	h := NewHandlers(testServer, services)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/admin/rotate-secrets", bytes.NewReader([]byte(`{"secrets":"s1,s2"}`)))
	req.Header.Set("Content-Type", "application/json")
	// Wrong admin token
	req.Header.Set("X-Admin-Token", "wrongtoken")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err = h.Admin.RotateSecrets(c)
	require.Error(t, err)
	var he *echo.HTTPError
	require.True(t, errors.As(err, &he), "expected echo.HTTPError for unauthorized response")
	require.Equal(t, http.StatusUnauthorized, he.Code)
}

func TestAdminRotateSecretsEndpoint_MissingToken(t *testing.T) {
	_, testServer, cleanup := testhelpers.SetupTest(t)
	defer cleanup()

	testServer.Config.Auth.AdminToken = "admintoken"

	services, err := svc.NewServices(testServer, nil)
	require.NoError(t, err)
	h := NewHandlers(testServer, services)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/admin/rotate-secrets", bytes.NewReader([]byte(`{"secrets":"s1,s2"}`)))
	req.Header.Set("Content-Type", "application/json")
	// No X-Admin-Token header set
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err = h.Admin.RotateSecrets(c)
	require.Error(t, err)
	var he2 *echo.HTTPError
	require.True(t, errors.As(err, &he2), "expected echo.HTTPError for unauthorized response")
	require.Equal(t, http.StatusUnauthorized, he2.Code)
}
