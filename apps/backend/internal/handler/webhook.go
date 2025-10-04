package handler

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/petonlabs/go-boilerplate/internal/middleware"
	"github.com/petonlabs/go-boilerplate/internal/server"
	"github.com/petonlabs/go-boilerplate/internal/service"
)

type WebhookHandler struct {
	Handler
}

func NewWebhookHandler(s *server.Server, services *service.Services) *WebhookHandler {
	return &WebhookHandler{Handler: NewHandler(s, services)}
}

// ClerkWebhookPayload is a minimal shape used for syncing user data
type ClerkWebhookPayload struct {
	Data map[string]any `json:"data"`
	Type string         `json:"type"`
}

func (h *WebhookHandler) HandleClerkWebhook(c echo.Context) error {
	logger := middleware.GetLogger(c).With().Str("operation", "clerk_webhook").Logger()
	// Read raw body for signature verification and later storage
	req := c.Request()
	bodyBytes, err := io.ReadAll(req.Body)
	if err != nil {
		logger.Error().Err(err).Msg("failed to read request body")
		return c.NoContent(http.StatusBadRequest)
	}

	// restore Body so Echo or downstream can read it if needed
	req.Body = io.NopCloser(bytes.NewReader(bodyBytes))

	// Verify signature if configured
	var signingSecret = ""
	if h.server != nil && h.server.Config != nil && h.server.Config.Auth.WebhookSigningSecret != "" {
		signingSecret = h.server.Config.Auth.WebhookSigningSecret
	}
	if signingSecret != "" {
		// Clerk uses Svix style signature header; try common headers
		sig := c.Request().Header.Get("Svix-Signature")
		if sig == "" {
			sig = c.Request().Header.Get("Clerk-Signature")
		}
		if sig == "" {
			logger.Warn().Msg("no webhook signature provided")
			return c.NoContent(http.StatusUnauthorized)
		}

		// Svix/Clerk signature header usually contains fields like: t=..., v1=hex
		// extract v1 value and compare to computed HMAC SHA256 hex signature
		var sigV1 string
		parts := strings.Split(sig, ",")
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if strings.HasPrefix(p, "v1=") {
				sigV1 = strings.TrimPrefix(p, "v1=")
				break
			}
		}
		if sigV1 == "" {
			// maybe the header is just the hex signature
			sigV1 = sig
		}

		mac := hmac.New(sha256.New, []byte(signingSecret))
		mac.Write(bodyBytes)
		expected := hex.EncodeToString(mac.Sum(nil))
		if !hmac.Equal([]byte(expected), []byte(sigV1)) {
			logger.Warn().Msg("webhook signature mismatch")
			return c.NoContent(http.StatusUnauthorized)
		}
	}

	var payload ClerkWebhookPayload
	if err := json.Unmarshal(bodyBytes, &payload); err != nil {
		logger.Error().Err(err).Msg("failed to parse webhook payload")
		return c.NoContent(http.StatusBadRequest)
	}

	// Extract a few known fields safely
	data := payload.Data
	externalID, _ := data["external_id"].(string)
	clerkID, _ := data["id"].(string)
	firstName, _ := data["first_name"].(string)
	lastName, _ := data["last_name"].(string)
	imageURL, _ := data["image_url"].(string)

	// Marshal the raw data into jsonb for storage
	rawJSON, _ := json.Marshal(data)

	// upsert user via service
	if h.services == nil || h.services.Auth == nil {
		logger.Error().Msg("auth service not available")
		return c.NoContent(http.StatusInternalServerError)
	}

	if err := h.services.Auth.SyncUser(context.Background(), clerkID, externalID, firstName, lastName, imageURL, rawJSON); err != nil {
		logger.Error().Err(err).Msg("failed to sync user from webhook")
		return c.NoContent(http.StatusInternalServerError)
	}

	return c.NoContent(http.StatusOK)
}
