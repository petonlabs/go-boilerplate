package testing

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/petonlabs/go-boilerplate/internal/config"
	"github.com/petonlabs/go-boilerplate/internal/database"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// extractLibpqParam scans a libpq-style DSN (space-separated key=value pairs)
// and returns the value for the given key if present. It handles quoted
// values with single or double quotes. Returns empty string if not found.
func extractLibpqParam(dsn, key string) string {
	// Simple scan: look for key= and read until whitespace.
	lower := strings.ToLower(dsn)
	search := strings.ToLower(key) + "="
	idx := strings.Index(lower, search)
	if idx == -1 {
		return ""
	}
	// start of value in original dsn (preserve original case/quotes)
	start := idx + len(search)
	if start >= len(dsn) {
		return ""
	}
	rest := dsn[start:]
	// If value is quoted, extract until matching quote
	if rest[0] == '\'' || rest[0] == '"' {
		q := rest[0]
		// find next unescaped quote
		for i := 1; i < len(rest); i++ {
			if rest[i] == q {
				return strings.TrimSpace(rest[1:i])
			}
		}
		// no closing quote, return trimmed remainder
		return strings.TrimSpace(rest[1:])
	}
	// unquoted: read until whitespace
	end := len(rest)
	for i := 0; i < len(rest); i++ {
		if rest[i] == ' ' || rest[i] == '\t' || rest[i] == '\n' || rest[i] == '\r' {
			end = i
			break
		}
	}
	return strings.TrimSpace(rest[:end])
}

type TestDB struct {
	Pool      *pgxpool.Pool
	Container testcontainers.Container
	Config    *config.Config
}

// SetupTestDB creates a Postgres container and applies migrations
func SetupTestDB(t *testing.T) (*TestDB, func()) {
	t.Helper()

	ctx := context.Background()
	// Allow overriding container startup with an external DSN for local testing
	if dsn := os.Getenv("TEST_DATABASE_DSN"); dsn != "" {
		logger := zerolog.New(zerolog.NewConsoleWriter()).With().Timestamp().Logger()

		// Parse DSN to create a minimal config
		pgCfg, err := pgxpool.ParseConfig(dsn)
		require.NoError(t, err, "failed to parse TEST_DATABASE_DSN")

		// Extract host and port (pgx provides ConnConfig.Config().Host but keep it simple)
		// Determine sslmode: prefer DSN query param, then TEST_DATABASE_SSL_MODE env, else default to disable
		parsedURL, perr := url.Parse(dsn)
		var sslMode string
		if perr == nil {
			sslMode = parsedURL.Query().Get("sslmode")
		}
		// If no sslmode from URL query, try libpq-style key parsing from raw DSN
		if strings.TrimSpace(sslMode) == "" {
			sslMode = extractLibpqParam(dsn, "sslmode")
		}
		if strings.TrimSpace(sslMode) == "" {
			sslMode = os.Getenv("TEST_DATABASE_SSL_MODE")
			if strings.TrimSpace(sslMode) == "" {
				sslMode = "disable"
			}
		}

		cfg := &config.Config{
			Database: config.DatabaseConfig{
				Host:            pgCfg.ConnConfig.Host,
				Port:            int(pgCfg.ConnConfig.Port),
				User:            pgCfg.ConnConfig.User,
				Password:        pgCfg.ConnConfig.Password,
				Name:            pgCfg.ConnConfig.Database,
				SSLMode:         sslMode,
				MaxOpenConns:    25,
				MaxIdleConns:    25,
				ConnMaxLifetime: 300,
				ConnMaxIdleTime: 300,
			},
			Primary:     config.Primary{Env: "test"},
			Server:      config.ServerConfig{Port: "8080", ReadTimeout: 30, WriteTimeout: 30, IdleTimeout: 30, CORSAllowedOrigins: []string{"*"}},
			Integration: config.IntegrationConfig{ResendAPIKey: "test-key"},
			Redis:       config.RedisConfig{Address: "localhost:6379"},
			Auth:        config.AuthConfig{SecretKey: "test-secret"},
		}

		var db *database.Database
		var lastErr error
		for i := 0; i < 5; i++ {
			db, lastErr = database.New(cfg, &logger, nil)
			if lastErr == nil {
				if err := db.Pool.Ping(ctx); err == nil {
					break
				}
				lastErr = err
				db.Pool.Close()
			}
			time.Sleep(2 * time.Second)
		}
		require.NoError(t, lastErr, "failed to connect to database via TEST_DATABASE_DSN")

		// Apply migrations on the external DSN so schema is prepared for tests.
		if err := database.Migrate(ctx, &logger, cfg); err != nil {
			if db != nil && db.Pool != nil {
				db.Pool.Close()
			}
			require.NoError(t, err, "failed to apply database migrations via TEST_DATABASE_DSN")
		}

		testDB := &TestDB{Pool: db.Pool, Container: nil, Config: cfg}
		cleanup := func() {
			if db.Pool != nil {
				db.Pool.Close()
			}
		}
		return testDB, cleanup
	}
	dbName := fmt.Sprintf("test_db_%s", uuid.New().String()[:8])
	dbUser := "testuser"
	dbPassword := "testpassword"

	req := testcontainers.ContainerRequest{
		Image:        "postgres:15-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_DB":       dbName,
			"POSTGRES_USER":     dbUser,
			"POSTGRES_PASSWORD": dbPassword,
		},
		WaitingFor: wait.ForLog("database system is ready to accept connections").WithStartupTimeout(30 * time.Second),
	}

	// Call GenericContainer inside a recover wrapper because testcontainers may panic
	// when Docker isn't available (MustExtractDockerHost). Convert panics to errors so
	// we can skip tests gracefully.
	var pgContainer testcontainers.Container
	var err error
	func() {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("panic starting container: %v", r)
			}
		}()
		pgContainer, err = testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
			ContainerRequest: req,
			Started:          true,
		})
	}()
	if err != nil {
		// If Docker is not available in the environment, skip these tests rather than fail.
		// Match only specific known error messages to avoid masking unrelated errors.
		es := strings.ToLower(err.Error())
		if strings.Contains(es, "rootless docker not found") || strings.Contains(es, "cannot connect to the docker daemon") || strings.Contains(es, "dial unix /var/run/docker.sock") {
			t.Skipf("skipping container-based tests: %v", err)
		}
		require.NoError(t, err, "failed to start postgres container")
	}

	host, err := pgContainer.Host(ctx)
	require.NoError(t, err, "failed to get container host")

	mappedPort, err := pgContainer.MappedPort(ctx, "5432")
	require.NoError(t, err, "failed to get mapped port")
	port := mappedPort.Int()

	// Make sure the test cleans up the container
	t.Cleanup(func() {
		if err := pgContainer.Terminate(ctx); err != nil {
			t.Logf("failed to terminate container: %v", err)
		}
	})

	// Create configuration
	// Determine sslmode for container-created DB: prefer TEST_DATABASE_SSL_MODE env or default to disable
	sslMode := os.Getenv("TEST_DATABASE_SSL_MODE")
	if sslMode == "" {
		sslMode = "disable"
	}

	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Host:            host,
			Port:            port,
			User:            dbUser,
			Password:        dbPassword,
			Name:            dbName,
			SSLMode:         sslMode,
			MaxOpenConns:    25,
			MaxIdleConns:    25,
			ConnMaxLifetime: 300,
			ConnMaxIdleTime: 300,
		},
		Primary: config.Primary{
			Env: "test",
		},
		Server: config.ServerConfig{
			Port:               "8080",
			ReadTimeout:        30,
			WriteTimeout:       30,
			IdleTimeout:        30,
			CORSAllowedOrigins: []string{"*"},
		},
		Integration: config.IntegrationConfig{
			ResendAPIKey: "test-key",
		},
		Redis: config.RedisConfig{
			Address: "localhost:6379",
		},
		Auth: config.AuthConfig{
			SecretKey: "test-secret",
		},
	}

	logger := zerolog.New(zerolog.NewConsoleWriter()).With().Timestamp().Logger()

	var db *database.Database
	var lastErr error
	for i := 0; i < 5; i++ {
		// Sleep before first attempt too to give PostgreSQL time to initialize
		time.Sleep(2 * time.Second)

		db, lastErr = database.New(cfg, &logger, nil)
		if lastErr == nil {
			// Try a ping to verify the connection
			if err := db.Pool.Ping(ctx); err == nil {
				break
			} else {
				lastErr = err
				logger.Warn().Err(err).Msg("Failed to ping database, will retry")
				db.Pool.Close() // Close the failed connection
			}
		} else {
			logger.Warn().Err(lastErr).Msgf("Failed to connect to database (attempt %d/5)", i+1)
		}
	}
	require.NoError(t, lastErr, "failed to connect to database after multiple attempts")

	// Apply migrations
	err = database.Migrate(ctx, &logger, cfg)
	require.NoError(t, err, "failed to apply database migrations")

	testDB := &TestDB{
		Pool:      db.Pool,
		Container: pgContainer,
		Config:    cfg,
	}

	// Return cleanup function that just closes the pool (container is managed by t.Cleanup)
	cleanup := func() {
		if db.Pool != nil {
			db.Pool.Close()
		}
	}

	return testDB, cleanup
}

// CleanupTestDB closes the database connection and terminates the container
func (db *TestDB) CleanupTestDB(ctx context.Context, logger *zerolog.Logger) error {
	logger.Info().Msg("cleaning up test database")

	if db.Pool != nil {
		db.Pool.Close()
	}

	if db.Container != nil {
		if err := db.Container.Terminate(ctx); err != nil {
			return fmt.Errorf("failed to terminate container: %w", err)
		}
	}

	return nil
}
