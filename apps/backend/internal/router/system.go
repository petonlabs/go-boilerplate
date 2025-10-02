package router

import (
	"github.com/sriniously/go-boilerplate/internal/handler"

	"github.com/labstack/echo/v4"
)

func registerSystemRoutes(r *echo.Echo, h *handler.Handlers) {
	r.GET("/status", h.Health.CheckHealth)
	r.GET("/health", h.Health.CheckHealth)
	r.GET("/dspy/health", h.Dspy.CheckHealth)

	r.Static("/static", "static")

	r.GET("/docs", h.OpenAPI.ServeOpenAPIUI)
}
