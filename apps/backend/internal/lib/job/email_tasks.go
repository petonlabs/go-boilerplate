package job

import (
	"encoding/json"
	"time"

	"github.com/hibiken/asynq"
)

const (
	TaskWelcome       = "email:welcome"
	TaskPasswordReset = "email:password_reset"
)

type WelcomeEmailPayload struct {
	To        string `json:"to"`
	FirstName string `json:"first_name"`
}

type PasswordResetPayload struct {
	To        string `json:"to"`
	Token     string `json:"token"`
	ExpiresAt int64  `json:"expires_at"`
}

func NewPasswordResetTask(to, token string, expiresAt int64) (*asynq.Task, error) {
	payload, err := json.Marshal(PasswordResetPayload{
		To:        to,
		Token:     token,
		ExpiresAt: expiresAt,
	})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TaskPasswordReset, payload,
		asynq.MaxRetry(3),
		asynq.Queue("default"),
		asynq.Timeout(30*time.Second)), nil
}

func NewWelcomeEmailTask(to, firstName string) (*asynq.Task, error) {
	payload, err := json.Marshal(WelcomeEmailPayload{
		To:        to,
		FirstName: firstName,
	})
	if err != nil {
		return nil, err
	}

	return asynq.NewTask(TaskWelcome, payload,
		asynq.MaxRetry(3),
		asynq.Queue("default"),
		asynq.Timeout(30*time.Second)), nil
}
