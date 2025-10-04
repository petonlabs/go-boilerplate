package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/petonlabs/go-boilerplate/internal/middleware"
	"github.com/petonlabs/go-boilerplate/internal/server"
	"github.com/petonlabs/go-boilerplate/internal/service"
)

type AuthHandler struct {
	Handler
}

func NewAuthHandler(s *server.Server, services *service.Services) *AuthHandler {
	return &AuthHandler{Handler: NewHandler(s, services)}
}

type registerReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *AuthHandler) Register(c echo.Context) error {
	logger := middleware.GetLogger(c).With().Str("operation", "register").Logger()
	var req registerReq
	if err := c.Bind(&req); err != nil {
		logger.Error().Err(err).Msg("invalid register payload")
		return c.NoContent(http.StatusBadRequest)
	}
	id, err := h.services.Auth.RegisterUser(context.Background(), req.Email, req.Password)
	if err != nil {
		logger.Error().Err(err).Msg("failed to register user")
		return c.NoContent(http.StatusInternalServerError)
	}
	return c.JSON(http.StatusCreated, map[string]string{"id": id})
}

type loginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *AuthHandler) Login(c echo.Context) error {
	logger := middleware.GetLogger(c).With().Str("operation", "login").Logger()
	var req loginReq
	if err := c.Bind(&req); err != nil {
		logger.Error().Err(err).Msg("invalid login payload")
		return c.NoContent(http.StatusBadRequest)
	}
	id, err := h.services.Auth.Login(context.Background(), req.Email, req.Password)
	if err != nil {
		logger.Info().Err(err).Msg("authentication failed")
		return c.NoContent(http.StatusUnauthorized)
	}
	return c.JSON(http.StatusOK, map[string]string{"id": id})
}

type pwResetReq struct {
	Email string `json:"email"`
}

func (h *AuthHandler) RequestPasswordReset(c echo.Context) error {
	logger := middleware.GetLogger(c).With().Str("operation", "request_password_reset").Logger()
	var req pwResetReq
	if err := c.Bind(&req); err != nil {
		logger.Error().Err(err).Msg("invalid payload")
		return c.NoContent(http.StatusBadRequest)
	}
	token, err := h.services.Auth.RequestPasswordReset(context.Background(), req.Email, time.Duration(h.server.Config.Auth.PasswordResetTTL)*time.Second)
	if err != nil {
		logger.Error().Err(err).Msg("failed to create password reset token")
		return c.NoContent(http.StatusInternalServerError)
	}
	// In production, we'd email this token. Return it here for tests/dev.
	return c.JSON(http.StatusOK, map[string]string{"token": token})
}

type resetReq struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
}

func (h *AuthHandler) ResetPassword(c echo.Context) error {
	logger := middleware.GetLogger(c).With().Str("operation", "reset_password").Logger()
	var req resetReq
	if err := c.Bind(&req); err != nil {
		logger.Error().Err(err).Msg("invalid payload")
		return c.NoContent(http.StatusBadRequest)
	}
	if err := h.services.Auth.ResetPassword(context.Background(), req.Token, req.NewPassword); err != nil {
		logger.Error().Err(err).Msg("failed to reset password")
		return c.NoContent(http.StatusInternalServerError)
	}
	return c.NoContent(http.StatusOK)
}

type deletionReq struct {
	UserID  string `json:"user_id"`
	Seconds int64  `json:"seconds"` // optional override in seconds
}

func (h *AuthHandler) ScheduleDeletion(c echo.Context) error {
	logger := middleware.GetLogger(c).With().Str("operation", "schedule_deletion").Logger()
	var req deletionReq
	if err := c.Bind(&req); err != nil {
		logger.Error().Err(err).Msg("invalid payload")
		return c.NoContent(http.StatusBadRequest)
	}
	// Default TTL from config (in seconds)
	ttl := h.server.Config.Auth.DeletionDefaultTTL
	// If user provided seconds override, use it
	if req.Seconds > 0 {
		ttl = int(req.Seconds)
	}
	if err := h.services.Auth.ScheduleDeletion(context.Background(), req.UserID, time.Duration(ttl)*time.Second); err != nil {
		logger.Error().Err(err).Msg("failed to schedule deletion")
		return c.NoContent(http.StatusInternalServerError)
	}
	return c.NoContent(http.StatusOK)
}

type cancelReq struct {
	UserID string `json:"user_id"`
}

// CancelDeletion clears deletion_scheduled_at to interrupt a scheduled deletion
func (h *AuthHandler) CancelDeletion(c echo.Context) error {
	logger := middleware.GetLogger(c).With().Str("operation", "cancel_deletion").Logger()
	var req cancelReq
	if err := c.Bind(&req); err != nil {
		logger.Error().Err(err).Msg("invalid payload")
		return c.NoContent(http.StatusBadRequest)
	}
	_, err := h.server.DB.Pool.Exec(context.Background(), `UPDATE users SET deletion_scheduled_at = NULL WHERE id::text = $1`, req.UserID)
	if err != nil {
		logger.Error().Err(err).Msg("failed to cancel deletion")
		return c.NoContent(http.StatusInternalServerError)
	}
	return c.NoContent(http.StatusOK)
}
