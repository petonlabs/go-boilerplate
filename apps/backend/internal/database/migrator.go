package database

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"net"
	"net/url"
	"strconv"

	"github.com/petonlabs/go-boilerplate/internal/config"

	"github.com/jackc/pgx/v5"
	tern "github.com/jackc/tern/v2/migrate"
	"github.com/rs/zerolog"
)

//go:embed migrations/*.sql
var migrations embed.FS

func Migrate(ctx context.Context, logger *zerolog.Logger, cfg *config.Config) error {
	hostPort := net.JoinHostPort(cfg.Database.Host, strconv.Itoa(cfg.Database.Port))

	// URL-encode the password
	encodedPassword := url.QueryEscape(cfg.Database.Password)
	dsn := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=%s",
		cfg.Database.User,
		encodedPassword,
		hostPort,
		cfg.Database.Name,
		cfg.Database.SSLMode,
	)

	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		return err
	}
	defer func() {
		_ = conn.Close(ctx)
	}()

	m, err := tern.NewMigrator(ctx, conn, "schema_version")
	if err != nil {
		return fmt.Errorf("constructing database migrator: %w", err)
	}
	subtree, err := fs.Sub(migrations, "migrations")
	if err != nil {
		return fmt.Errorf("retrieving database migrations subtree: %w", err)
	}
	if err := m.LoadMigrations(subtree); err != nil {
		return fmt.Errorf("loading database migrations: %w", err)
	}
	from, err := m.GetCurrentVersion(ctx)
	if err != nil {
		return fmt.Errorf("retreiving current database migration version")
	}
	if err := m.Migrate(ctx); err != nil {
		return err
	}

	// Ensure applied_at is populated for audit purposes for any rows created by the migrator.
	// If this update fails we log a warning but continue the migration process because
	// applied_at is optional and should not block schema changes.
	if _, err := conn.Exec(ctx, `UPDATE schema_version SET applied_at = now() WHERE applied_at IS NULL`); err != nil {
		logger.Warn().Err(err).Msg("failed to populate applied_at on schema_version; continuing")
	}
	// Check for potential overflow before conversion. int(^int32(0)) is -1
	// so the previous check always triggered. Use a proper MaxInt32 value.
	migrationCount := len(m.Migrations)
	const maxInt32 = 1<<31 - 1
	if migrationCount > maxInt32 {
		return fmt.Errorf("migration count exceeds int32 range")
	}
	if from == int32(migrationCount) {
		logger.Info().Msgf("database schema up to date, version %d", migrationCount)
	} else {
		logger.Info().Msgf("migrated database schema, from %d to %d", from, migrationCount)
	}
	return nil
}
