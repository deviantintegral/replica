---
id: 2
group: "project-foundation"
dependencies: [1]
status: "pending"
created: "2026-01-14"
skills:
  - go
  - cobra
---
# Cobra CLI and Version Command

## Objective
Implement the Cobra CLI framework with a working `replica version` command that displays version information.

## Skills Required
- **go**: Go programming and build flags
- **cobra**: Cobra CLI framework setup and command structure

## Acceptance Criteria
- [ ] Cobra root command initialized
- [ ] `replica version` command outputs version, commit, and build date
- [ ] Version information injected via ldflags at build time
- [ ] `replica --help` shows usage information

Use your internal Todo tool to track these and keep on track.

## Technical Requirements
- Use `github.com/spf13/cobra` package
- Version info via `-ldflags` at build time:
  - `main.version` (semver)
  - `main.commit` (git SHA)
  - `main.date` (build timestamp)

## Input Dependencies
- Task 1: Go module and project structure

## Output Artifacts
- `cmd/replica/main.go` with Cobra root command
- `cmd/replica/version.go` with version command
- Working `replica version` command

## Implementation Notes

<details>
<summary>Detailed Implementation Instructions</summary>

1. **Add Cobra dependency**:
   ```bash
   go get github.com/spf13/cobra@latest
   ```

2. **Create root command** (`cmd/replica/root.go`):
   ```go
   package main

   import (
       "os"
       "github.com/spf13/cobra"
   )

   var rootCmd = &cobra.Command{
       Use:   "replica",
       Short: "Replica - A distributed content management system",
       Long:  `Replica is a distributed CMS that enables bidirectional content synchronization between independent instances.`,
   }

   func Execute() {
       if err := rootCmd.Execute(); err != nil {
           os.Exit(1)
       }
   }
   ```

3. **Create version command** (`cmd/replica/version.go`):
   ```go
   package main

   import (
       "fmt"
       "github.com/spf13/cobra"
   )

   // These variables are set via ldflags at build time
   var (
       version = "dev"
       commit  = "none"
       date    = "unknown"
   )

   var versionCmd = &cobra.Command{
       Use:   "version",
       Short: "Print version information",
       Run: func(cmd *cobra.Command, args []string) {
           fmt.Printf("replica version %s\n", version)
           fmt.Printf("  commit: %s\n", commit)
           fmt.Printf("  built:  %s\n", date)
       },
   }

   func init() {
       rootCmd.AddCommand(versionCmd)
   }
   ```

4. **Update main.go** (`cmd/replica/main.go`):
   ```go
   package main

   func main() {
       Execute()
   }
   ```

5. **Build with ldflags** (for testing):
   ```bash
   go build -ldflags "-X main.version=0.1.0 -X main.commit=$(git rev-parse --short HEAD) -X main.date=$(date -u +%Y-%m-%dT%H:%M:%SZ)" -o replica ./cmd/replica
   ```

6. **Verify**:
   ```bash
   ./replica version
   ./replica --help
   ```

</details>
