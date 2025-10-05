package handler

import (
	"crypto/subtle"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/petonlabs/go-boilerplate/internal/middleware"
	"github.com/petonlabs/go-boilerplate/internal/server"
	"github.com/petonlabs/go-boilerplate/internal/service"
)

type AdminHandler struct{ Handler }

func NewAdminHandler(s *server.Server, services *service.Services) *AdminHandler {
	return &AdminHandler{Handler: NewHandler(s, services)}
}

type rotateReq struct {
	Secrets string `json:"secrets"`
}

// RotateSecrets rotates the token HMAC secrets. Protected by X-Admin-Token header.
func (h *AdminHandler) RotateSecrets(c echo.Context) error {
	logger := middleware.GetLogger(c).With().Str("operation", "admin_rotate_secrets").Logger()
	// Simple header-based auth for admin tooling/tests
	adminHeader := c.Request().Header.Get("X-Admin-Token")
	if h.server == nil {
		logger.Warn().Msg("admin rotate secrets unauthorized")
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}
	if cfg := h.server.GetConfig(); cfg == nil || cfg.Auth.AdminToken == "" {
		logger.Warn().Msg("admin rotate secrets unauthorized")
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	} else if subtle.ConstantTimeCompare([]byte(adminHeader), []byte(cfg.Auth.AdminToken)) != 1 {
		// Use constant-time comparison to avoid timing attacks
		logger.Warn().Msg("admin rotate secrets unauthorized")
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}
	var req rotateReq
	if err := c.Bind(&req); err != nil {
		logger.Error().Err(err).Msg("invalid rotate payload")
		return echo.NewHTTPError(http.StatusBadRequest, "invalid rotate payload")
	}
	// Validate that the client provided a non-empty secrets string.
	if strings.TrimSpace(req.Secrets) == "" {
		logger.Error().Msg("rotate payload missing secrets")
		return echo.NewHTTPError(http.StatusBadRequest, "missing secrets")
	}
	if err := h.services.Auth.RotateTokenHMACSecrets(req.Secrets, "admin_api"); err != nil {
		logger.Error().Err(err).Msg("failed to rotate secrets")
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	// Log an audit entry that the secrets were rotated and persisted.
	logger.Info().Str("actor", "admin_api").Msg("admin rotated token HMAC secrets and persisted to config (masked preview logged by service)")
	return c.NoContent(http.StatusOK)
}
