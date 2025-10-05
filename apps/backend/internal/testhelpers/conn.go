package testhelpers

import (
	"context"
	"fmt"
	"time"

	"github.com/petonlabs/go-boilerplate/internal/config"
	"github.com/petonlabs/go-boilerplate/internal/database"
	"github.com/rs/zerolog"
)

// newDB is a package-level variable so tests can replace it. It defaults to
// calling database.New with the provided args.
var newDB = database.New

// pingDB is a package-level hook used to ping the database. By default it
// calls db.Pool.Ping(ctx) but tests can override it to avoid real network I/O.
var pingDB = func(db *database.Database, ctx context.Context) error {
	if db == nil || db.Pool == nil {
		return fmt.Errorf("no database pool available")
	}
	return db.Pool.Ping(ctx)
}

// connectWithRetry attempts to create a database connection and ping it up to
// retries times. It logs warnings on connection or ping failures and sleeps
// 2 seconds between attempts (but not after the last attempt). Returns the
// connected *database.Database or the last error encountered.
func connectWithRetry(cfg *config.Config, logger *zerolog.Logger, retries int) (*database.Database, error) {
	var db *database.Database
	var lastErr error
	if retries <= 0 {
		retries = 1
	}
	for i := 0; i < retries; i++ {
		db, lastErr = newDB(cfg, logger, nil)
		if lastErr == nil {
			// Try a ping to verify the connection using a background context
			ctx := context.Background()
			if pingErr := pingDB(db, ctx); pingErr == nil {
				return db, nil
			} else {
				lastErr = pingErr
				logger.Warn().Err(pingErr).Msgf("Failed to ping database (attempt %d/%d)", i+1, retries)
				if db != nil && db.Pool != nil {
					db.Pool.Close()
				}
			}
		} else {
			logger.Warn().Err(lastErr).Msgf("Failed to connect to database (attempt %d/%d)", i+1, retries)
		}

		if i < retries-1 {
			time.Sleep(2 * time.Second)
		}
	}
	return nil, lastErr
}
