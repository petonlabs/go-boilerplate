package handler

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/petonlabs/go-boilerplate/internal/lib/job"
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
	id, err := h.services.Auth.RegisterUser(c.Request().Context(), req.Email, req.Password)
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
	id, err := h.services.Auth.Login(c.Request().Context(), req.Email, req.Password)
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
	token, err := h.services.Auth.RequestPasswordReset(c.Request().Context(), req.Email, time.Duration(h.server.Config.Auth.PasswordResetTTL)*time.Second)
	if err != nil {
		// If the email doesn't exist, treat as success to avoid user enumeration.
		if err == sql.ErrNoRows {
			// Silent success: do not enqueue email and return 204
			return c.NoContent(http.StatusNoContent)
		}
		logger.Error().Err(err).Msg("failed to create password reset token")
		return c.NoContent(http.StatusInternalServerError)
	}
	// Enqueue password reset email job if job client is configured
	if h.server != nil && h.server.Job != nil && h.server.Job.Client != nil {
		expiresAt := time.Now().Add(time.Duration(h.server.Config.Auth.PasswordResetTTL) * time.Second).Unix()
		if task, err := job.NewPasswordResetTask(req.Email, token, expiresAt); err == nil {
			_, _ = h.server.Job.Client.Enqueue(task)
		}
	}
	// In production the token should be delivered only via email.
	// Return the token in the response only for development or test environments
	env := ""
	if h.server != nil && h.server.Config != nil && h.server.Config.Primary.Env != "" {
		env = h.server.Config.Primary.Env
	}
	if env == "development" || env == "test" {
		return c.JSON(http.StatusOK, map[string]string{"token": token})
	}
	// For production, do not return the token in the response. Use 204 No Content.
	return c.NoContent(http.StatusNoContent)
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
	if err := h.services.Auth.ResetPassword(c.Request().Context(), req.Token, req.NewPassword); err != nil {
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
		// Safely convert int64 -> int taking platform bounds into account.
		// On 32-bit platforms int may be 32 bits so an unchecked cast can overflow.
		maxInt := int(^uint(0) >> 1)
		if req.Seconds > int64(maxInt) {
			logger.Warn().Int64("seconds", req.Seconds).Msg("seconds value too large; clamping to max int")
			ttl = maxInt
		} else {
			ttl = int(req.Seconds)
		}
	}
	if err := h.services.Auth.ScheduleDeletion(c.Request().Context(), req.UserID, time.Duration(ttl)*time.Second); err != nil {
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
	if err := h.services.Auth.CancelDeletion(c.Request().Context(), req.UserID); err != nil {
		logger.Error().Err(err).Msg("failed to cancel deletion")
		return c.NoContent(http.StatusInternalServerError)
	}
	return c.NoContent(http.StatusOK)
}
