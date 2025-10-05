package handler

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/petonlabs/go-boilerplate/internal/middleware"
	"github.com/petonlabs/go-boilerplate/internal/server"
	"github.com/petonlabs/go-boilerplate/internal/service"
)

type WebhookHandler struct {
	Handler
}

// DefaultWebhookToleranceSec is the default allowed clock skew (in seconds)
// for Svix webhook timestamps. Default = 5 minutes.
const DefaultWebhookToleranceSec = 300

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
	if h.server != nil {
		if cfg := h.server.GetConfig(); cfg != nil && cfg.Auth.WebhookSigningSecret != "" {
			signingSecret = cfg.Auth.WebhookSigningSecret
		}
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

		// Svix/Clerk signature header may be the Svix style which includes
		// Svix-Id and Svix-Timestamp and a base64-encoded v1 signature computed
		// over the string: "<svix-id>.<svix-timestamp>.<raw_body>". Fallback to
		// legacy behavior (HMAC over body with hex-encoded signature) if the
		// Svix headers are not present.

		svixID := c.Request().Header.Get("Svix-Id")
		svixTs := c.Request().Header.Get("Svix-Timestamp")

		// helper to extract v1 token from signature header; supports formats
		// like "v1=<sig>", "v1,<sig>" or comma-separated list where v1 is a key
		extractV1 := func(sigHeader string) string {
			parts := strings.Split(sigHeader, ",")
			for i, p := range parts {
				p = strings.TrimSpace(p)
				if strings.HasPrefix(p, "v1=") {
					return strings.TrimPrefix(p, "v1=")
				}
				if p == "v1" && i+1 < len(parts) {
					return strings.TrimSpace(parts[i+1])
				}
				// handle case where header is simply "v1,<sig>" -> first part == "v1"
			}
			// no v1 key found; maybe header is just the signature
			return strings.TrimSpace(sigHeader)
		}

		sigV1 := extractV1(sig)

		// If we have Svix id and timestamp, validate using the Svix signing scheme
		if svixID != "" && svixTs != "" {
			// enforce replay window: parse timestamp and ensure it's within tolerance
			tolerance := DefaultWebhookToleranceSec
			if h.server != nil {
				if cfg := h.server.GetConfig(); cfg != nil && cfg.Auth.WebhookToleranceSec > 0 {
					tolerance = cfg.Auth.WebhookToleranceSec
				}
			}
			// parse timestamp
			if tsInt, err := strconv.ParseInt(svixTs, 10, 64); err == nil {
				now := time.Now().Unix()
				if tsInt > now+int64(tolerance) || tsInt < now-int64(tolerance) {
					logger.Warn().Msg("webhook timestamp outside tolerance window")
					return c.NoContent(http.StatusUnauthorized)
				}
			} else {
				logger.Warn().Err(err).Msg("invalid svix timestamp")
				return c.NoContent(http.StatusUnauthorized)
			}
			mac := hmac.New(sha256.New, []byte(signingSecret))
			mac.Write([]byte(svixID + "." + svixTs + "."))
			mac.Write(bodyBytes)
			expectedMAC := mac.Sum(nil)

			// Signature should be base64-encoded for Svix
			var givenMAC []byte
			// try base64 first
			if gm, err := base64.StdEncoding.DecodeString(sigV1); err == nil {
				givenMAC = gm
			} else if gm, err := hex.DecodeString(sigV1); err == nil {
				// fall back to hex if tests or callers provided hex
				givenMAC = gm
			} else {
				logger.Warn().Msg("webhook signature encoding invalid")
				return c.NoContent(http.StatusUnauthorized)
			}

			if !hmac.Equal(expectedMAC, givenMAC) {
				logger.Warn().Msg("webhook signature mismatch")
				return c.NoContent(http.StatusUnauthorized)
			}
		} else {
			// Legacy: compute HMAC over body and compare hex-encoded signature
			mac := hmac.New(sha256.New, []byte(signingSecret))
			mac.Write(bodyBytes)
			expected := hex.EncodeToString(mac.Sum(nil))
			if !hmac.Equal([]byte(expected), []byte(sigV1)) {
				logger.Warn().Msg("webhook signature mismatch")
				return c.NoContent(http.StatusUnauthorized)
			}
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
	email, _ := data["email"].(string)
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

	if err := h.services.Auth.SyncUser(c.Request().Context(), clerkID, externalID, email, firstName, lastName, imageURL, rawJSON); err != nil {
		logger.Error().Err(err).Msg("failed to sync user from webhook")
		return c.NoContent(http.StatusInternalServerError)
	}

	return c.NoContent(http.StatusOK)
}
