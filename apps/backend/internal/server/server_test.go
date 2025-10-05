package server

import (
	"sync"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"

	"github.com/petonlabs/go-boilerplate/internal/config"
)

func TestGetSetConfigRaceFree(t *testing.T) {
	logger := zerolog.Nop()
	// create initial config
	cfg := &config.Config{
		Primary: config.Primary{Env: "test"},
		Server:  config.ServerConfig{Port: "8080", ReadTimeout: 5, WriteTimeout: 5, IdleTimeout: 30, CORSAllowedOrigins: []string{"*"}},
		Auth:    config.AuthConfig{TokenHMACSecret: "s1", SecretKey: "k1"},
	}

	// create server without initializing DB/Redis (avoid external dependencies)
	srv := &Server{Logger: &logger}
	srv.SetConfig(cfg)

	// start goroutines that read token secret
	var wg sync.WaitGroup
	stop := make(chan struct{})
	reader := func() {
		defer wg.Done()
		for {
			select {
			case <-stop:
				return
			default:
				_ = srv.GetTokenHMACSecret()
			}
		}
	}

	// start 5 readers
	wg.Add(5)
	for i := 0; i < 5; i++ {
		go reader()
	}

	// writer: swap config a few times
	for i := 0; i < 1000; i++ {
		newCfg := srv.GetConfig()
		if newCfg == nil {
			newCfg = &config.Config{}
		}
		// toggle token secret
		if i%2 == 0 {
			newCfg.Auth.TokenHMACSecret = "s1"
		} else {
			newCfg.Auth.TokenHMACSecret = "s2"
		}
		srv.SetConfig(newCfg)
	}

	// stop readers
	close(stop)
	wg.Wait()

	// quick sanity check
	secret := srv.GetTokenHMACSecret()
	require.NotEmpty(t, secret)

	// short sleep to let background goroutines finish cleanup if any
	time.Sleep(10 * time.Millisecond)
}
