# Application settings
APP_NAME := uc-service
VERSION ?= 1.0.0
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%I:%M:%S%p')
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS := -ldflags="-s -w -X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME) -X main.gitCommit=$(GIT_COMMIT)"

# Directories
BIN_DIR := bin
DEPLOY_DIR := deploy
CMD_DIR := cmd/server
COVERAGE_FILE := coverage.out

# Go parameters
GOCMD := go
GOBUILD := $(GOCMD) build
GORUN := $(GOCMD) run
GOTEST := $(GOCMD) test
GOMOD := $(GOCMD) mod
GOFMT := $(GOCMD) fmt
GOVET := $(GOCMD) vet
GO_FILES := $(shell find . -name '*.go' -not -path './vendor/*' -type f)

# Tools
GOLANGCI_LINT := golangci-lint
AIR := air
SWAG := swag

.DEFAULT_GOAL := help

.PHONY: all build run dev test test-coverage clean lint fmt vet tidy deps docker docker-compose-up docker-compose-down swag help

all: clean lint test build

## build: Build the application binary
build:
	@echo "Building $(APP_NAME)..."
	@mkdir -p $(BIN_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BIN_DIR)/$(APP_NAME) ./$(CMD_DIR)
	@echo "Build complete: $(BIN_DIR)/$(APP_NAME)"

## run: Run the application directly
run:
	@echo "Running $(APP_NAME)..."
	$(GORUN) $(LDFLAGS) ./$(CMD_DIR)

## dev: Run the application with hot reload (requires air)
dev:
	@echo "Starting development server with hot reload..."
	@which $(AIR) > /dev/null || (echo "Error: air not found. Install with: go install github.com/air-verse/air@latest" && exit 1)
	$(AIR)

## test: Run all tests
test:
	@echo "Running tests..."
	$(GOTEST) -v -race -timeout 30s ./...

## test-coverage: Run tests with coverage report
test-coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -v -race -coverprofile=$(COVERAGE_FILE) -covermode=atomic ./...
	@echo "Coverage report generated: $(COVERAGE_FILE)"
	@$(GOCMD) tool cover -func=$(COVERAGE_FILE) | grep total | awk '{print "Total coverage: " $$3}'

## test-coverage-html: Generate HTML coverage report
test-coverage-html: test-coverage
	@echo "Generating HTML coverage report..."
	@$(GOCMD) tool cover -html=$(COVERAGE_FILE) -o coverage.html
	@echo "HTML coverage report: coverage.html"

## clean: Remove build artifacts and cache
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BIN_DIR)
	@rm -f $(COVERAGE_FILE) coverage.html
	@$(GOCMD) clean -cache -testcache -modcache
	@echo "Clean complete"

## fmt: Format Go source code
fmt:
	@echo "Formatting code..."
	@$(GOFMT) ./...
	@echo "Format complete"

## vet: Run go vet
vet:
	@echo "Running go vet..."
	@$(GOVET) ./...
	@echo "Vet complete"

## lint: Run golangci-lint
lint:
	@echo "Running linter..."
	@which $(GOLANGCI_LINT) > /dev/null || (echo "Error: golangci-lint not found. Install from https://golangci-lint.run/usage/install/" && exit 1)
	@$(GOLANGCI_LINT) run ./...
	@echo "Lint complete"

## tidy: Tidy and verify Go modules
tidy:
	@echo "Tidying Go modules..."
	@$(GOMOD) tidy
	@$(GOMOD) verify
	@echo "Tidy complete"

## deps: Download Go module dependencies
deps:
	@echo "Downloading dependencies..."
	@$(GOMOD) download
	@$(GOMOD) verify
	@echo "Dependencies downloaded"

## docker: Build Docker image
docker:
	@echo "Building Docker image..."
	@test -f $(DEPLOY_DIR)/Dockerfile || (echo "Error: Dockerfile not found at $(DEPLOY_DIR)/Dockerfile" && exit 1)
	docker build -f $(DEPLOY_DIR)/Dockerfile -t $(APP_NAME):$(VERSION) -t $(APP_NAME):latest .
	@echo "Docker image built: $(APP_NAME):$(VERSION)"

## docker-compose-up: Start services with docker-compose
docker-compose-up:
	@echo "Starting services with docker-compose..."
	@test -f docker-compose.yml || (echo "Warning: docker-compose.yml not found in root directory" && exit 1)
	docker-compose up -d

## docker-compose-down: Stop services with docker-compose
docker-compose-down:
	@echo "Stopping services with docker-compose..."
	@test -f docker-compose.yml || (echo "Warning: docker-compose.yml not found in root directory" && exit 1)
	docker-compose down

## swag: Generate Swagger documentation
swag:
	@echo "Generating Swagger docs..."
	@which $(SWAG) > /dev/null || (echo "Error: swag not found. Install with: go install github.com/swaggo/swag/cmd/swag@latest" && exit 1)
	$(SWAG) init -g $(CMD_DIR)/main.go -o api/docs
	@echo "Swagger docs generated"

## install: Install the application binary to GOPATH/bin
install: build
	@echo "Installing $(APP_NAME) to $(GOPATH)/bin..."
	@cp $(BIN_DIR)/$(APP_NAME) $(GOPATH)/bin/
	@echo "Install complete"

## check: Run all checks (fmt, vet, lint, test)
check: fmt vet lint test
	@echo "All checks passed!"

## help: Show this help message
help:
	@echo "$(APP_NAME) Makefile"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'
