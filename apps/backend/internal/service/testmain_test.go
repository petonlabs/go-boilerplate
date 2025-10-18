//go:build !integration
// +build !integration

package service

import (
	"os"
	"testing"

	testhelpers "github.com/petonlabs/go-boilerplate/internal/testhelpers"
)

func TestMain(m *testing.M) {
	// Setup shared container once for all tests in this package
	if err := testhelpers.SetupSharedContainer(); err != nil {
		panic(err)
	}

	// Run tests
	code := m.Run()

	// Cleanup
	testhelpers.CleanupSharedContainer()

	os.Exit(code)
}
