package handler

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"

	svc "github.com/petonlabs/go-boilerplate/internal/service"
	testhelpers "github.com/petonlabs/go-boilerplate/internal/testing"
)

func TestRequestPasswordReset_ProdDoesNotReturnToken(t *testing.T) {
	_, testServer, cleanup := testhelpers.SetupTest(t)
	defer cleanup()

	services, err := svc.NewServices(testServer, nil)
	require.NoError(t, err)
	authSvc := services.Auth
	email := "prod@example.com"
	userID, err := authSvc.RegisterUser(context.Background(), email, "password123")
	require.NoError(t, err)
	require.NotEmpty(t, userID)

	testServer.Config.Primary.Env = "production"

	payload := map[string]string{"email": email}
	b, _ := json.Marshal(payload)
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/auth/password/request", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := NewAuthHandler(testServer, services)
	require.NoError(t, h.RequestPasswordReset(c))
	require.Equal(t, http.StatusNoContent, rec.Code)

	var token sql.NullString
	err = testServer.DB.Pool.QueryRow(context.Background(), `SELECT password_reset_token FROM users WHERE email=$1`, email).Scan(&token)
	require.NoError(t, err)
	require.True(t, token.Valid)
}

func TestClerkWebhookSignatures(t *testing.T) {
	scenarios := []struct {
		name      string
		tsOffset  time.Duration
		tsInvalid bool
		wantCode  int
	}{
		{"valid", 0, false, http.StatusOK},
		{"expired", -1 * time.Hour, false, http.StatusUnauthorized},
		{"invalid_ts", 0, true, http.StatusUnauthorized},
		{"future", 1 * time.Hour, false, http.StatusUnauthorized},
	}

	for _, sc := range scenarios {
		t.Run(sc.name, func(t *testing.T) {
			_, testServer, cleanup := testhelpers.SetupTest(t)
			defer cleanup()

			testServer.Config.Auth.WebhookSigningSecret = "testsecret"

			payload := map[string]any{"data": map[string]any{"id": "user_1", "external_id": "ext_1"}}
			b, _ := json.Marshal(payload)

			svixID := uuid.New().String()
			var svixTs string
			if sc.tsInvalid {
				svixTs = "not-a-timestamp"
			} else {
				svixTs = strconv.FormatInt(time.Now().Add(sc.tsOffset).Unix(), 10)
			}

			mac := hmac.New(sha256.New, []byte(testServer.Config.Auth.WebhookSigningSecret))
			mac.Write([]byte(svixID + "." + svixTs + "."))
			mac.Write(b)
			signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))

			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/webhooks/clerk", bytes.NewReader(b))
			req.Header.Set("Svix-Id", svixID)
			req.Header.Set("Svix-Timestamp", svixTs)
			req.Header.Set("Svix-Signature", "v1,"+signature)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			h := NewWebhookHandler(testServer, &svc.Services{Auth: svc.NewAuthService(testServer)})
			require.NoError(t, h.HandleClerkWebhook(c))
			require.Equal(t, sc.wantCode, rec.Code)
		})
	}
}
