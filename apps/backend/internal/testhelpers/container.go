package testhelpers

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
	// If the DSN looks like a URL, prefer parsing query params from it.
	if strings.Contains(dsn, "://") {
		if u, err := url.Parse(dsn); err == nil {
			if v := u.Query().Get(key); v != "" {
				return v
			}
		}
		// fallthrough to libpq-style parsing if URL parse didn't yield the param
	}
	// Token-scan the libpq-style DSN to ensure the key match is a true parameter
	// boundary. Tokens are space-separated; values may be quoted with single or
	// double quotes. This parser will iterate tokens and compare the key part
	// case-insensitively.
	i := 0
	n := len(dsn)
	wanted := strings.ToLower(key)

	for i < n {
		for i < n && (dsn[i] == ' ' || dsn[i] == '\t' || dsn[i] == '\n' || dsn[i] == '\r') {
			i++
		}
		if i >= n {
			break
		}

		// Read key until '=' or whitespace
		keyStart := i
		for i < n && dsn[i] != '=' && dsn[i] != ' ' && dsn[i] != '\t' && dsn[i] != '\n' && dsn[i] != '\r' {
			i++
		}
		if i >= n || dsn[i] != '=' {
			// No '=', skip to next whitespace
			for i < n && dsn[i] != ' ' && dsn[i] != '\t' && dsn[i] != '\n' && dsn[i] != '\r' {
				i++
			}
			continue
		}
		keyStr := dsn[keyStart:i]
		i++ // skip '='

		if i >= n {
			// key= at end with empty value
			if strings.EqualFold(keyStr, wanted) {
				return ""
			}
			break
		}
		var val string
		if dsn[i] == '\'' || dsn[i] == '"' {
			q := dsn[i]
			i++
			// Build unescaped value handling backslash-escapes
			var sb strings.Builder
			for i < n {
				if dsn[i] == '\\' {
					// Escape: consume backslash and append next char literally if present
					i++
					if i < n {
						sb.WriteByte(dsn[i])
						i++
						continue
					}
					// trailing backslash, append as-is
					sb.WriteByte('\\')
					break
				}
				if dsn[i] == q {
					i++
					break
				}
				sb.WriteByte(dsn[i])
				i++
			}
			val = sb.String()
		} else {
			// unquoted value: read until next whitespace
			valStart := i
			for i < n && dsn[i] != ' ' && dsn[i] != '\t' && dsn[i] != '\n' && dsn[i] != '\r' {
				i++
			}
			val = dsn[valStart:i]
		}

		if strings.EqualFold(strings.TrimSpace(keyStr), wanted) {
			return strings.TrimSpace(val)
		}
		// else continue scanning
	}
	return ""
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
	var db *database.Database
	var lastErr error
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

		db, lastErr = connectWithRetry(cfg, &logger, 5, nil)
		require.NoError(t, lastErr, "failed to connect to database via TEST_DATABASE_DSN after multiple attempts")

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

	db, lastErr = connectWithRetry(cfg, &logger, 5, nil)
	require.NoError(t, lastErr, "failed to connect to database after multiple attempts")

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
