package job

import (
	"encoding/json"
	"time"

	"github.com/hibiken/asynq"
)

const (
	TaskUserDelete = "user:delete"
)

type UserDeletePayload struct {
	UserID string `json:"user_id"`
}

func NewUserDeleteTask(userID string) (*asynq.Task, error) {
	payload, err := json.Marshal(UserDeletePayload{UserID: userID})
	if err != nil {
		return nil, err
	}

	return asynq.NewTask(TaskUserDelete, payload,
		asynq.MaxRetry(5),
		asynq.Queue("critical"),
		asynq.Timeout(60*time.Second)), nil
}
