package router

import (
	"github.com/petonlabs/go-boilerplate/internal/handler"

	"github.com/labstack/echo/v4"
)

func registerSystemRoutes(r *echo.Echo, h *handler.Handlers) {
	r.GET("/status", h.Health.CheckHealth)
	r.GET("/health", h.Health.CheckHealth)
	r.GET("/dspy/health", h.Dspy.CheckHealth)

	r.Static("/static", "static")

	r.GET("/docs", h.OpenAPI.ServeOpenAPIUI)
	// Clerk webhook endpoint
	r.POST("/webhooks/clerk", h.Webhook.HandleClerkWebhook)
	// Auth endpoints
	r.POST("/auth/register", h.Auth.Register)
	r.POST("/auth/login", h.Auth.Login)
	r.POST("/auth/password/request", h.Auth.RequestPasswordReset)
	r.POST("/auth/password/reset", h.Auth.ResetPassword)
	r.POST("/auth/schedule_deletion", h.Auth.ScheduleDeletion)
	r.POST("/auth/cancel_deletion", h.Auth.CancelDeletion)
}
