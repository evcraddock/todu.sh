# Makefile for todu.sh

# Variables
BINARY_NAME=todu
MAIN_PATH=./cmd/todu
GO=go
GOFLAGS=
LDFLAGS=

# Build information
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')

.PHONY: help
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*##"; printf ""} /^[a-zA-Z_-]+:.*?##/ { printf "  %-15s %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

.PHONY: build
build: ## Build the binary
	$(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BINARY_NAME) $(MAIN_PATH)

.PHONY: install
install: ## Install the binary to GOPATH/bin
	$(GO) install $(GOFLAGS) -ldflags "$(LDFLAGS)" $(MAIN_PATH)

.PHONY: clean
clean: ## Remove build artifacts
	rm -f $(BINARY_NAME)
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

.DEFAULT_GOAL := help
