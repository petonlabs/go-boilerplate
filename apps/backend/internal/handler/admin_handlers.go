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
	if h.server == nil || h.server.Config == nil || h.server.Config.Auth.AdminToken == "" {
		logger.Warn().Msg("admin rotate secrets unauthorized")
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}
	// Use constant-time comparison to avoid timing attacks
	if subtle.ConstantTimeCompare([]byte(adminHeader), []byte(h.server.Config.Auth.AdminToken)) != 1 {
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
	if err := h.services.Auth.RotateTokenHMACSecrets(req.Secrets); err != nil {
		logger.Error().Err(err).Msg("failed to rotate secrets")
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusOK)
}
