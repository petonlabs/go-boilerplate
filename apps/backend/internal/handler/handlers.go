package handler

import (
	"github.com/petonlabs/go-boilerplate/internal/server"
	"github.com/petonlabs/go-boilerplate/internal/service"
)

type Handlers struct {
	Health  *HealthHandler
	OpenAPI *OpenAPIHandler
	Dspy    *DspyHandler
	Webhook *WebhookHandler
	Auth    *AuthHandler
	Admin   *AdminHandler
}

func NewHandlers(s *server.Server, services *service.Services) *Handlers {
	return &Handlers{
		Health:  NewHealthHandler(s, services),
		OpenAPI: NewOpenAPIHandler(s, services),
		Dspy:    NewDspyHandler(s, services),
		Webhook: NewWebhookHandler(s, services),
		Auth:    NewAuthHandler(s, services),
		Admin:   NewAdminHandler(s, services),
	}
}
