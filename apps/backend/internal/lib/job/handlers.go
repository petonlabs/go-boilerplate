package job

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hibiken/asynq"
	"github.com/petonlabs/go-boilerplate/internal/config"
	"github.com/petonlabs/go-boilerplate/internal/lib/email"
	"github.com/rs/zerolog"
)

var emailClient *email.Client

func (j *JobService) InitHandlers(config *config.Config, logger *zerolog.Logger) {
	emailClient = email.NewClient(config, logger)
}

func (j *JobService) handleUserDeleteTask(ctx context.Context, t *asynq.Task) error {
	var p UserDeletePayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("failed to unmarshal user delete payload: %w", err)
	}

	j.logger.Info().Str("user_id", p.UserID).Msg("Processing user deletion task")

	if j.db == nil || j.db.Pool == nil {
		j.logger.Error().Msg("database not available to deletion worker")
		return fmt.Errorf("db not available")
	}

	// Ensure the deletion is still scheduled (check deletion_scheduled_at)
	var scheduledAt *time.Time
	err := j.db.Pool.QueryRow(ctx, `SELECT deletion_scheduled_at FROM users WHERE id::text=$1`, p.UserID).Scan(&scheduledAt)
	if err != nil {
		j.logger.Error().Err(err).Str("user_id", p.UserID).Msg("failed to query user for deletion")
		return err
	}
	if scheduledAt == nil {
		j.logger.Info().Str("user_id", p.UserID).Msg("deletion no longer scheduled, skipping")
		return nil
	}
	if time.Now().Before(*scheduledAt) {
		j.logger.Info().Str("user_id", p.UserID).Msg("deletion scheduled in the future, skipping")
		return nil
	}

	// Perform deletion: here we soft-delete by setting deleted_at to now and clearing sensitive fields
	_, err = j.db.Pool.Exec(ctx, `UPDATE users SET deleted_at = now(), email = NULL, password_hash = NULL WHERE id::text = $1`, p.UserID)
	if err != nil {
		j.logger.Error().Err(err).Str("user_id", p.UserID).Msg("failed to delete user")
		return err
	}

	j.logger.Info().Str("user_id", p.UserID).Msg("User deletion completed")
	return nil
}

func (j *JobService) handleWelcomeEmailTask(ctx context.Context, t *asynq.Task) error {
	var p WelcomeEmailPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("failed to unmarshal welcome email payload: %w", err)
	}

	j.logger.Info().
		Str("type", "welcome").
		Str("to", p.To).
		Msg("Processing welcome email task")

	err := emailClient.SendWelcomeEmail(
		p.To,
		p.FirstName,
	)
	if err != nil {
		j.logger.Error().
			Str("type", "welcome").
			Str("to", p.To).
			Err(err).
			Msg("Failed to send welcome email")
		return err
	}

	j.logger.Info().
		Str("type", "welcome").
		Str("to", p.To).
		Msg("Successfully sent welcome email")
	return nil
}
