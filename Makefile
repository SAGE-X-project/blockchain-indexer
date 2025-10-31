# Blockchain Indexer Makefile
#
# Usage:
#   make build          - Build the indexer binary
#   make test           - Run all tests
#   make test-coverage  - Run tests with coverage report
#   make lint           - Run linters
#   make fmt            - Format code
#   make clean          - Clean build artifacts
#   make help           - Show this help message

.PHONY: help build test test-unit test-integration test-e2e test-coverage \
        lint fmt vet check install-tools generate clean docker run

# Variables
PROJECT_NAME := blockchain-indexer
BINARY_NAME := indexer
BINARY_PATH := build/$(BINARY_NAME)
GO := go
GOFILES := $(shell find . -type f -name '*.go' -not -path "./vendor/*")
GOPACKAGES := $(shell go list ./... | grep -v /vendor/)

# Version info (will be injected at build time)
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')

# Build flags
LDFLAGS := -s -w \
	-X main.version=$(VERSION) \
	-X main.commit=$(COMMIT) \
	-X main.buildTime=$(BUILD_TIME)

# Test flags
TEST_FLAGS := -v -race -timeout 5m
COVERAGE_FILE := coverage.out
COVERAGE_HTML := coverage.html

# Tools
GOLANGCI_LINT := $(shell command -v golangci-lint 2> /dev/null)
GOIMPORTS := $(shell command -v goimports 2> /dev/null)

# Colors for output
COLOR_RESET := \033[0m
COLOR_BOLD := \033[1m
COLOR_GREEN := \033[32m
COLOR_YELLOW := \033[33m
COLOR_BLUE := \033[34m

##@ General

help: ## Display this help message
	@echo "$(COLOR_BOLD)$(PROJECT_NAME) - Makefile Commands$(COLOR_RESET)"
	@echo ""
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make $(COLOR_BLUE)<target>$(COLOR_RESET)\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  $(COLOR_BLUE)%-20s$(COLOR_RESET) %s\n", $$1, $$2 } /^##@/ { printf "\n$(COLOR_BOLD)%s$(COLOR_RESET)\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

fmt: ## Format Go code
	@echo "$(COLOR_GREEN)Formatting code...$(COLOR_RESET)"
	@$(GO) fmt ./...
	@if [ -n "$(GOIMPORTS)" ]; then \
		goimports -w $(GOFILES); \
	else \
		echo "$(COLOR_YELLOW)goimports not found, skipping import formatting$(COLOR_RESET)"; \
	fi

vet: ## Run go vet
	@echo "$(COLOR_GREEN)Running go vet...$(COLOR_RESET)"
	@$(GO) vet ./...

lint: ## Run linters
	@echo "$(COLOR_GREEN)Running linters...$(COLOR_RESET)"
	@if [ -n "$(GOLANGCI_LINT)" ]; then \
		golangci-lint run --timeout 5m; \
	else \
		echo "$(COLOR_YELLOW)golangci-lint not found, please run 'make install-tools'$(COLOR_RESET)"; \
		exit 1; \
	fi

check: fmt vet lint ## Run all checks (fmt, vet, lint)

##@ Build

build: ## Build the indexer binary
	@echo "$(COLOR_GREEN)Building $(BINARY_NAME)...$(COLOR_RESET)"
	@mkdir -p bin
	@$(GO) build -ldflags "$(LDFLAGS)" -o $(BINARY_PATH) ./cmd/indexer
	@echo "$(COLOR_GREEN)✓ Binary built: $(BINARY_PATH)$(COLOR_RESET)"

build-all: ## Build binaries for all platforms
	@echo "$(COLOR_GREEN)Building for all platforms...$(COLOR_RESET)"
	@mkdir -p bin
	@GOOS=linux GOARCH=amd64 $(GO) build -ldflags "$(LDFLAGS)" -o bin/$(BINARY_NAME)-linux-amd64 ./cmd/indexer
	@GOOS=linux GOARCH=arm64 $(GO) build -ldflags "$(LDFLAGS)" -o bin/$(BINARY_NAME)-linux-arm64 ./cmd/indexer
	@GOOS=darwin GOARCH=amd64 $(GO) build -ldflags "$(LDFLAGS)" -o bin/$(BINARY_NAME)-darwin-amd64 ./cmd/indexer
	@GOOS=darwin GOARCH=arm64 $(GO) build -ldflags "$(LDFLAGS)" -o bin/$(BINARY_NAME)-darwin-arm64 ./cmd/indexer
	@GOOS=windows GOARCH=amd64 $(GO) build -ldflags "$(LDFLAGS)" -o bin/$(BINARY_NAME)-windows-amd64.exe ./cmd/indexer
	@echo "$(COLOR_GREEN)✓ All platform binaries built$(COLOR_RESET)"

install: ## Install the binary to $GOPATH/bin
	@echo "$(COLOR_GREEN)Installing $(BINARY_NAME)...$(COLOR_RESET)"
	@$(GO) install -ldflags "$(LDFLAGS)" ./cmd/indexer
	@echo "$(COLOR_GREEN)✓ Installed to $(GOPATH)/bin/$(BINARY_NAME)$(COLOR_RESET)"

##@ Testing

test: ## Run all tests
	@echo "$(COLOR_GREEN)Running all tests...$(COLOR_RESET)"
	@$(GO) test $(TEST_FLAGS) ./...

test-unit: ## Run unit tests only
	@echo "$(COLOR_GREEN)Running unit tests...$(COLOR_RESET)"
	@$(GO) test $(TEST_FLAGS) -short ./...

test-integration: ## Run integration tests only
	@echo "$(COLOR_GREEN)Running integration tests...$(COLOR_RESET)"
	@$(GO) test $(TEST_FLAGS) -run Integration ./test/integration/...

test-e2e: ## Run end-to-end tests
	@echo "$(COLOR_GREEN)Running e2e tests...$(COLOR_RESET)"
	@$(GO) test $(TEST_FLAGS) ./test/e2e/...

test-coverage: ## Run tests with coverage report
	@echo "$(COLOR_GREEN)Running tests with coverage...$(COLOR_RESET)"
	@$(GO) test $(TEST_FLAGS) -coverprofile=$(COVERAGE_FILE) -covermode=atomic ./...
	@$(GO) tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML)
	@echo "$(COLOR_GREEN)✓ Coverage report: $(COVERAGE_HTML)$(COLOR_RESET)"
	@$(GO) tool cover -func=$(COVERAGE_FILE) | grep total | awk '{print "Total coverage: " $$3}'

test-race: ## Run tests with race detection
	@echo "$(COLOR_GREEN)Running tests with race detection...$(COLOR_RESET)"
	@$(GO) test -race -timeout 10m ./...

bench: ## Run benchmarks
	@echo "$(COLOR_GREEN)Running benchmarks...$(COLOR_RESET)"
	@$(GO) test -bench=. -benchmem -run=^$$ ./...

##@ Code Generation

generate: ## Generate code (mocks, protobuf, graphql)
	@echo "$(COLOR_GREEN)Generating code...$(COLOR_RESET)"
	@$(GO) generate ./...

generate-mocks: ## Generate mock implementations
	@echo "$(COLOR_GREEN)Generating mocks...$(COLOR_RESET)"
	@which mockgen > /dev/null || (echo "$(COLOR_YELLOW)Installing mockgen...$(COLOR_RESET)" && go install github.com/golang/mock/mockgen@latest)
	@$(GO) generate -run mockgen ./...

generate-proto: ## Generate protobuf code
	@echo "$(COLOR_GREEN)Generating protobuf code...$(COLOR_RESET)"
	@which protoc > /dev/null || (echo "$(COLOR_YELLOW)protoc not found, please install it$(COLOR_RESET)" && exit 1)
	@which protoc-gen-go > /dev/null || (echo "$(COLOR_YELLOW)Installing protoc-gen-go...$(COLOR_RESET)" && go install google.golang.org/protobuf/cmd/protoc-gen-go@latest)
	@which protoc-gen-go-grpc > /dev/null || (echo "$(COLOR_YELLOW)Installing protoc-gen-go-grpc...$(COLOR_RESET)" && go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest)
	@protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		api/proto/*.proto

generate-graphql: ## Generate GraphQL code
	@echo "$(COLOR_GREEN)Generating GraphQL code...$(COLOR_RESET)"
	@which gqlgen > /dev/null || (echo "$(COLOR_YELLOW)Installing gqlgen...$(COLOR_RESET)" && go install github.com/99designs/gqlgen@latest)
	@cd api/graphql && gqlgen generate

##@ Dependencies

deps: ## Download dependencies
	@echo "$(COLOR_GREEN)Downloading dependencies...$(COLOR_RESET)"
	@$(GO) mod download

deps-tidy: ## Tidy dependencies
	@echo "$(COLOR_GREEN)Tidying dependencies...$(COLOR_RESET)"
	@$(GO) mod tidy

deps-verify: ## Verify dependencies
	@echo "$(COLOR_GREEN)Verifying dependencies...$(COLOR_RESET)"
	@$(GO) mod verify

deps-update: ## Update dependencies
	@echo "$(COLOR_GREEN)Updating dependencies...$(COLOR_RESET)"
	@$(GO) get -u ./...
	@$(GO) mod tidy

##@ Tools

install-tools: ## Install development tools
	@echo "$(COLOR_GREEN)Installing development tools...$(COLOR_RESET)"
	@$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@$(GO) install golang.org/x/tools/cmd/goimports@latest
	@$(GO) install github.com/golang/mock/mockgen@latest
	@$(GO) install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	@$(GO) install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	@$(GO) install github.com/99designs/gqlgen@latest
	@echo "$(COLOR_GREEN)✓ Development tools installed$(COLOR_RESET)"

##@ Docker

docker-build: ## Build Docker image
	@echo "$(COLOR_GREEN)Building Docker image...$(COLOR_RESET)"
	@docker build -t $(PROJECT_NAME):$(VERSION) -f deployments/docker/Dockerfile .
	@docker tag $(PROJECT_NAME):$(VERSION) $(PROJECT_NAME):latest
	@echo "$(COLOR_GREEN)✓ Docker image built: $(PROJECT_NAME):$(VERSION)$(COLOR_RESET)"

docker-run: ## Run Docker container
	@echo "$(COLOR_GREEN)Running Docker container...$(COLOR_RESET)"
	@docker run --rm -it \
		-v $(PWD)/config:/config \
		-v $(PWD)/data:/data \
		-p 8080:8080 \
		-p 9090:9090 \
		-p 8081:8081 \
		$(PROJECT_NAME):latest

docker-compose-up: ## Start services with docker-compose
	@echo "$(COLOR_GREEN)Starting services with docker-compose...$(COLOR_RESET)"
	@docker-compose -f deployments/docker/docker-compose.yml up -d

docker-compose-down: ## Stop services with docker-compose
	@echo "$(COLOR_GREEN)Stopping services with docker-compose...$(COLOR_RESET)"
	@docker-compose -f deployments/docker/docker-compose.yml down

##@ Cleanup

clean: ## Clean build artifacts
	@echo "$(COLOR_GREEN)Cleaning build artifacts...$(COLOR_RESET)"
	@rm -rf bin/
	@rm -rf $(COVERAGE_FILE) $(COVERAGE_HTML)
	@rm -rf vendor/
	@find . -type f -name '*.test' -delete
	@echo "$(COLOR_GREEN)✓ Cleaned$(COLOR_RESET)"

clean-all: clean ## Clean all generated files and caches
	@echo "$(COLOR_GREEN)Cleaning all generated files...$(COLOR_RESET)"
	@$(GO) clean -cache -testcache -modcache
	@rm -rf data/
	@echo "$(COLOR_GREEN)✓ All cleaned$(COLOR_RESET)"

##@ Run

run: build ## Build and run the indexer
	@echo "$(COLOR_GREEN)Running indexer...$(COLOR_RESET)"
	@$(BINARY_PATH) --config config/config.yaml

run-dev: ## Run in development mode
	@echo "$(COLOR_GREEN)Running in development mode...$(COLOR_RESET)"
	@$(GO) run ./cmd/indexer --config config/config.yaml --log-level debug

##@ CI/CD

ci: deps check test ## Run CI pipeline (deps, check, test)

ci-full: deps check test-coverage ## Run full CI pipeline with coverage

##@ Information

version: ## Show version information
	@echo "Version:    $(VERSION)"
	@echo "Commit:     $(COMMIT)"
	@echo "Build Time: $(BUILD_TIME)"

info: ## Show project information
	@echo "$(COLOR_BOLD)Project Information$(COLOR_RESET)"
	@echo "Name:       $(PROJECT_NAME)"
	@echo "Binary:     $(BINARY_NAME)"
	@echo "Version:    $(VERSION)"
	@echo "Commit:     $(COMMIT)"
	@echo "Build Time: $(BUILD_TIME)"
	@echo ""
	@echo "$(COLOR_BOLD)Go Environment$(COLOR_RESET)"
	@$(GO) version
	@$(GO) env GOPATH GOROOT GOOS GOARCH

.DEFAULT_GOAL := help
