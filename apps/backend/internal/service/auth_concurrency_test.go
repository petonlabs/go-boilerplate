package service

import (
	"sync"
	"testing"
	"time"

	testhelpers "github.com/petonlabs/go-boilerplate/internal/testing"
)

func TestRotateSecretsConcurrency(t *testing.T) {
	// short-running concurrency test: rotate secrets while readers compute digests
	_, testServer, cleanup := testhelpers.SetupTest(t)
	defer cleanup()

	svc := NewAuthService(testServer)

	// initialize with known secrets
	err := svc.RotateTokenHMACSecrets("s1,s2,s3", "test")
	if err != nil {
		t.Fatalf("failed to set initial secrets: %v", err)
	}

	var wg sync.WaitGroup
	stop := make(chan struct{})
	// start readers
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-stop:
					return
				default:
					_ = svc // exercise ResetPassword digest path indirectly by calling GetTokenSecrets
					_ = svc.GetTokenSecrets()
					// small sleep to yield
					time.Sleep(1 * time.Millisecond)
				}
			}
		}()
	}

	// start writer rotator
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 20; i++ {
			err := svc.RotateTokenHMACSecrets("a"+time.Now().Format("150405.000")+",b", "test")
			if err != nil {
				t.Logf("rotate error: %v", err)
			}
			// small pause
			time.Sleep(2 * time.Millisecond)
		}
		close(stop)
	}()

	wg.Wait()
}
