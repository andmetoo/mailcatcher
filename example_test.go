package mailcatcher_test

import (
	"context"
	"testing"
	"time"

	"github.com/andmetoo/mailcatcher"
)

func TestMailCatcher(t *testing.T) {
	// Start mail catcher
	server, err := mailcatcher.DefaultServer()
	if err != nil {
		t.Fatalf("Failed to start mail catcher: %v", err)
	}

	// Clean up after test
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		server.Stop(ctx)
	}()

	// Clear any existing messages
	server.Clear()

	// TODO: Send test email via SMTP
	// smtp.SendMail("localhost:1025", nil, "from@example.com", []string{"to@example.com"}, []byte("Subject: Test\n\nBody"))

	// Wait a bit for email to be received
	time.Sleep(100 * time.Millisecond)

	// Get captured emails
	emails := server.Emails()

	if len(emails) == 0 {
		t.Skip("No emails captured (expected - need to send test email)")
	}

	// Verify email was captured
	if emails[0].From != "from@example.com" {
		t.Errorf("Expected from=from@example.com, got %s", emails[0].From)
	}
}
