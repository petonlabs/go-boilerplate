package dspy_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/sriniously/go-boilerplate/internal/dspy"
)

func TestDspyPing(t *testing.T) {
	if os.Getenv("DSPY_ENABLED") != "true" {
		t.Skip("DSPY not enabled")
	}
	c, err := dspy.New()
	if err != nil {
		t.Skipf("skipping: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := c.Ping(ctx); err != nil {
		t.Fatalf("ping failed: %v", err)
	}
}
