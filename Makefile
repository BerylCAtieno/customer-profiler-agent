.PHONY: build run test test-all test-health test-agent test-profile clean help

# Build the main server
build:
	@echo "Building server..."
	@go build -o bin/server cmd/server/main.go
	@echo "✓ Server built successfully: bin/server"

# Build the test client
build-test:
	@echo "Building test client..."
	@go build -o bin/test-client cmd/test/main.go
	@echo "✓ Test client built successfully: bin/test-client"

# Build both server and test client
build-all: build build-test

# Run the server
run:
	@echo "Starting Customer Profiler Agent..."
	@go run cmd/server/main.go

# Run all tests
test-all: build-test
	@echo "Running all tests..."
	@./bin/test-client -test all

# Test health endpoint
test-health: build-test
	@./bin/test-client -test health

# Test agent card endpoint
test-agent: build-test
	@./bin/test-client -test agent-card

# Test profile generation
test-profile: build-test
	@./bin/test-client -test profile

# Test with custom business idea
test-custom: build-test
	@./bin/test-client -test custom -idea "$(IDEA)"

# Quick test without building
quick-test:
	@go run cmd/test/main.go -test all

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
	@echo "✓ Clean complete"

# Install dependencies
deps:
	@echo "Installing dependencies..."
	@go mod download
	@go get -u github.com/gin-gonic/gin
	@echo "✓ Dependencies installed"

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...
	@echo "✓ Code formatted"

# Run linter
lint:
	@echo "Running linter..."
	@golangci-lint run || echo "Install golangci-lint: https://golangci-lint.run/usage/install/"

# Show help
help:
	@echo "Customer Profiler Agent - Makefile Commands"
	@echo ""
	@echo "Building:"
	@echo "  make build          Build the server"
	@echo "  make build-test     Build the test client"
	@echo "  make build-all      Build both server and test client"
	@echo ""
	@echo "Running:"
	@echo "  make run            Run the server"
	@echo ""
	@echo "Testing:"
	@echo "  make test-all       Run all tests"
	@echo "  make test-health    Test health endpoint"
	@echo "  make test-agent     Test agent card endpoint"
	@echo "  make test-profile   Test profile generation"
	@echo "  make test-custom    Test with custom idea (use IDEA='your idea')"
	@echo "  make quick-test     Run tests without building"
	@echo ""
	@echo "Maintenance:"
	@echo "  make clean          Remove build artifacts"
	@echo "  make deps           Install dependencies"
	@echo "  make fmt            Format code"
	@echo "  make lint           Run linter"
	@echo ""
	@echo "Example:"
	@echo "  make test-custom IDEA='A mobile app for pet owners'"