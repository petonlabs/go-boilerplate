package router

import (
	"github.com/labstack/echo/v4"
	"github.com/petonlabs/go-boilerplate/internal/handler"
	"github.com/petonlabs/go-boilerplate/internal/middleware"
	"net/http"
)

func registerAdminRoutes(g *echo.Group, h *handler.Handlers, m *middleware.Middlewares) {
	adminGroup := g.Group("/admin")
	adminGroup.Use(m.Auth.RequireAuth, m.Auth.RequireRole("admin"))

	adminGroup.GET("/health", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})
}
