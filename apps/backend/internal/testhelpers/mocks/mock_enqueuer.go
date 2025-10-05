package mocks

import (
	"sync"

	"github.com/hibiken/asynq"
)

// MockEnqueuer records enqueued tasks for assertions in tests.
type MockEnqueuer struct {
	mu    sync.Mutex
	tasks []*asynq.Task
}

func NewMockEnqueuer() *MockEnqueuer { return &MockEnqueuer{} }

func (m *MockEnqueuer) Enqueue(t *asynq.Task, _ ...asynq.Option) (*asynq.TaskInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.tasks = append(m.tasks, t)
	return &asynq.TaskInfo{Type: t.Type()}, nil
}

// GetTasks returns a copy of the enqueued tasks. The returned slice is a shallow
// copy of the []*asynq.Task slice to avoid exposing internal state for mutation.
func (m *MockEnqueuer) GetTasks() []*asynq.Task {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]*asynq.Task, len(m.tasks))
	copy(out, m.tasks)
	return out
}

func (m *MockEnqueuer) Close() error { return nil }
