package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/petonlabs/go-boilerplate/internal/dspy"
	"github.com/petonlabs/go-boilerplate/internal/server"
	"github.com/petonlabs/go-boilerplate/internal/service"
)

type DspyHandler struct {
	Handler
}

func NewDspyHandler(s *server.Server, services *service.Services) *DspyHandler {
	return &DspyHandler{
		Handler: NewHandler(s, services),
	}
}

func (h *DspyHandler) CheckHealth(c echo.Context) error {
	client, err := dspy.New()
	if err != nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]interface{}{
			"status": "disabled",
			"error":  err.Error(),
		})
	}

	ctx, cancel := context.WithTimeout(c.Request().Context(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx); err != nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]interface{}{
			"status": "unhealthy",
			"error":  err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"status": "ok",
	})
}
