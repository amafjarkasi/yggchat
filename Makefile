# Yggdrasil Mesh Chat - Makefile
# ================================

# Variables
BINARY_NAME=yggchat
VERSION?=1.0.0
BUILD_TIME=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)
GO=go
GOFLAGS=-ldflags "-s -w -X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)"

# Default target
.PHONY: all
all: clean build

# Build for current platform
.PHONY: build
build:
	$(GO) build $(GOFLAGS) -o $(BINARY_NAME)

# Build for Windows
.PHONY: build-windows
build-windows:
	GOOS=windows GOARCH=amd64 $(GO) build $(GOFLAGS) -o $(BINARY_NAME).exe

# Build for Linux
.PHONY: build-linux
build-linux:
	GOOS=linux GOARCH=amd64 $(GO) build $(GOFLAGS) -o $(BINARY_NAME)-linux

# Build for macOS (Intel)
.PHONY: build-mac-intel
build-mac-intel:
	GOOS=darwin GOARCH=amd64 $(GO) build $(GOFLAGS) -o $(BINARY_NAME)-mac-intel

# Build for macOS (Apple Silicon)
.PHONY: build-mac-arm
build-mac-arm:
	GOOS=darwin GOARCH=arm64 $(GO) build $(GOFLAGS) -o $(BINARY_NAME)-mac-arm

# Build for Raspberry Pi
.PHONY: build-pi
build-pi:
	GOOS=linux GOARCH=arm GOARM=7 $(GO) build $(GOFLAGS) -o $(BINARY_NAME)-pi

# Build for all platforms
.PHONY: build-all
build-all: build-windows build-linux build-mac-intel build-mac-arm build-pi

# Build optimized (stripped)
.PHONY: build-release
build-release:
	$(GO) build -ldflags="-s -w" -o $(BINARY_NAME)

# Run the application (Web Console)
.PHONY: run
run: build
	./$(BINARY_NAME)

# Run in TUI mode
.PHONY: run-tui
run-tui: build
	./$(BINARY_NAME) --tui

# Run with custom config
.PHONY: run-alice
run-alice: build
	./$(BINARY_NAME) --config alice.json --username Alice

# Run with custom config (Bob)
.PHONY: run-bob
run-bob: build
	./$(BINARY_NAME) --config bob.json --username Bob --port 8081

# Run tests
.PHONY: test
test:
	$(GO) test -v ./...

# Run tests with coverage
.PHONY: test-cover
test-cover:
	$(GO) test -cover ./...

# Run tests and generate coverage report
.PHONY: test-coverage
test-coverage:
	$(GO) test -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run specific test suite
.PHONY: test-crypto
test-crypto:
	$(GO) test -v -run TestECDHKeyExchange ./...

.PHONY: test-security
test-security:
	$(GO) test -v -run "TestSafeSenderName|TestSanitizeFilename|TestEscapeHTML" ./...

.PHONY: test-config
test-config:
	$(GO) test -v -run "TestConfigLoadSave|TestHistorySaveLoad" ./...

# Run benchmarks
.PHONY: bench
bench:
	$(GO) test -bench=. ./...

# Format code
.PHONY: fmt
fmt:
	$(GO) fmt ./...

# Run linter
.PHONY: lint
lint:
	golangci-lint run

# Vet code
.PHONY: vet
vet:
	$(GO) vet ./...

# Tidy dependencies
.PHONY: tidy
tidy:
	$(GO) mod tidy

# Clean build artifacts
.PHONY: clean
clean:
	rm -f $(BINARY_NAME) $(BINARY_NAME).exe $(BINARY_NAME)-*
	rm -f coverage.out coverage.html
	rm -rf downloads/*

# Docker build
.PHONY: docker-build
docker-build:
	docker build -t $(BINARY_NAME):$(VERSION) .

# Docker run
.PHONY: docker-run
docker-run:
	docker run -d --name $(BINARY_NAME) -p 8080:8080 -p 9000:9000 $(BINARY_NAME):$(VERSION)

# Docker stop
.PHONY: docker-stop
docker-stop:
	docker stop $(BINARY_NAME) && docker rm $(BINARY_NAME)

# Generate logo
.PHONY: logo
logo:
	python scripts/generate_logo.py

# Show help
.PHONY: help
help:
	@echo "Yggdrasil Mesh Chat - Makefile Commands"
	@echo "========================================"
	@echo ""
	@echo "Build Commands:"
	@echo "  make build          - Build for current platform"
	@echo "  make build-windows  - Build for Windows"
	@echo "  make build-linux    - Build for Linux"
	@echo "  make build-mac-intel - Build for macOS (Intel)"
	@echo "  make build-mac-arm  - Build for macOS (Apple Silicon)"
	@echo "  make build-pi       - Build for Raspberry Pi"
	@echo "  make build-all      - Build for all platforms"
	@echo "  make build-release  - Build optimized release"
	@echo ""
	@echo "Run Commands:"
	@echo "  make run            - Build and run (Web Console)"
	@echo "  make run-tui        - Build and run (Terminal UI)"
	@echo "  make run-alice      - Run with Alice config"
	@echo "  make run-bob        - Run with Bob config (port 8081)"
	@echo ""
	@echo "Test Commands:"
	@echo "  make test           - Run all tests"
	@echo "  make test-cover     - Run tests with coverage"
	@echo "  make test-coverage  - Generate HTML coverage report"
	@echo "  make test-crypto    - Run crypto tests"
	@echo "  make test-security  - Run security tests"
	@echo "  make test-config    - Run config tests"
	@echo "  make bench          - Run benchmarks"
	@echo ""
	@echo "Code Quality:"
	@echo "  make fmt            - Format code"
	@echo "  make lint           - Run linter"
	@echo "  make vet            - Run vet"
	@echo "  make tidy           - Tidy dependencies"
	@echo ""
	@echo "Docker Commands:"
	@echo "  make docker-build   - Build Docker image"
	@echo "  make docker-run     - Run Docker container"
	@echo "  make docker-stop    - Stop Docker container"
	@echo ""
	@echo "Other Commands:"
	@echo "  make clean          - Clean build artifacts"
	@echo "  make logo           - Regenerate logo"
	@echo "  make help           - Show this help"
