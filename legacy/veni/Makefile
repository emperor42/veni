# VENI Makefile
# Build and development commands for the VENI middleware project

# Binary output directory and name
BINARY_DIR=bin
BINARY_NAME=veni

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=gofmt

# Package paths
PKG_PATH=./pkg/veni
EXAMPLE_PATH=./examples

# Build flags
VERSION?=0.1.0
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS=-ldflags "-X main.Version=$$(VERSION) -X main.BuildTime=$$(BUILD_TIME)"

# Colors for output
RED=\033[0;31m
GREEN=\033[0;32m
YELLOW=\033[0;33m
CYAN=\033[0;36m
NC=\033[0m # No Color

.PHONY: all build run test clean fmt lint help install deps

# Default target
all: deps build

## build: Build the example server binary
build:
	@echo "$$(CYAN)Building VENI example server...$$(NC)"
	@mkdir -p $(BINARY_DIR)
	$(GOBUILD) $(LDFLAGS) -o $$(BINARY_DIR)/$$(BINARY_NAME) $(EXAMPLE_PATH)
	@echo "$(GREEN)Build complete: $$(BINARY_DIR)/$$(BINARY_NAME)$(NC)"

## build-linux: Build for Linux (cross-compile)
build-linux:
	@echo "$$(CYAN)Building for Linux...$$(NC)"
	@mkdir -p $(BINARY_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $$(BINARY_DIR)/$$(BINARY_NAME)-linux $(EXAMPLE_PATH)
	@echo "$$(GREEN)Linux build complete$$(NC)"

## build-windows: Build for Windows (cross-compile)
build-windows:
	@echo "$$(CYAN)Building for Windows...$$(NC)"
	@mkdir -p $(BINARY_DIR)
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $$(BINARY_DIR)/$$(BINARY_NAME).exe $(EXAMPLE_PATH)
	@echo "$$(GREEN)Windows build complete$$(NC)"

## build-darwin: Build for macOS (cross-compile)
build-darwin:
	@echo "$$(CYAN)Building for macOS...$$(NC)"
	@mkdir -p $(BINARY_DIR)
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $$(BINARY_DIR)/$$(BINARY_NAME)-darwin $(EXAMPLE_PATH)
	@echo "$$(GREEN)macOS build complete$$(NC)"

## build-all: Build for all platforms
build-all: build-linux build-windows build-darwin
	@echo "$$(GREEN)All platform builds complete$$(NC)"

## run: Run the example server
run:
	@echo "$$(CYAN)Starting VENI example server...$$(NC)"
	cd $(EXAMPLE_PATH) && $(GOCMD) run main.go

## test: Run all tests with verbose output
test:
	@echo "$$(CYAN)Running tests...$$(NC)"
	$(GOTEST) -v -race -coverprofile=coverage.out $(PKG_PATH)
	@echo "$$(GREEN)Tests complete$$(NC)"

## test-coverage: Run tests and generate coverage report
test-coverage: test
	@echo "$$(CYAN)Generating coverage report...$$(NC)"
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "$$(GREEN)Coverage report: coverage.html$$(NC)"

## clean: Remove build artifacts and generated files
clean:
	@echo "$$(YELLOW)Cleaning build artifacts...$$(NC)"
	@rm -rf $(BINARY_DIR)
	@rm -f coverage.out coverage.html
	$(GOCLEAN)
	@echo "$$(GREEN)Clean complete$$(NC)"

## fmt: Format all Go source files
fmt:
	@echo "$$(CYAN)Formatting Go files...$$(NC)"
	$(GOFMT) -w -s .
	@echo "$$(GREEN)Format complete$$(NC)"

## lint: Run the linter (requires golangci-lint installed)
lint:
	@echo "$$(CYAN)Running linter...$$(NC)"
	@which golangci-lint > /dev/null || (echo "$$(RED)golangci-lint not installed. Install with:$$(NC)" && echo "curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b \$$(go env GOPATH)/bin v1.55.2" && exit 1)
	golangci-lint run ./...
	@echo "$$(GREEN)Lint complete$$(NC)"

## vet: Run go vet on all packages
vet:
	@echo "$$(CYAN)Running go vet...$$(NC)"
	$(GOCMD) vet ./...
	@echo "$$(GREEN)Vet complete$$(NC)"

## deps: Download and tidy module dependencies
deps:
	@echo "$$(CYAN)Downloading dependencies...$$(NC)"
	$(GOMOD) download
	$(GOMOD) tidy
	@echo "$$(GREEN)Dependencies ready$$(NC)"

## install: Install the VENI package to GOPATH
install:
	@echo "$$(CYAN)Installing VENI package...$$(NC)"
	$(GOCMD) install .
	@echo "$$(GREEN)Install complete$$(NC)"

## check: Run all checks (fmt, vet, test)
check: fmt vet test
	@echo "$$(GREEN)All checks passed$$(NC)"

## ci: Run CI pipeline (deps, fmt-check, vet, test)
ci: deps fmt-check vet test
	@echo "$$(GREEN)CI pipeline complete$$(NC)"

## fmt-check: Check formatting without modifying files
fmt-check:
	@echo "$$(CYAN)Checking formatting...$$(NC)"
	@files=$$(gofmt -l .); \
	if [ -n "$$files" ]; then \
		echo "$$(RED)The following files need formatting:$$(NC)"; \
		echo "$$files"; \
		exit 1; \
	fi
	@echo "$$(GREEN)Formatting check passed$$(NC)"

## dev: Start development server with auto-reload (requires air)
dev:
	@echo "$$(CYAN)Starting development server...$$(NC)"
	@which air > /dev/null || (echo "$$(RED)air not installed. Install with:$$(NC)" && echo "go install github.com/cosmtrek/air@latest" && exit 1)
	cd $(EXAMPLE_PATH) && air

## help: Display this help message
help:
	@echo ""
	@echo "$$(CYAN)VENI - Web Component Middleware$$(NC)"
	@echo ""
	@echo "$$(YELLOW)Usage:$$(NC)"
	@echo "  make [target]"
	@echo ""
	@echo "$$(YELLOW)Targets:$$(NC)"
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## /  /' | column -t -s ':'
	@echo ""

# Print build info
info:
	@echo "$$(CYAN)Build Information:$$(NC)"
	@echo "  Version:    $(VERSION)"
	@echo "  Build Time: $(BUILD_TIME)"
	@echo "  Go Version: $(shell go version)"
	@echo "  Platform:   $(shell uname -s) $(shell uname -m)"