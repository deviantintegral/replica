---
id: 3
group: "build-system"
dependencies: [2]
status: "completed"
created: "2026-01-15"
skills:
  - makefile
  - go
---
# Makefile Build System

## Objective
Create a Makefile with standardized build targets for development workflow automation. This provides consistent commands for building, testing, and linting the project.

## Skills Required
- makefile: GNU Make syntax and best practices
- go: Go build toolchain and flags

## Acceptance Criteria
- [ ] `make build` produces a static binary with version info injected
- [ ] `make test` runs all unit tests
- [ ] `make lint` runs golangci-lint
- [ ] `make fmt` formats all Go code with gofmt
- [ ] `make docker` builds the container image
- [ ] `make clean` removes build artifacts
- [ ] `make help` shows available targets

Use your internal Todo tool to track these and keep on track.

## Technical Requirements
- GNU Make syntax
- Static binary compilation with CGO enabled (for SQLite)
- Version info injection via ldflags
- golangci-lint for linting

## Input Dependencies
- Task 3: Cobra CLI with version command

## Output Artifacts
- `Makefile` - Build automation
- `.golangci.yml` - golangci-lint configuration

## Implementation Notes

<details>
<summary>Detailed Implementation Guide</summary>

### Makefile Structure

Create `Makefile`:

```makefile
# Build variables
BINARY_NAME := replica
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -ldflags "-X github.com/deviantintegral/replica/internal/version.Version=$(VERSION) \
    -X github.com/deviantintegral/replica/internal/version.Commit=$(COMMIT) \
    -X github.com/deviantintegral/replica/internal/version.BuildDate=$(BUILD_DATE)"

# Go variables
GOBIN := $(shell go env GOPATH)/bin
GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)

.PHONY: all build test lint fmt docker clean help

## help: Show this help message
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@sed -n 's/^## //p' $(MAKEFILE_LIST) | column -t -s ':'

## all: Build the binary (default target)
all: build

## build: Build the replica binary
build:
	CGO_ENABLED=1 go build $(LDFLAGS) -o bin/$(BINARY_NAME) ./cmd/replica

## test: Run all unit tests
test:
	go test -race -coverprofile=coverage.out ./...

## lint: Run golangci-lint
lint:
	golangci-lint run ./...

## fmt: Format Go code
fmt:
	gofmt -w .
	go mod tidy

## docker: Build Docker image
docker:
	docker build -t $(BINARY_NAME):$(VERSION) .

## clean: Remove build artifacts
clean:
	rm -rf bin/
	rm -f coverage.out

## coverage: Show test coverage report
coverage: test
	go tool cover -html=coverage.out
```

### golangci-lint Configuration

Create `.golangci.yml`:

```yaml
run:
  timeout: 5m
  modules-download-mode: readonly

linters:
  enable:
    - errcheck
    - gosimple
    - govet
    - ineffassign
    - staticcheck
    - unused
    - gofmt
    - goimports
    - misspell
    - unconvert
    - unparam
    - revive

linters-settings:
  revive:
    rules:
      - name: exported
        arguments:
          - checkPrivateReceivers
          - sayRepetitiveInsteadOfStutters

issues:
  exclude-use-default: false
  max-issues-per-linter: 0
  max-same-issues: 0
```

### Static Binary Notes

CGO_ENABLED=1 is required for go-sqlite3. For fully static Linux binaries:

```makefile
build-static:
	CGO_ENABLED=1 CC=musl-gcc go build -ldflags "-linkmode external -extldflags '-static' $(LDFLAGS)" -o bin/$(BINARY_NAME) ./cmd/replica
```

This is primarily for Docker builds; regular development can use dynamic linking.

### Verification

```bash
make help
make build
./bin/replica version
make lint
make test
```

</details>
