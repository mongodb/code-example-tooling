.PHONY: help build test test-unit test-webhook clean run run-dry install

# Default target
help:
	@echo "Examples Copier - Makefile"
	@echo ""
	@echo "Available targets:"
	@echo "  make build              - Build all binaries"
	@echo "  make test               - Run all tests"
	@echo "  make test-unit          - Run unit tests only"
	@echo "  make test-webhook       - Build webhook test tool"
	@echo "  make run                - Run application"
	@echo "  make run-dry            - Run in dry-run mode"
	@echo "  make run-local          - Run in local dev mode (recommended)"
	@echo "  make run-local-quick    - Quick local run (no cloud logging)"
	@echo "  make install            - Install all tools to \$$GOPATH/bin"
	@echo "  make clean              - Remove built binaries"
	@echo ""
	@echo "Testing with webhooks:"
	@echo "  make test-webhook-example  - Test with example payload"
	@echo "  make test-webhook-pr PR=123 OWNER=org REPO=repo - Test with real PR"
	@echo ""
	@echo "Quick start for local testing:"
	@echo "  make run-local-quick    # Start app (Terminal 1)"
	@echo "  make test-webhook-example  # Send test webhook (Terminal 2)"
	@echo ""

# Build all binaries
build:
	@echo "Building examples-copier..."
	@go build -o examples-copier .
	@echo "Building config-validator..."
	@go build -o config-validator ./cmd/config-validator
	@echo "Building test-webhook..."
	@go build -o test-webhook ./cmd/test-webhook
	@echo "✓ All binaries built successfully"

# Run all tests
test: test-unit
	@echo "✓ All tests passed"

# Run unit tests
test-unit:
	@echo "Running unit tests..."
	@go test ./services -v

# Run unit tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	@go test ./services -coverprofile=coverage.out
	@go tool cover -html=coverage.out -o coverage.html
	@echo "✓ Coverage report generated: coverage.html"

# Build test-webhook tool
test-webhook:
	@echo "Building test-webhook..."
	@go build -o test-webhook ./cmd/test-webhook
	@echo "✓ test-webhook built"

# Test with example payload
test-webhook-example: test-webhook
	@echo "Testing with example payload..."
	@./test-webhook -payload test-payloads/example-pr-merged.json

# Test with real PR (requires PR, OWNER, REPO variables)
test-webhook-pr: test-webhook
	@if [ -z "$(PR)" ] || [ -z "$(OWNER)" ] || [ -z "$(REPO)" ]; then \
		echo "Error: PR, OWNER, and REPO must be set"; \
		echo "Usage: make test-webhook-pr PR=123 OWNER=myorg REPO=myrepo"; \
		exit 1; \
	fi
	@./test-webhook -pr $(PR) -owner $(OWNER) -repo $(REPO)

# Test with real PR using helper script
test-pr:
	@if [ -z "$(PR)" ]; then \
		echo "Error: PR must be set"; \
		echo "Usage: make test-pr PR=123"; \
		exit 1; \
	fi
	@./scripts/test-with-pr.sh $(PR)

# Run application
run: build
	@echo "Starting examples-copier..."
	@./examples-copier

# Run in dry-run mode
run-dry: build
	@echo "Starting examples-copier in dry-run mode..."
	@DRY_RUN=true ./examples-copier

# Run in local development mode (recommended)
run-local: build
	@echo "Starting examples-copier in local development mode..."
	@./scripts/run-local.sh

# Run with cloud logging disabled (quick local testing)
run-local-quick: build
	@echo "Starting examples-copier (local, no cloud logging)..."
	@COPIER_DISABLE_CLOUD_LOGGING=true DRY_RUN=true ./examples-copier

# Validate configuration
validate: build
	@echo "Validating configuration..."
	@./examples-copier -validate

# Install binaries to $GOPATH/bin
install:
	@echo "Installing binaries..."
	@go install .
	@go install ./cmd/config-validator
	@go install ./cmd/test-webhook
	@echo "✓ Binaries installed to \$$GOPATH/bin"

# Clean built binaries
clean:
	@echo "Cleaning built binaries..."
	@rm -f examples-copier config-validator test-webhook
	@rm -f coverage.out coverage.html
	@echo "✓ Clean complete"

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...
	@echo "✓ Code formatted"

# Run linter
lint:
	@echo "Running linter..."
	@golangci-lint run ./...
	@echo "✓ Linting complete"

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy
	@echo "✓ Dependencies updated"

# Show version info
version:
	@echo "Go version:"
	@go version
	@echo ""
	@echo "Module info:"
	@go list -m

# Development setup
dev-setup: deps build
	@echo "Setting up development environment..."
	@chmod +x scripts/test-with-pr.sh
	@echo "✓ Development environment ready"

# Quick test cycle
quick-test: build test-unit
	@echo "✓ Quick test cycle complete"

# Full test cycle with webhook testing
full-test: build test-unit test-webhook-example
	@echo "✓ Full test cycle complete"

