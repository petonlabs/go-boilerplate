package handler

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"

	svc "github.com/petonlabs/go-boilerplate/internal/service"
	testhelpers "github.com/petonlabs/go-boilerplate/internal/testing"
)

func TestClerkWebhookSignatureValid(t *testing.T) {
	_, testServer, cleanup := testhelpers.SetupTest(t)
	defer cleanup()

	// ensure signing secret present in config for test
	testServer.Config.Auth.WebhookSigningSecret = "testsecret"

	// prepare payload
	payload := map[string]any{"data": map[string]any{"id": "user_1", "external_id": "ext_1"}}
	b, _ := json.Marshal(payload)

	mac := hmac.New(sha256.New, []byte(testServer.Config.Auth.WebhookSigningSecret))
	mac.Write(b)
	sig := hex.EncodeToString(mac.Sum(nil))

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/webhooks/clerk", bytes.NewReader(b))
	req.Header.Set("Svix-Signature", "v1="+sig)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := NewWebhookHandler(testServer, &svc.Services{Auth: svc.NewAuthService(testServer)})
	err := h.HandleClerkWebhook(c)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, rec.Code)
}
