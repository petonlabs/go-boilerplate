package job

import (
	"errors"

	"github.com/hibiken/asynq"
	"github.com/petonlabs/go-boilerplate/internal/config"
	"github.com/petonlabs/go-boilerplate/internal/database"
	"github.com/petonlabs/go-boilerplate/internal/lib/email"
	"github.com/rs/zerolog"
)

type JobService struct {
	// Client is an abstraction over asynq.Client so tests can inject a mock.
	Client Enqueuer
	server *asynq.Server
	logger *zerolog.Logger
	db     *database.Database
	// email client will be initialized by InitHandlers
	email *email.Client
}

// Enqueuer abstracts the subset of asynq.Client used by our app so tests
// can inject a mock implementation.
type Enqueuer interface {
	Enqueue(*asynq.Task, ...asynq.Option) (*asynq.TaskInfo, error)
	Close() error
}

func NewJobService(logger *zerolog.Logger, cfg *config.Config, db *database.Database) (*JobService, error) {
	if db == nil {
		return nil, errors.New("database is required for JobService")
	}
	if cfg == nil || cfg.Redis.Address == "" {
		return nil, errors.New("redis address required in config for JobService")
	}
	redisAddr := cfg.Redis.Address

	client := asynq.NewClient(asynq.RedisClientOpt{
		Addr: redisAddr,
	})

	server := asynq.NewServer(
		asynq.RedisClientOpt{Addr: redisAddr},
		asynq.Config{
			Concurrency: 10,
			Queues: map[string]int{
				"critical": 6, // Higher priority queue for important emails
				"default":  3, // Default priority for most emails
				"low":      1, // Lower priority for non-urgent emails
			},
		},
	)

	return &JobService{
		Client: client,
		server: server,
		logger: logger,
		db:     db,
	}, nil
}

func (j *JobService) Start() error {
	// Register task handlers
	mux := asynq.NewServeMux()
	mux.HandleFunc(TaskWelcome, j.handleWelcomeEmailTask)
	mux.HandleFunc(TaskUserDelete, j.handleUserDeleteTask)

	j.logger.Info().Msg("Starting background job server")
	if err := j.server.Start(mux); err != nil {
		return err
	}

	return nil
}

func (j *JobService) Stop() {
	j.logger.Info().Msg("Stopping background job server")
	// server may be nil in tests where we only inject a client mock
	if j.server != nil {
		j.server.Shutdown()
	}
	if j.Client != nil {
		if err := j.Client.Close(); err != nil {
			j.logger.Warn().Err(err).Msg("Error closing job client")
		}
	}
}
