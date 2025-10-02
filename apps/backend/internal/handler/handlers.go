package handler

import (
	"github.com/petonlabs/go-boilerplate/internal/server"
	"github.com/petonlabs/go-boilerplate/internal/service"
)

type Handlers struct {
	Health  *HealthHandler
	OpenAPI *OpenAPIHandler
	Dspy    *DspyHandler
}

func NewHandlers(s *server.Server, services *service.Services) *Handlers {
	return &Handlers{
		Health:  NewHealthHandler(s),
		OpenAPI: NewOpenAPIHandler(s),
		Dspy:    NewDspyHandler(s),
	}
}
