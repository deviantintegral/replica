# Makefile for replica - API replication proxy
# https://github.com/deviantintegral/replica

# Build configuration
BINARY_NAME := replica
MODULE_PATH := github.com/deviantintegral/replica
VERSION_PKG := $(MODULE_PATH)/internal/version
OUTPUT_DIR := bin

# Version information from git
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Go build settings
GO := go
CGO_ENABLED := 1
GOFLAGS := -mod=readonly
LDFLAGS := -X $(VERSION_PKG).Version=$(VERSION) \
           -X $(VERSION_PKG).Commit=$(COMMIT) \
           -X $(VERSION_PKG).BuildDate=$(BUILD_DATE)

# Docker settings
DOCKER_IMAGE := replica
DOCKER_TAG := $(VERSION)

# Default target
.DEFAULT_GOAL := help

.PHONY: help all build test lint fmt docker clean coverage coverage-html mutation mutation-html deps

## help: Show available targets
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^## ' $(MAKEFILE_LIST) | sed -e 's/^## /  /'

## deps: Download and verify dependencies
deps:
	$(GO) mod download
	$(GO) mod verify

## all: Build the binary (alias for build)
all: build

## build: Build the replica binary with version info via ldflags
build:
	@mkdir -p $(OUTPUT_DIR)
	CGO_ENABLED=$(CGO_ENABLED) $(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(OUTPUT_DIR)/$(BINARY_NAME) ./cmd/replica

## test: Run all unit tests with race detection and coverage
test:
	CGO_ENABLED=$(CGO_ENABLED) $(GO) test $(GOFLAGS) -race -cover ./...

## lint: Run golangci-lint
lint:
	golangci-lint run

## fmt: Format Go code with gofmt and run go mod tidy
fmt:
	$(GO) fmt ./...
	$(GO) mod tidy

## docker: Build Docker image
docker:
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .
	docker tag $(DOCKER_IMAGE):$(DOCKER_TAG) $(DOCKER_IMAGE):latest

## clean: Remove build artifacts
clean:
	rm -rf $(OUTPUT_DIR)
	$(GO) clean -cache -testcache

## coverage: Run tests with coverage and check 80% threshold
coverage:
	@mkdir -p $(OUTPUT_DIR)
	CGO_ENABLED=$(CGO_ENABLED) $(GO) test $(GOFLAGS) -race -coverprofile=$(OUTPUT_DIR)/coverage.out -covermode=atomic ./...
	@COVERAGE=$$($(GO) tool cover -func=$(OUTPUT_DIR)/coverage.out | grep total | awk '{print $$3}' | sed 's/%//'); \
	echo "Total coverage: $${COVERAGE}%"; \
	THRESHOLD=80; \
	if [ "$$(echo "$$COVERAGE < $$THRESHOLD" | bc -l)" -eq 1 ]; then \
		echo "Coverage $${COVERAGE}% is below threshold $${THRESHOLD}%"; \
		exit 1; \
	fi; \
	echo "Coverage $${COVERAGE}% meets threshold $${THRESHOLD}%"

## coverage-html: Generate HTML coverage report
coverage-html:
	@mkdir -p $(OUTPUT_DIR)
	CGO_ENABLED=$(CGO_ENABLED) $(GO) test $(GOFLAGS) -race -coverprofile=$(OUTPUT_DIR)/coverage.out -covermode=atomic ./...
	$(GO) tool cover -html=$(OUTPUT_DIR)/coverage.out -o $(OUTPUT_DIR)/coverage.html
	$(GO) tool cover -func=$(OUTPUT_DIR)/coverage.out
	@echo ""
	@echo "Coverage report generated: $(OUTPUT_DIR)/coverage.html"

## mutation: Run gremlins mutation testing with 60% threshold
mutation:
	gremlins unleash --threshold 0.6

## mutation-html: Run gremlins mutation testing with HTML output
mutation-html:
	@mkdir -p $(OUTPUT_DIR)
	gremlins unleash --threshold 0.6 --output html --output-file $(OUTPUT_DIR)/mutation.html
	@echo ""
	@echo "Mutation report generated: $(OUTPUT_DIR)/mutation.html"
