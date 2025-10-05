package testhelpers

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/petonlabs/go-boilerplate/internal/config"
	"github.com/petonlabs/go-boilerplate/internal/database"
	loggerConfig "github.com/petonlabs/go-boilerplate/internal/logger"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

func TestConnectWithRetry_SuccessFirstAttempt(t *testing.T) {
	t.Parallel()
	// Use dependency injection via ConnectHooks to avoid mutating package globals
	hooks := &ConnectHooks{
		NewDB: func(cfg *config.Config, logger *zerolog.Logger, loggerService *loggerConfig.LoggerService) (*database.Database, error) {
			return &database.Database{Pool: &pgxpool.Pool{}}, nil
		},
		PingDB: func(db *database.Database, ctx context.Context) error {
			require.NotNil(t, db)
			return nil
		},
	}

	cfg := &config.Config{}
	logger := zerolog.New(nil)
	db, err := connectWithRetry(cfg, &logger, 3, hooks)
	require.NoError(t, err)
	require.NotNil(t, db)
}

func TestConnectWithRetry_RetryThenSuccess(t *testing.T) {
	t.Parallel()

	calls := 0
	hooks := &ConnectHooks{
		NewDB: func(cfg *config.Config, logger *zerolog.Logger, loggerService *loggerConfig.LoggerService) (*database.Database, error) {
			calls++
			if calls == 1 {
				return nil, errors.New("temporary failure")
			}
			return &database.Database{Pool: &pgxpool.Pool{}}, nil
		},
		PingDB: func(db *database.Database, ctx context.Context) error {
			require.NotNil(t, db)
			return nil
		},
	}

	cfg := &config.Config{}
	logger := zerolog.New(nil)
	db, err := connectWithRetry(cfg, &logger, 2, hooks)
	require.NoError(t, err)
	require.NotNil(t, db)
	require.Equal(t, 2, calls, "expected two attempts")
}

func TestConnectWithRetry_AllAttemptsFail(t *testing.T) {
	t.Parallel()
	hooks := &ConnectHooks{
		NewDB: func(cfg *config.Config, logger *zerolog.Logger, loggerService *loggerConfig.LoggerService) (*database.Database, error) {
			return nil, errors.New("cannot connect")
		},
		PingDB: func(db *database.Database, ctx context.Context) error {
			return errors.New("should not be called")
		},
	}

	cfg := &config.Config{}
	logger := zerolog.Nop()
	db, err := connectWithRetry(cfg, &logger, 3, hooks)
	require.Error(t, err)
	require.Nil(t, db)
}
