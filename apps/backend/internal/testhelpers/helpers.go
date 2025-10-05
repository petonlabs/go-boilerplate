package testhelpers

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/petonlabs/go-boilerplate/internal/lib/job"
	"github.com/petonlabs/go-boilerplate/internal/server"
	"github.com/petonlabs/go-boilerplate/internal/testhelpers/mocks"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

// SetupTest prepares a test environment with a database and server
func SetupTest(t *testing.T) (*TestDB, *server.Server, func()) {
	t.Helper()

	logger := zerolog.Nop() // Silent logger for tests

	testDB, dbCleanup := SetupTestDB(t)

	testServer := CreateTestServer(&logger, testDB)

	// by default tests don't have a JobService; allow attaching a mock enqueuer
	// later via AttachMockEnqueuer

	cleanup := func() {
		if testDB.Pool != nil {
			testDB.Pool.Close()
		}

		dbCleanup()
	}

	return testDB, testServer, cleanup
}

// AttachMockEnqueuer attaches a MockEnqueuer to the provided server and
// returns the mock so tests can assert enqueued tasks.
func AttachMockEnqueuer(s *server.Server, m *mocks.MockEnqueuer) {
	if s == nil || m == nil {
		return
	}
	// Create a minimal JobService with the mock as its Client so handlers
	// that check s.Job.Client can call Enqueue without touching Redis.
	s.Job = &job.JobService{Client: m}
}

// MustMarshalJSON marshals an object to JSON or fails the test
func MustMarshalJSON(t *testing.T, v interface{}) []byte {
	t.Helper()

	jsonBytes, err := json.Marshal(v)
	require.NoError(t, err, "failed to marshal to JSON")

	return jsonBytes
}

// ProjectRoot returns the absolute path to the project root
func ProjectRoot(t *testing.T) string {
	t.Helper()

	dir, err := os.Getwd()
	require.NoError(t, err, "failed to get working directory")

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}

		parentDir := filepath.Dir(dir)
		if parentDir == dir {
			t.Fatal("could not find project root (go.mod)")
			return ""
		}

		dir = parentDir
	}
}

// Ptr returns a pointer to the given value
// Useful for creating pointers to values for optional fields
func Ptr[T any](v T) *T {
	return &v
}
