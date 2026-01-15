---
id: 1
group: "project-foundation"
dependencies: [0]
status: "completed"
created: "2026-01-15"
skills:
  - go
---
# Go Module and Project Structure

## Objective
Initialize the Go module and establish the directory structure following Go conventions. This creates the foundational layout for all Replica code.

## Skills Required
- go: Go module initialization and project organization

## Acceptance Criteria
- [ ] `go.mod` exists with module path `github.com/deviantintegral/replica` and Go 1.25+
- [ ] Directory structure matches the project layout specification
- [ ] LICENSE file contains AGPL-3.0 license text
- [ ] Basic `main.go` compiles successfully with `go build`

Use your internal Todo tool to track these and keep on track.

## Technical Requirements
- Go 1.25+ as minimum version
- Module path: `github.com/deviantintegral/replica`
- License: AGPL-3.0

## Input Dependencies
- Task 1: Development environment with Go installed

## Output Artifacts
- `go.mod` - Go module definition
- `go.sum` - Dependency checksums (initially empty)
- `LICENSE` - AGPL-3.0 license file
- Directory structure for all packages

## Implementation Notes

<details>
<summary>Detailed Implementation Guide</summary>

### Go Module Initialization

```bash
go mod init github.com/deviantintegral/replica
```

Edit `go.mod` to ensure Go version is 1.25:
```go
module github.com/deviantintegral/replica

go 1.25
```

### Directory Structure

Create the following directories:

```
replica/
├── cmd/
│   └── replica/           # Main entry point
│       └── main.go
├── internal/
│   ├── config/            # Configuration loading (Task 07)
│   ├── database/          # Database abstraction (Tasks 09-13)
│   │   ├── migrations/    # golang-migrate migrations (Task 12)
│   │   ├── sqlite/        # SQLite driver (Task 09)
│   │   ├── mysql/         # MariaDB/MySQL driver (Task 10)
│   │   └── postgres/      # PostgreSQL driver (Task 11)
│   └── testutil/          # Test utilities (Task 14)
├── docs/
│   └── adrs/              # Architecture Decision Records
├── .github/
│   └── workflows/         # CI/CD pipelines (Task 05)
└── LICENSE
```

### Main Entry Point

Create `cmd/replica/main.go` as a minimal placeholder:

```go
package main

func main() {
    // Cobra CLI setup will be added in Task 02
}
```

### LICENSE File

Download AGPL-3.0 license text from:
https://www.gnu.org/licenses/agpl-3.0.txt

Or copy from a template. The file should be named `LICENSE` (no extension).

### Verification

After creating the structure:
```bash
go build ./cmd/replica
```

This should compile without errors (producing an empty binary since main does nothing).

</details>
