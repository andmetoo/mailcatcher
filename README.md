# mailcatcher

[![Test](https://github.com/andmetoo/mailcatcher/workflows/Test/badge.svg)](https://github.com/andmetoo/mailcatcher/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/andmetoo/mailcatcher)](https://goreportcard.com/report/github.com/andmetoo/mailcatcher)
[![Go Reference](https://pkg.go.dev/badge/github.com/andmetoo/mailcatcher.svg)](https://pkg.go.dev/github.com/andmetoo/mailcatcher)
[![License](https://img.shields.io/github/license/andmetoo/mailcatcher)](LICENSE)

SMTP mail catcher for testing - works as **Go library** or **standalone server**.

## Features

- ✅ **Dual Mode**: Use as Go library or standalone application
- ✅ **SMTP Server**: Captures emails on port 1025 (configurable)
- ✅ **HTTP API**: JSON REST API on port 8025 (configurable)
- ✅ **Docker Support**: Multi-arch images (amd64, arm64)
- ✅ **Thread-Safe**: Safe for concurrent use
- ✅ **Subject Parsing**: Extracts email subject from headers
- ✅ **CORS Enabled**: Ready for web UI integration
- ✅ **Zero Config**: Works out of the box

## Installation

### As Go Library

```bash
go get github.com/andmetoo/mailcatcher
```

### As Standalone Application

**Binary releases:**
```bash
# Download from GitHub Releases
wget https://github.com/andmetoo/mailcatcher/releases/latest/download/mailcatcher_Linux_x86_64.tar.gz
tar -xzf mailcatcher_Linux_x86_64.tar.gz
./mailcatcher
```

**Using Go:**
```bash
go install github.com/andmetoo/mailcatcher/cmd/mailcatcher@latest
```

**Using Docker:**
```bash
docker run -p 1025:1025 -p 8025:8025 ghcr.io/andmetoo/mailcatcher:latest
```

**Using Docker Compose:**
```bash
docker-compose up
```

## Usage

### 1. Standalone Application

```bash
# Start with default ports (SMTP: 1025, HTTP: 8025)
mailcatcher

# Custom ports
mailcatcher -smtp-port 2525 -http-port 8080

# With verbose logging
mailcatcher -verbose

# Using environment variables
export MAILCATCHER_SMTP_PORT=2525
export MAILCATCHER_HTTP_PORT=8080
mailcatcher

# Show version
mailcatcher -version
```

### 2. Go Library (Integration Tests)

```go
package myapp_test

import (
    "testing"
    "net/smtp"
    "github.com/andmetoo/mailcatcher"
)

func TestEmailSending(t *testing.T) {
    // Start default server (ports 1025/8025)
    server, err := mailcatcher.DefaultServer()
    if err != nil {
        t.Fatal(err)
    }
    defer mailcatcher.StopDefault()

    // Clear any existing emails
    server.Clear()

    // Your app sends email to localhost:1025
    smtp.SendMail("localhost:1025", nil,
        "sender@example.com",
        []string{"recipient@example.com"},
        []byte("Subject: Test\r\n\r\nBody"))

    // Check captured emails
    emails := server.Emails()
    if len(emails) != 1 {
        t.Errorf("Expected 1 email, got %d", len(emails))
    }

    if emails[0].Subject != "Test" {
        t.Errorf("Wrong subject: %s", emails[0].Subject)
    }
}
```

### 3. Custom Configuration

```go
// Create server with custom ports
server := mailcatcher.New(2525, 8080)
err := server.Start()
if err != nil {
    log.Fatal(err)
}
defer server.Stop(context.Background())

// Use default ports (1025/8025)
server := mailcatcher.NewWithDefaults()

// Enable custom logging
server.SetLogger(log.Default())
```

### 4. Programmatic API

```go
// Get all emails
emails := server.Emails()

// Get specific email
email := server.Email("msg-0")
if email != nil {
    fmt.Println(email.Subject)
    fmt.Println(email.From)
    fmt.Println(email.To)
    fmt.Println(email.Body)
}

// Clear all emails
server.Clear()
```

## HTTP API

### GET /api/v1/emails

Returns all captured emails.

```bash
curl http://localhost:8025/api/v1/emails
```

Response:
```json
{
  "total": 2,
  "count": 2,
  "items": [
    {
      "id": "msg-0",
      "from": "sender@example.com",
      "to": ["recipient@example.com"],
      "subject": "Test Email",
      "body": "Subject: Test Email\r\n\r\nEmail body...",
      "time": "2025-01-15T10:30:00Z"
    }
  ]
}
```

### GET /api/v1/emails/{id}

Returns specific email by ID.

```bash
curl http://localhost:8025/api/v1/emails/msg-0
```

### DELETE /api/v1/emails

Clears all captured emails.

```bash
curl -X DELETE http://localhost:8025/api/v1/emails
```

## Environment Variables

```bash
# SMTP server port (default: 1025)
MAILCATCHER_SMTP_PORT=1025

# HTTP API server port (default: 8025)
MAILCATCHER_HTTP_PORT=8025
```

## Docker

### Standalone Container

```bash
docker run -d \
  -p 1025:1025 \
  -p 8025:8025 \
  --name mailcatcher \
  ghcr.io/andmetoo/mailcatcher:latest
```

### Docker Compose

```yaml
version: '3.8'

services:
  mailcatcher:
    image: ghcr.io/andmetoo/mailcatcher:latest
    ports:
      - "1025:1025"
      - "8025:8025"
    environment:
      - MAILCATCHER_SMTP_PORT=1025
      - MAILCATCHER_HTTP_PORT=8025
    restart: unless-stopped
```

## Email Structure

```go
type Email struct {
    ID      string    `json:"id"`      // Auto-generated: msg-0, msg-1, ...
    From    string    `json:"from"`    // Sender address
    To      []string  `json:"to"`      // Recipient addresses
    Subject string    `json:"subject"` // Parsed from headers
    Body    string    `json:"body"`    // Full email with headers
    Time    time.Time `json:"time"`    // Capture timestamp
}
```

## Use Cases

- ✅ Integration testing of email-sending code
- ✅ Local development without real SMTP server
- ✅ Email preview in development environment
- ✅ CI/CD pipeline testing
- ✅ Manual email testing for non-Go projects

## Similar Projects

- [MailHog](https://github.com/mailhog/MailHog) - Standalone service (Go)
- [MailCatcher](https://mailcatcher.me/) - Standalone service (Ruby)
- [smtp4dev](https://github.com/rnwood/smtp4dev) - Standalone service (.NET)

**mailcatcher** is unique as it works both as a **Go library** and **standalone application**.

## Development

### Using Makefile

```bash
# Show all available commands
make help

# Build standalone application
make build

# Run tests
make test

# Run tests with coverage
make test-coverage

# Run linters
make lint

# Format code
make fmt

# Run all checks (tests + lint)
make check

# Run full CI suite locally
make ci

# Build for all platforms
make build-all

# Clean build artifacts
make clean
```

### Manual Build

```bash
# Build library
go build ./...

# Build standalone app
go build ./cmd/mailcatcher

# Run tests
go test -v ./...

# Run with coverage
go test -v -coverprofile=coverage.out ./...
```

### Docker Development

```bash
# Build Docker image
make docker-build

# Run Docker container
make docker-run

# Start with docker-compose
make docker-compose-up
```

### Release

Releases are automated via GitHub Actions and GoReleaser:

```bash
git tag v1.0.0
git push origin v1.0.0
```

This will:
- Build binaries for Linux, macOS, Windows (amd64, arm64)
- Create Docker images for amd64 and arm64
- Publish to GitHub Releases
- Push images to GitHub Container Registry

## License

MIT License - see [LICENSE](LICENSE) file.

## Author

andmetoo

## Contributing

Pull requests welcome! Please ensure:

```bash
go test -v ./...          # Tests pass
go build ./...            # Code compiles
golangci-lint run         # Linting passes
```
