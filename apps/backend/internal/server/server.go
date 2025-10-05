package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/newrelic/go-agent/v3/integrations/nrredis-v9"
	"github.com/petonlabs/go-boilerplate/internal/config"
	"github.com/petonlabs/go-boilerplate/internal/database"
	"github.com/petonlabs/go-boilerplate/internal/lib/job"
	loggerPkg "github.com/petonlabs/go-boilerplate/internal/logger"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

type Server struct {
	// configPtr holds an immutable pointer to the active config. Use
	// GetConfig/SetConfig to access or replace it atomically.
	configPtr     atomic.Pointer[config.Config]
	Logger        *zerolog.Logger
	LoggerService *loggerPkg.LoggerService
	DB            *database.Database
	Redis         *redis.Client
	httpServer    *http.Server
	Job           *job.JobService
}

func New(cfg *config.Config, logger *zerolog.Logger, loggerService *loggerPkg.LoggerService) (*Server, error) {
	db, err := database.New(cfg, logger, loggerService)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	// Redis client with New Relic integration
	redisClient := redis.NewClient(&redis.Options{
		Addr: cfg.Redis.Address,
	})

	// Add New Relic Redis hooks if available
	if loggerService != nil && loggerService.GetApplication() != nil {
		redisClient.AddHook(nrredis.NewHook(redisClient.Options()))
	}

	// Test Redis connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := redisClient.Ping(ctx).Err(); err != nil {
		logger.Error().Err(err).Msg("Failed to connect to Redis, continuing without Redis")
		// Don't fail startup if Redis is unavailable
	}

	// job service (inject DB at construction so handlers have access)
	jobService, err := job.NewJobService(logger, cfg, db)
	if err != nil {
		return nil, err
	}
	jobService.InitHandlers(cfg, logger)

	// Start job server
	if err := jobService.Start(); err != nil {
		return nil, err
	}

	server := &Server{
		Logger:        logger,
		LoggerService: loggerService,
		DB:            db,
		Redis:         redisClient,
		Job:           jobService,
	}
	// Store initial config atomically
	server.SetConfig(cfg)

	// Start metrics collection
	// Runtime metrics are automatically collected by New Relic Go agent

	return server, nil
}

func (s *Server) SetupHTTPServer(handler http.Handler) {
	cfg := s.getConfig()
	if cfg == nil {
		// Fallback: if no config is available, initialize with conservative timeouts
		// to avoid Slowloris attack vectors (gosec G112).
		s.httpServer = &http.Server{
			Handler:           handler,
			ReadHeaderTimeout: 5 * time.Second,
			ReadTimeout:       5 * time.Second,
			WriteTimeout:      10 * time.Second,
			IdleTimeout:       60 * time.Second,
		}
		return
	}

	s.httpServer = &http.Server{
		Addr:              ":" + cfg.Server.Port,
		Handler:           handler,
		ReadHeaderTimeout: time.Duration(cfg.Server.ReadTimeout) * time.Second,
		ReadTimeout:       time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout:      time.Duration(cfg.Server.WriteTimeout) * time.Second,
		IdleTimeout:       time.Duration(cfg.Server.IdleTimeout) * time.Second,
	}
}

func (s *Server) Start() error {
	if s.httpServer == nil {
		return errors.New("HTTP server not initialized")
	}

	cfg := s.getConfig()
	// Use empty strings if cfg is nil
	port := ""
	env := ""
	if cfg != nil {
		port = cfg.Server.Port
		env = cfg.Primary.Env
	}

	s.Logger.Info().
		Str("port", port).
		Str("env", env).
		Msg("starting server")

	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown HTTP server: %w", err)
	}

	if err := s.DB.Close(); err != nil {
		return fmt.Errorf("failed to close database connection: %w", err)
	}

	if s.Job != nil {
		s.Job.Stop()
	}

	return nil
}

// GetTokenHMACSecret returns the current TokenHMACSecret from the server config
// under a read lock. It returns an empty string if Config or Auth is nil.
func (s *Server) GetTokenHMACSecret() string {
	cfg := s.GetConfig()
	if cfg == nil {
		return ""
	}
	return cfg.Auth.TokenHMACSecret
}

// SetTokenHMACSecret sets the TokenHMACSecret in the server config under a write lock.
// This is intended as a deliberate, synchronized persistence path for runtime secret
// rotation. Callers should ensure secrets are distributed securely in production.
func (s *Server) SetTokenHMACSecret(newSecret string) {
	// Use an atomic compare-and-swap loop to avoid races between Load and Store.
	// We make a deep-ish copy of mutable fields (slices and pointer sub-structs)
	// so the new config snapshot does not share backing memory with the old one.
	for {
		oldPtr := s.configPtr.Load()
		if oldPtr == nil {
			// Nothing to update
			return
		}

		// shallow copy of the struct value pointed to by oldPtr
		copyCfg := *oldPtr

		// copy mutable slice fields to avoid sharing backing arrays
		if oldPtr.Server.CORSAllowedOrigins != nil {
			copied := make([]string, len(oldPtr.Server.CORSAllowedOrigins))
			copy(copied, oldPtr.Server.CORSAllowedOrigins)
			copyCfg.Server.CORSAllowedOrigins = copied
		}

		// copy Observability pointer if present
		if oldPtr.Observability != nil {
			obs := *oldPtr.Observability
			copyCfg.Observability = &obs
		}

		// mutate the copy
		copyCfg.Auth.TokenHMACSecret = newSecret

		// attempt to swap; if success we're done, otherwise retry
		if s.configPtr.CompareAndSwap(oldPtr, &copyCfg) {
			return
		}

		// CAS failed; loop and retry with latest value
	}
}

// getConfig returns a snapshot copy of the current server config under a
// read lock. The returned value is safe for callers to read without holding
// the server's internal lock. Note this performs a shallow copy of the
// top-level config and nested structs; slices and pointer fields are copied
// where necessary to avoid shared mutable references for common cases.
func (s *Server) getConfig() *config.Config {
	// Atomic load of pointer
	p := s.configPtr.Load()
	if p == nil {
		return nil
	}
	// Shallow copy top-level struct
	cfg := *p

	// Copy slice fields to avoid shared backing arrays
	if cfg.Server.CORSAllowedOrigins != nil {
		copied := make([]string, len(cfg.Server.CORSAllowedOrigins))
		copy(copied, cfg.Server.CORSAllowedOrigins)
		cfg.Server.CORSAllowedOrigins = copied
	}

	// Deep copy Observability pointer if present
	if p.Observability != nil {
		obs := *p.Observability
		// If Observability contains slices/pointers, copy them here as needed
		cfg.Observability = &obs
	}

	return &cfg
}

// GetConfig returns a snapshot of the current server config. It is safe for
// callers to read the returned value without holding any locks. To update the
// server's config at runtime, use SetConfig which will replace the stored
// config under a write lock.
func (s *Server) GetConfig() *config.Config {
	return s.getConfig()
}

// SetConfig atomically replaces the server's config with the provided
// snapshot. The input is copied to avoid sharing mutable backing data.
func (s *Server) SetConfig(cfg *config.Config) {
	if cfg == nil {
		s.configPtr.Store(nil)
		return
	}
	copyCfg := *cfg
	if cfg.Server.CORSAllowedOrigins != nil {
		copied := make([]string, len(cfg.Server.CORSAllowedOrigins))
		copy(copied, cfg.Server.CORSAllowedOrigins)
		copyCfg.Server.CORSAllowedOrigins = copied
	}
	if cfg.Observability != nil {
		obs := *cfg.Observability
		copyCfg.Observability = &obs
	}
	s.configPtr.Store(&copyCfg)
}
