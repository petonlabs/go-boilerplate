package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/petonlabs/go-boilerplate/internal/config"
	"github.com/petonlabs/go-boilerplate/internal/database"
	"github.com/petonlabs/go-boilerplate/internal/handler"
	"github.com/petonlabs/go-boilerplate/internal/logger"
	"github.com/petonlabs/go-boilerplate/internal/repository"
	"github.com/petonlabs/go-boilerplate/internal/router"
	"github.com/petonlabs/go-boilerplate/internal/server"
	"github.com/petonlabs/go-boilerplate/internal/service"
)

const DefaultContextTimeout = 30

func main() {
	// Parse flags for healthcheck
	healthcheck := flag.Bool("healthcheck", false, "Run healthcheck and exit")
	flag.Parse()

	// If healthcheck flag is set, perform healthcheck and exit
	if *healthcheck {
		port := os.Getenv("SERVER_PORT")
		if port == "" {
			port = "8080"
		}
		resp, err := http.Get(fmt.Sprintf("http://localhost:%s/health", port))
		if err != nil {
			fmt.Fprintf(os.Stderr, "healthcheck failed: %v\n", err)
			os.Exit(1)
		}
		// Close body before os.Exit to avoid exitAfterDefer issue
		if resp.StatusCode != http.StatusOK {
			_ = resp.Body.Close()
			fmt.Fprintf(os.Stderr, "healthcheck failed: status %d\n", resp.StatusCode)
			os.Exit(1)
		}
		_ = resp.Body.Close()
		os.Exit(0)
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		panic("failed to load config: " + err.Error())
	}

	if err := run(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "application error: %v\n", err)
		os.Exit(1)
	}
}

func run(cfg *config.Config) error {
	// Initialize New Relic logger service
	loggerService := logger.NewLoggerService(cfg.Observability)
	defer loggerService.Shutdown()

	log := logger.NewLoggerWithService(cfg.Observability, loggerService)

	if cfg.Primary.Env != "local" {
		if err := database.Migrate(context.Background(), &log, cfg); err != nil {
			return fmt.Errorf("failed to migrate database: %w", err)
		}
	}

	// Initialize server
	srv, err := server.New(cfg, &log, loggerService)
	if err != nil {
		return fmt.Errorf("failed to initialize server: %w", err)
	}

	// Initialize repositories, services, and handlers
	repos := repository.NewRepositories(srv)
	services, serviceErr := service.NewServices(srv, repos)
	if serviceErr != nil {
		return fmt.Errorf("could not create services: %w", serviceErr)
	}
	handlers := handler.NewHandlers(srv, services)

	// Initialize router
	r := router.NewRouter(srv, handlers, services)

	// Setup HTTP server
	srv.SetupHTTPServer(r)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	// Start server
	go func() {
		if err = srv.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error().Err(err).Msg("failed to start server")
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), DefaultContextTimeout*time.Second)
	defer cancel()

	if err = srv.Shutdown(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("server forced to shutdown")
		return fmt.Errorf("server shutdown error: %w", err)
	}

	log.Info().Msg("server exited properly")
	return nil
}
