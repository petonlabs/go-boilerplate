package testhelpers

import (
	"context"
	"fmt"
	"time"

	"github.com/petonlabs/go-boilerplate/internal/config"
	"github.com/petonlabs/go-boilerplate/internal/database"
	loggerPkg "github.com/petonlabs/go-boilerplate/internal/logger"
	"github.com/rs/zerolog"
)

// connectWithRetry previously relied on package-level hooks (newDB/pingDB)
// which made tests mutate shared globals. We now use ConnectHooks for
// dependency injection so tests can pass replacements safely.

// connectWithRetry attempts to create a database connection and ping it up to
// retries times. It logs warnings on connection or ping failures and sleeps
// 2 seconds between attempts (but not after the last attempt). Returns the
// connected *database.Database or the last error encountered.
// ConnectHooks contains injectable functions used by connectWithRetry. Tests
// can provide their own implementations to avoid network I/O.
type ConnectHooks struct {
	NewDB  func(cfg *config.Config, logger *zerolog.Logger, loggerService *loggerPkg.LoggerService) (*database.Database, error)
	PingDB func(db *database.Database, ctx context.Context) error
}

// connectWithRetry attempts to create a database connection and ping it up to
// retries times. Hooks may be provided for NewDB and PingDB; when nil, defaults
// are used which call into the real database package and pgx pool.
func connectWithRetry(cfg *config.Config, logger *zerolog.Logger, retries int, hooks *ConnectHooks) (*database.Database, error) {
	var db *database.Database
	var lastErr error
	if retries <= 0 {
		retries = 1
	}

	// Determine which implementations to use
	newDBImpl := database.New
	pingDBImpl := func(db *database.Database, ctx context.Context) error {
		if db == nil || db.Pool == nil {
			return fmt.Errorf("no database pool available")
		}
		return db.Pool.Ping(ctx)
	}
	if hooks != nil {
		if hooks.NewDB != nil {
			newDBImpl = func(cfg *config.Config, logger *zerolog.Logger, loggerService *loggerPkg.LoggerService) (*database.Database, error) {
				return hooks.NewDB(cfg, logger, loggerService)
			}
		}
		if hooks.PingDB != nil {
			pingDBImpl = hooks.PingDB
		}
	}

	for i := 0; i < retries; i++ {
		db, lastErr = newDBImpl(cfg, logger, nil)
		if lastErr == nil {
			// Try a ping to verify the connection using a background context
			ctx := context.Background()
			if pingErr := pingDBImpl(db, ctx); pingErr == nil {
				return db, nil
			} else {
				lastErr = pingErr
				// Log at Debug level for initial connection attempts to reduce noise
				// Only the final failure will be logged at Warn level by the caller
				logger.Debug().Err(pingErr).Msgf("Failed to ping database (attempt %d/%d)", i+1, retries)
				if db != nil && db.Pool != nil {
					db.Pool.Close()
				}
			}
		} else {
			// Log at Debug level for initial connection attempts to reduce noise
			logger.Debug().Err(lastErr).Msgf("Failed to connect to database (attempt %d/%d)", i+1, retries)
		}

		if i < retries-1 {
			time.Sleep(2 * time.Second)
		}
	}
	return nil, lastErr
}
