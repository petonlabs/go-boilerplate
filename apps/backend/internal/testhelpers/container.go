package testhelpers

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"
	"sync"
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

var (
	// sharedContainer holds a single container instance shared across all tests
	sharedContainer testcontainers.Container
	// sharedConfig holds the config for the shared container
	sharedConfig *config.Config
	// containerMutex protects access to the shared container
	containerMutex sync.Mutex
	// sharedContainerInitialized tracks if the shared container has been set up
	sharedContainerInitialized bool
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

// SetupTestDB creates or reuses a Postgres container and returns a connection to it
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
			cleanupDatabaseTables(t, db.Pool)
			if db.Pool != nil {
				db.Pool.Close()
			}
		}
		return testDB, cleanup
	}

	// Use shared container if available
	containerMutex.Lock()
	if sharedContainerInitialized && sharedContainer != nil && sharedConfig != nil {
		cfg := sharedConfig
		containerMutex.Unlock()

		logger := zerolog.New(zerolog.NewConsoleWriter()).With().Timestamp().Logger()
		db, lastErr = connectWithRetry(cfg, &logger, 5, nil)
		require.NoError(t, lastErr, "failed to connect to shared database after multiple attempts")

		testDB := &TestDB{
			Pool:      db.Pool,
			Container: sharedContainer,
			Config:    cfg,
		}

		cleanup := func() {
			cleanupDatabaseTables(t, db.Pool)
			if db.Pool != nil {
				db.Pool.Close()
			}
		}
		return testDB, cleanup
	}
	containerMutex.Unlock()

	// If shared container is not available (e.g., Docker not available), skip test
	if sharedContainerInitialized && sharedContainer == nil {
		t.Skip("skipping container-based tests: Docker not available")
	}

	// Fallback: create a new container for this specific test
	// This shouldn't normally happen with TestMain, but keeps backward compatibility
	dbName := fmt.Sprintf("test_db_%s", uuid.New().String()[:8])
	dbUser := "testuser"
	dbPassword := "testpassword"

	// Disable ryuk container to reduce verbosity and resource usage
	os.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true")

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

	t.Cleanup(func() {
		if err := pgContainer.Terminate(ctx); err != nil {
			t.Logf("failed to terminate container: %v", err)
		}
	})

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

	cleanup := func() {
		if db.Pool != nil {
			db.Pool.Close()
		}
	}

	return testDB, cleanup
}

// cleanupDatabaseTables truncates all tables to ensure test isolation
func cleanupDatabaseTables(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()

	if pool == nil {
		return
	}

	ctx := context.Background()

	// Try to ping first to check if pool is still active
	if err := pool.Ping(ctx); err != nil {
		// Pool is closed or not available, skip cleanup
		return
	}

	// Truncate user tables only (exclude schema_version and other system tables)
	// RESTART IDENTITY resets auto-increment counters
	// CASCADE removes dependent rows in other tables if needed
	_, err := pool.Exec(ctx, `
		DO $$ 
		DECLARE
			r RECORD;
		BEGIN
			FOR r IN (
				SELECT tablename 
				FROM pg_tables 
				WHERE schemaname = 'public' 
				AND tablename NOT IN ('schema_version', 'tern_migrations')
			) LOOP
				EXECUTE 'TRUNCATE TABLE ' || quote_ident(r.tablename) || ' RESTART IDENTITY CASCADE';
			END LOOP;
		END $$;
	`)

	if err != nil {
		t.Logf("warning: failed to truncate tables: %v", err)
	}
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

// SetupSharedContainer creates a single shared Postgres container for all tests
func SetupSharedContainer() error {
	containerMutex.Lock()
	defer containerMutex.Unlock()

	if sharedContainerInitialized {
		return nil
	}

	// Skip shared container setup if external DSN is provided
	if dsn := os.Getenv("TEST_DATABASE_DSN"); dsn != "" {
		sharedContainerInitialized = true
		return nil
	}

	ctx := context.Background()
	dbName := "test_db_shared"
	dbUser := "testuser"
	dbPassword := "testpassword"

	// Disable ryuk container to reduce verbosity and resource usage
	// Ryuk is used for cleanup but we handle cleanup ourselves with t.Cleanup
	os.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true")

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

	var pgContainer testcontainers.Container
	var err error
	func() {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("panic starting shared container: %v", r)
			}
		}()
		pgContainer, err = testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
			ContainerRequest: req,
			Started:          true,
		})
	}()
	if err != nil {
		// If Docker is not available, just mark as initialized and let tests skip
		es := strings.ToLower(err.Error())
		if strings.Contains(es, "rootless docker not found") || strings.Contains(es, "cannot connect to the docker daemon") || strings.Contains(es, "dial unix /var/run/docker.sock") {
			sharedContainerInitialized = true
			return nil
		}
		return fmt.Errorf("failed to start shared postgres container: %w", err)
	}

	host, err := pgContainer.Host(ctx)
	if err != nil {
		return fmt.Errorf("failed to get container host: %w", err)
	}

	mappedPort, err := pgContainer.MappedPort(ctx, "5432")
	if err != nil {
		return fmt.Errorf("failed to get mapped port: %w", err)
	}
	port := mappedPort.Int()

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

	// Connect to the database and apply migrations
	db, err := connectWithRetry(cfg, &logger, 5, nil)
	if err != nil {
		_ = pgContainer.Terminate(ctx)
		return fmt.Errorf("failed to connect to shared database: %w", err)
	}

	if err := database.Migrate(ctx, &logger, cfg); err != nil {
		if db.Pool != nil {
			db.Pool.Close()
		}
		_ = pgContainer.Terminate(ctx)
		return fmt.Errorf("failed to apply migrations to shared database: %w", err)
	}

	// Close the initial connection - each test will create its own
	if db.Pool != nil {
		db.Pool.Close()
	}

	sharedContainer = pgContainer
	sharedConfig = cfg
	sharedContainerInitialized = true

	return nil
}

// CleanupSharedContainer terminates the shared container
func CleanupSharedContainer() {
	containerMutex.Lock()
	defer containerMutex.Unlock()

	if sharedContainer != nil {
		ctx := context.Background()
		_ = sharedContainer.Terminate(ctx)
		sharedContainer = nil
	}
	sharedConfig = nil
	sharedContainerInitialized = false
}

// TestMain sets up a shared container for all tests in the testhelpers package
func TestMain(m *testing.M) {
	// Setup shared container
	if err := SetupSharedContainer(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to setup shared container: %v\n", err)
		os.Exit(1)
	}

	// Run tests
	code := m.Run()

	// Cleanup
	CleanupSharedContainer()

	os.Exit(code)
}
