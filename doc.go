// Package mailcatcher provides an in-process SMTP mail catcher for Go testing.
//
// It captures emails sent via SMTP and exposes them through both a
// programmatic API and HTTP endpoints, making it ideal for integration
// testing of email-sending functionality.
//
// # Quick Start
//
// Use DefaultServer() for simple integration testing:
//
//	func TestEmailSending(t *testing.T) {
//	    server, err := mailcatcher.DefaultServer()
//	    if err != nil {
//	        t.Fatal(err)
//	    }
//	    defer mailcatcher.StopDefault()
//
//	    // Your code sends email to localhost:1025
//	    // ...
//
//	    // Check captured emails
//	    emails := server.Emails()
//	    if len(emails) != 1 {
//	        t.Errorf("Expected 1 email, got %d", len(emails))
//	    }
//	}
//
// # Custom Configuration
//
// Create a server with custom ports:
//
//	server := mailcatcher.New(2525, 8080)
//	if err := server.Start(); err != nil {
//	    log.Fatal(err)
//	}
//	defer server.Stop(context.Background())
//
// Or use default ports (1025 for SMTP, 8025 for HTTP):
//
//	server := mailcatcher.NewWithDefaults()
//
// # HTTP API
//
// The server exposes a REST API on port 8025 (configurable):
//
//   - GET /api/v1/emails - Returns all captured emails
//   - GET /api/v1/emails/{id} - Returns a specific email
//   - DELETE /api/v1/emails - Clears all emails
//
// Example:
//
//	curl http://localhost:8025/api/v1/emails
//
// # Features
//
//   - Thread-safe email storage
//   - Subject parsing from email headers
//   - CORS-enabled HTTP API
//   - Configurable ports
//   - Optional custom logging
package mailcatcher
