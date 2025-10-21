package mailcatcher

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"
)

var (
	globalServer *Server
	globalMu     sync.Mutex
)

// DefaultServer returns the global mail catcher server or starts it if not running.
// This is useful for integration tests where you want a shared instance.
// Ports can be configured via environment variables:
//   - MAILCATCHER_SMTP_PORT (default: 1025)
//   - MAILCATCHER_HTTP_PORT (default: 8025)
func DefaultServer() (*Server, error) {
	globalMu.Lock()
	defer globalMu.Unlock()

	if globalServer != nil {
		return globalServer, nil
	}

	// Get ports from environment
	smtpPort := getEnvInt("MAILCATCHER_SMTP_PORT", 1025)
	httpPort := getEnvInt("MAILCATCHER_HTTP_PORT", 8025)

	server := New(smtpPort, httpPort)
	if err := server.Start(); err != nil {
		return nil, fmt.Errorf("failed to start mail catcher: %w", err)
	}

	globalServer = server
	return globalServer, nil
}

// StopDefault stops the global mail catcher server.
func StopDefault() error {
	globalMu.Lock()
	defer globalMu.Unlock()

	if globalServer == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := globalServer.Stop(ctx)
	globalServer = nil
	return err
}

// Helper functions

func getEnvInt(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	intValue, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}

	return intValue
}
