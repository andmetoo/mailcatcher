# Release Notes - v1.0.0

## 🎉 Initial Release

mailcatcher - SMTP mail catcher for testing, available as **Go library** and **standalone application**.

## ✨ Features

### Core Functionality
- ✅ SMTP server on port 1025 (configurable)
- ✅ HTTP REST API on port 8025 (configurable)
- ✅ Thread-safe email storage
- ✅ Email subject parsing from headers
- ✅ CORS-enabled API for web UIs

### Dual Mode
- ✅ **Go Library**: Perfect for integration tests
- ✅ **Standalone App**: Works for any language/framework

### Deployment Options
- ✅ Binary releases for Linux, macOS, Windows (amd64, arm64)
- ✅ Docker images (multi-arch: amd64, arm64)
- ✅ Docker Compose support
- ✅ Easy installation via `go install`

### Developer Experience
- ✅ Zero configuration needed
- ✅ Environment variable support
- ✅ Optional custom logging
- ✅ Comprehensive documentation
- ✅ 72%+ test coverage

## 📦 Installation

### As Go Library
```bash
go get github.com/andmetoo/mailcatcher
```

### As Standalone Application
```bash
# Binary
wget https://github.com/andmetoo/mailcatcher/releases/download/v1.0.0/mailcatcher_Linux_x86_64.tar.gz

# Go install
go install github.com/andmetoo/mailcatcher/cmd/mailcatcher@v1.0.0

# Docker
docker pull ghcr.io/andmetoo/mailcatcher:v1.0.0
```

## 🚀 Quick Start

### Library Usage
```go
server, _ := mailcatcher.DefaultServer()
defer mailcatcher.StopDefault()

// Your app sends email to localhost:1025
// ...

emails := server.Emails()
fmt.Println(emails[0].Subject)
```

### Standalone
```bash
mailcatcher
# SMTP: localhost:1025
# API: http://localhost:8025/api/v1/emails
```

## 📚 API Endpoints

- `GET /api/v1/emails` - List all emails
- `GET /api/v1/emails/{id}` - Get specific email
- `DELETE /api/v1/emails` - Clear all emails

## 🔧 Configuration

### Environment Variables
- `MAILCATCHER_SMTP_PORT` - SMTP port (default: 1025)
- `MAILCATCHER_HTTP_PORT` - HTTP port (default: 8025)

### Command Line
```bash
mailcatcher -smtp-port 2525 -http-port 8080 -verbose
```

## 🏗️ Architecture

- **Language**: Go 1.22+
- **SMTP**: github.com/emersion/go-smtp
- **Storage**: In-memory (thread-safe)
- **HTTP**: Standard library with CORS middleware

## 📊 Test Coverage

- Total: 72.2%
- 8 integration tests
- All tests passing

## 🐳 Docker

Multi-architecture support:
- `ghcr.io/andmetoo/mailcatcher:latest`
- `ghcr.io/andmetoo/mailcatcher:v1.0.0`
- Platforms: linux/amd64, linux/arm64

## 🔄 CI/CD

Automated via GitHub Actions:
- ✅ Testing on Go 1.22, 1.23
- ✅ Linting with golangci-lint
- ✅ Multi-platform Docker builds
- ✅ Automated releases with GoReleaser

## 📝 What's Changed

### API Design
- Clean, idiomatic Go API
- `New()`, `NewWithDefaults()` constructors
- `DefaultServer()` for singleton pattern
- `Emails()`, `Email(id)` for retrieval
- `Clear()` for cleanup
- `SetLogger()` for custom logging

### HTTP API
- RESTful design
- JSON responses
- CORS enabled
- Proper HTTP status codes

### Environment Variables
- `MAILCATCHER_*` prefix (consistent naming)
- Optional configuration
- Defaults work out of box

## 🎯 Use Cases

1. **Go Integration Tests**: Embed in test suite
2. **Local Development**: Test emails locally
3. **CI/CD Pipelines**: Automated email testing
4. **Non-Go Projects**: Use standalone binary
5. **Docker Environments**: Use container image

## 📖 Documentation

- README.md with comprehensive examples
- godoc for all public APIs
- Docker examples
- Environment variable documentation

## 🙏 Acknowledgments

Inspired by:
- [MailHog](https://github.com/mailhog/MailHog)
- [MailCatcher](https://mailcatcher.me/)

Unique selling point: **Dual mode** (library + standalone)

## 📜 License

MIT License

## 🔗 Links

- GitHub: https://github.com/andmetoo/mailcatcher
- Documentation: https://pkg.go.dev/github.com/andmetoo/mailcatcher
- Docker Hub: https://ghcr.io/andmetoo/mailcatcher
- Issues: https://github.com/andmetoo/mailcatcher/issues

---

**Ready for production use!**
