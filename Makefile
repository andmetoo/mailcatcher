.PHONY: help build test clean install docker run lint coverage release

# Variables
BINARY_NAME=mailcatcher
CMD_DIR=./cmd/mailcatcher
DIST_DIR=dist
DOCKER_IMAGE=ghcr.io/andmetoo/mailcatcher
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS=-ldflags "-s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)"

# Colors for output
GREEN=\033[0;32m
NC=\033[0m # No Color

help: ## Show this help
	@echo "$(GREEN)Available targets:$(NC)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(GREEN)%-15s$(NC) %s\n", $$1, $$2}'

build: ## Build the standalone application
	@echo "$(GREEN)Building $(BINARY_NAME)...$(NC)"
	@mkdir -p $(DIST_DIR)
	go build $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME) $(CMD_DIR)
	@echo "$(GREEN)Build complete: $(DIST_DIR)/$(BINARY_NAME)$(NC)"

build-all: ## Build for all platforms
	@echo "$(GREEN)Building for all platforms...$(NC)"
	@mkdir -p $(DIST_DIR)
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-linux-amd64 $(CMD_DIR)
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-linux-arm64 $(CMD_DIR)
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-darwin-amd64 $(CMD_DIR)
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-darwin-arm64 $(CMD_DIR)
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-windows-amd64.exe $(CMD_DIR)
	@echo "$(GREEN)Multi-platform build complete: $(DIST_DIR)/$(NC)"

install: ## Install the binary to $GOPATH/bin
	@echo "$(GREEN)Installing $(BINARY_NAME)...$(NC)"
	go install $(LDFLAGS) $(CMD_DIR)
	@echo "$(GREEN)Installed to $(shell go env GOPATH)/bin/$(BINARY_NAME)$(NC)"

test: ## Run tests
	@echo "$(GREEN)Running tests...$(NC)"
	go test -v -race ./...

test-coverage: ## Run tests with coverage
	@echo "$(GREEN)Running tests with coverage...$(NC)"
	go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)Coverage report: coverage.html$(NC)"

test-short: ## Run tests without race detector (faster)
	@echo "$(GREEN)Running tests (short)...$(NC)"
	go test -v ./...

lint: ## Run linters
	@echo "$(GREEN)Running linters...$(NC)"
	golangci-lint run --timeout=5m

fmt: ## Format code
	@echo "$(GREEN)Formatting code...$(NC)"
	go fmt ./...
	golangci-lint fmt

clean: ## Clean build artifacts
	@echo "$(GREEN)Cleaning...$(NC)"
	rm -rf $(DIST_DIR)/
	rm -f coverage.out coverage.html
	go clean -cache -testcache

run: build ## Build and run the application
	@echo "$(GREEN)Running $(BINARY_NAME)...$(NC)"
	$(DIST_DIR)/$(BINARY_NAME)

run-verbose: build ## Build and run with verbose logging
	@echo "$(GREEN)Running $(BINARY_NAME) (verbose)...$(NC)"
	$(DIST_DIR)/$(BINARY_NAME) -verbose

docker-build: ## Build Docker image
	@echo "$(GREEN)Building Docker image...$(NC)"
	docker build -t $(DOCKER_IMAGE):$(VERSION) -t $(DOCKER_IMAGE):latest .

docker-run: ## Run Docker container
	@echo "$(GREEN)Running Docker container...$(NC)"
	docker run --rm -p 1025:1025 -p 8025:8025 $(DOCKER_IMAGE):latest

docker-compose-up: ## Start with docker-compose
	@echo "$(GREEN)Starting with docker-compose...$(NC)"
	docker-compose up

docker-compose-down: ## Stop docker-compose
	@echo "$(GREEN)Stopping docker-compose...$(NC)"
	docker-compose down

docker-compose-build: ## Build with docker-compose
	@echo "$(GREEN)Building with docker-compose...$(NC)"
	docker-compose build

docker-push: ## Push Docker image to registry
	@echo "$(GREEN)Pushing Docker image...$(NC)"
	docker push $(DOCKER_IMAGE):$(VERSION)
	docker push $(DOCKER_IMAGE):latest

release-dry-run: ## Run goreleaser in dry-run mode
	@echo "$(GREEN)Running goreleaser (dry-run)...$(NC)"
	goreleaser release --snapshot --clean --skip=publish

release: ## Create a release with goreleaser
	@echo "$(GREEN)Creating release...$(NC)"
	@if [ -z "$(TAG)" ]; then \
		echo "$(RED)Error: TAG is not set. Use: make release TAG=v1.0.0$(NC)"; \
		exit 1; \
	fi
	git tag -a $(TAG) -m "Release $(TAG)"
	git push origin $(TAG)
	@echo "$(GREEN)Release $(TAG) created and pushed$(NC)"

check: test lint ## Run tests and linters

ci: mod-tidy fmt vet test lint ## Run all CI checks locally

version: ## Show version information
	@echo "Version:  $(VERSION)"
	@echo "Commit:   $(COMMIT)"
	@echo "Built:    $(DATE)"

dev: ## Start development environment
	@echo "$(GREEN)Starting development environment...$(NC)"
	@echo "$(GREEN)SMTP: localhost:1025$(NC)"
	@echo "$(GREEN)HTTP: http://localhost:8025/api/v1/emails$(NC)"
	MAILCATCHER_SMTP_PORT=1025 MAILCATCHER_HTTP_PORT=8025 go run $(CMD_DIR) -verbose

deps: ## Install development dependencies
	@echo "$(GREEN)Installing development dependencies...$(NC)"
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/goreleaser/goreleaser@latest

all: clean mod-tidy fmt vet test lint build ## Run all checks and build

.DEFAULT_GOAL := help
