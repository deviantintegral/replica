---
id: 2
group: "project-foundation"
dependencies: [1]
status: "completed"
created: "2026-01-15"
skills:
  - go
---
# Cobra CLI and Version Command

## Objective
Implement the Cobra CLI framework with a basic `version` command. This establishes the command-line interface foundation for all future CLI commands.

## Skills Required
- go: Go programming with Cobra CLI framework

## Acceptance Criteria
- [ ] Cobra CLI is initialized and configured
- [ ] `replica version` outputs version information
- [ ] Version includes build information (version, commit, date)
- [ ] `replica --help` shows usage information
- [ ] Root command has proper metadata (use, short, long descriptions)

Use your internal Todo tool to track these and keep on track.

## Technical Requirements
- Cobra CLI framework: `github.com/spf13/cobra`
- Version info injected via ldflags at build time
- Structured output for version command

## Input Dependencies
- Task 2: Go module and project structure

## Output Artifacts
- `cmd/replica/main.go` - Updated with Cobra initialization
- `cmd/replica/root.go` - Root command definition
- `cmd/replica/version.go` - Version command implementation
- `internal/version/version.go` - Version info struct and variables

## Implementation Notes

<details>
<summary>Detailed Implementation Guide</summary>

### Add Cobra Dependency

```bash
go get github.com/spf13/cobra
```

### Version Package

Create `internal/version/version.go`:

```go
package version

// These variables are set at build time via ldflags
var (
    Version   = "dev"
    Commit    = "unknown"
    BuildDate = "unknown"
)

type Info struct {
    Version   string `json:"version"`
    Commit    string `json:"commit"`
    BuildDate string `json:"buildDate"`
}

func Get() Info {
    return Info{
        Version:   Version,
        Commit:    Commit,
        BuildDate: BuildDate,
    }
}
```

### Root Command

Create `cmd/replica/root.go`:

```go
package main

import (
    "os"

    "github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
    Use:   "replica",
    Short: "Replica - A distributed CMS",
    Long:  `Replica is a distributed content management system supporting
multiple database backends and designed for offline-first workflows.`,
}

func Execute() {
    if err := rootCmd.Execute(); err != nil {
        os.Exit(1)
    }
}
```

### Version Command

Create `cmd/replica/version.go`:

```go
package main

import (
    "fmt"

    "github.com/deviantintegral/replica/internal/version"
    "github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
    Use:   "version",
    Short: "Print version information",
    Run: func(cmd *cobra.Command, args []string) {
        info := version.Get()
        fmt.Printf("replica version %s\n", info.Version)
        fmt.Printf("  commit: %s\n", info.Commit)
        fmt.Printf("  built:  %s\n", info.BuildDate)
    },
}

func init() {
    rootCmd.AddCommand(versionCmd)
}
```

### Main Entry Point

Update `cmd/replica/main.go`:

```go
package main

func main() {
    Execute()
}
```

### Build with Version Info

The Makefile (Task 03) will inject version info:

```bash
go build -ldflags "-X github.com/deviantintegral/replica/internal/version.Version=1.0.0 \
    -X github.com/deviantintegral/replica/internal/version.Commit=$(git rev-parse HEAD) \
    -X github.com/deviantintegral/replica/internal/version.BuildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
    ./cmd/replica
```

### Verification

```bash
go build ./cmd/replica
./replica version
./replica --help
```

</details>
