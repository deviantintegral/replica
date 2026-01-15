---
id: 3
group: "project-foundation"
dependencies: [2]
status: "pending"
created: "2026-01-14"
skills:
  - make
  - go
---
# Makefile Build System

## Objective
Create a Makefile with standard targets for building, testing, linting, and formatting the Go project.

## Skills Required
- **make**: Makefile syntax and targets
- **go**: Go build commands and tooling

## Acceptance Criteria
- [ ] `make build` produces static binary with version info via ldflags
- [ ] `make test` runs all tests with race detection
- [ ] `make lint` runs golangci-lint
- [ ] `make fmt` formats code with gofmt
- [ ] `make clean` removes build artifacts
- [ ] `make help` lists available targets

Use your internal Todo tool to track these and keep on track.

## Technical Requirements
- Use golangci-lint for linting
- Binary output to `./replica` (or `./replica.exe` on Windows)
- Version injection via ldflags
- Race detection enabled for tests

## Input Dependencies
- Task 2: Cobra CLI with version command (provides build target structure)

## Output Artifacts
- `Makefile` with targets: `build`, `test`, `lint`, `fmt`, `clean`, `help`
- `.golangci.yml` configuration file

## Implementation Notes

<details>
<summary>Detailed Implementation Instructions</summary>

1. **Create Makefile**:
   ```makefile
   .PHONY: build test lint fmt clean help

   # Build variables
   VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
   COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
   DATE ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
   LDFLAGS := -ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)"

   # Binary name
   BINARY := replica

   ## build: Build the replica binary
   build:
   	go build $(LDFLAGS) -o $(BINARY) ./cmd/replica

   ## test: Run tests with race detection
   test:
   	go test -race -cover ./...

   ## lint: Run golangci-lint
   lint:
   	golangci-lint run ./...

   ## fmt: Format code
   fmt:
   	gofmt -s -w .

   ## clean: Remove build artifacts
   clean:
   	rm -f $(BINARY)
   	rm -f coverage.out coverage.html

   ## help: Show this help message
   help:
   	@echo "Available targets:"
   	@sed -n 's/^## //p' $(MAKEFILE_LIST) | column -t -s ':'
   ```

2. **Create .golangci.yml**:
   ```yaml
   run:
     timeout: 5m
     go: "1.25"

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

   linters-settings:
     gofmt:
       simplify: true
     misspell:
       locale: US

   issues:
     exclude-use-default: false
     max-issues-per-linter: 0
     max-same-issues: 0
   ```

3. **Verify targets**:
   ```bash
   make help
   make fmt
   make lint
   make build
   make test
   ./replica version
   make clean
   ```

</details>
