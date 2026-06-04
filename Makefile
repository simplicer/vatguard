.PHONY: build run test clean lint fmt help

# Project
MODULE     := github.com/antonio-villamarin/vat-validator
BINARY     := bin/vat-validator
VERSION    := $(shell cat VERSION)
BUILD_TIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
COMMIT     := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS    := -ldflags "-X main.Version=$(VERSION) -X main.Commit=$(COMMIT) -X main.BuildTime=$(BUILD_TIME)"

# Go
GO      ?= go
GOFLAGS := -trimpath

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

build: ## Build the binary
	$(GO) build $(GOFLAGS) $(LDFLAGS) -o $(BINARY) ./cmd/server

run: build ## Build and run the server
	./$(BINARY)

test: ## Run all tests
	$(GO) test -v -race -count=1 ./...

test-short: ## Run tests without race detection
	$(GO) test -count=1 ./...

test-cover: ## Run tests with coverage report
	$(GO) test -race -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html

clean: ## Remove build artifacts
	rm -f $(BINARY)
	rm -f coverage.out coverage.html

fmt: ## Format code
	$(GO) fmt ./...

lint: ## Vet the code
	$(GO) vet ./...
