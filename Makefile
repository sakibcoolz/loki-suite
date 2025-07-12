# Loki Suite Makefile

.PHONY: help build run test clean docker-build docker-run deps fmt lint

# Default target
help:
	@echo "Available targets:"
	@echo "  build        - Build the application"
	@echo "  run          - Run the application"
	@echo "  test         - Run tests"
	@echo "  clean        - Clean build artifacts"
	@echo "  docker-build - Build Docker image"
	@echo "  docker-run   - Run with Docker Compose"
	@echo "  deps         - Download dependencies"
	@echo "  fmt          - Format code"
	@echo "  lint         - Run linter"

# Build the application
build:
	GOPRIVATE='github.com/sakibcoolz/*' GONOPROXY='github.com/sakibcoolz/*' GONOSUMDB='github.com/sakibcoolz/*' go build -o bin/loki-suite ./cmd

# Run the application
run:
	GOPRIVATE='github.com/sakibcoolz/*' GONOPROXY='github.com/sakibcoolz/*' GONOSUMDB='github.com/sakibcoolz/*' go run ./cmd

# Run tests
test:
	GOPRIVATE='github.com/sakibcoolz/*' GONOPROXY='github.com/sakibcoolz/*' GONOSUMDB='github.com/sakibcoolz/*' go test -v ./...

# Clean build artifacts
clean:
	rm -rf bin/
	go clean

# Build Docker image
docker-build:
	docker build -t github.com/sakibcoolz/loki-suite:latest .

# Run with Docker Compose
docker-run:
	docker-compose up -d

# Stop Docker Compose
docker-stop:
	docker-compose down

# Download dependencies
deps:
	GOPRIVATE='github.com/sakibcoolz/*' GONOPROXY='github.com/sakibcoolz/*' GONOSUMDB='github.com/sakibcoolz/*' go mod download
	GOPRIVATE='github.com/sakibcoolz/*' GONOPROXY='github.com/sakibcoolz/*' GONOSUMDB='github.com/sakibcoolz/*' go mod tidy

# Format code
fmt:
	go fmt ./...

# Run linter (requires golangci-lint)
lint:
	golangci-lint run

# Development setup
dev-setup: deps
	@echo "Setting up development environment..."
	@if ! command -v golangci-lint &> /dev/null; then \
		echo "Installing golangci-lint..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
	fi
	@if ! command -v air &> /dev/null; then \
		echo "Installing air for live reload..."; \
		go install github.com/cosmtrek/air@latest; \
	fi

# Run with live reload (requires air)
dev:
	air

# Database migrations (if needed)
migrate:
	@echo "Running database migrations..."
	go run . --migrate

# Generate Swagger docs (if swagger is added)
swagger:
	swag init

# Run security scan
security:
	gosec ./...

# Performance benchmark
benchmark:
	go test -bench=. -benchmem

# Generate coverage report
coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Check mod tidiness
check-mod:
	go mod tidy
	git diff --exit-code go.mod go.sum
