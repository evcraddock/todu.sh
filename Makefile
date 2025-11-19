# Makefile for todu.sh

# Variables
BUILD_DIR=.build
BINARY_NAME=todu
BINARY_PATH=$(BUILD_DIR)/$(BINARY_NAME)
MAIN_PATH=./cmd/todu
GO=go
GOFLAGS=

# Build information
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "none")

# Linker flags for version injection
LDFLAGS=-X 'github.com/evcraddock/todu.sh/cmd/todu/cmd.Version=$(VERSION)' \
        -X 'github.com/evcraddock/todu.sh/cmd/todu/cmd.Commit=$(COMMIT)' \
        -X 'github.com/evcraddock/todu.sh/cmd/todu/cmd.BuildDate=$(BUILD_TIME)'

.PHONY: help
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*##"; printf ""} /^[a-zA-Z_-]+:.*?##/ { printf "  %-15s %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

.PHONY: build
build: ## Build the binary
	@mkdir -p $(BUILD_DIR)
	$(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BINARY_PATH) $(MAIN_PATH)

.PHONY: install
install: ## Install the binary to GOPATH/bin
	$(GO) install $(GOFLAGS) -ldflags "$(LDFLAGS)" $(MAIN_PATH)

.PHONY: clean
clean: ## Remove build artifacts
	rm -rf $(BUILD_DIR)
	$(GO) clean

.PHONY: test
test: ## Run tests
	$(GO) test -v -race -coverprofile=coverage.out ./...

.PHONY: test-short
test-short: ## Run tests without race detector
	$(GO) test -v ./...

.PHONY: coverage
coverage: test ## Run tests and show coverage
	$(GO) tool cover -html=coverage.out

.PHONY: fmt
fmt: ## Format Go code
	$(GO) fmt ./...

.PHONY: vet
vet: ## Run go vet
	$(GO) vet ./...

.PHONY: lint
lint: fmt vet ## Run all linters (fmt, vet)
	@echo "Linting complete"

.PHONY: lint-ci
lint-ci: ## Run golangci-lint (requires golangci-lint installed)
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not installed. Run: brew install golangci-lint"; \
		exit 1; \
	fi

.PHONY: tidy
tidy: ## Tidy go.mod
	$(GO) mod tidy

.PHONY: verify
verify: fmt vet test ## Run format, vet, and test

.PHONY: all
all: clean lint test build ## Clean, lint, test, and build

.PHONY: dev
dev: ## Start the development environment (postgres, migration, api)
	./hack/shoreman.sh

.PHONY: dev-docker
dev-docker: ## Start the development environment using docker compose directly
	docker compose up

.PHONY: dev-logs
dev-logs: ## Tail the development logs
	@if [ -f dev.log ]; then \
		tail -f dev.log; \
	else \
		echo "No dev.log found. Start the dev environment with 'make dev'"; \
	fi

.PHONY: dev-stop
dev-stop: ## Stop the development environment
	@if [ -f .shoreman.pid ]; then \
		kill $(shell cat .shoreman.pid) 2>/dev/null || true; \
		rm -f .shoreman.pid; \
		echo "Development environment stopped"; \
	else \
		docker compose down 2>/dev/null || true; \
	fi

.PHONY: dev-clean
dev-clean: dev-stop ## Stop and remove all containers, volumes, and networks
	docker compose down -v
	rm -f dev.log dev-prev.log .shoreman.pid

.DEFAULT_GOAL := help
