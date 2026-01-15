---
id: 8
group: "configuration"
dependencies: [7]
status: "completed"
created: "2026-01-15"
skills:
  - go
---
# Zerolog Logging Integration

## Objective
Integrate zerolog for structured logging throughout the application. This provides consistent, performant logging with JSON output suitable for log aggregation systems.

## Skills Required
- go: Go programming with zerolog library

## Acceptance Criteria
- [ ] zerolog is configured based on config settings (level, format)
- [ ] Logger is initialized at application startup
- [ ] Log output includes timestamp, level, and structured fields
- [ ] Text format is available for development (pretty printing)
- [ ] Logger can be accessed throughout the application

Use your internal Todo tool to track these and keep on track.

## Technical Requirements
- zerolog library: `github.com/rs/zerolog`
- Configurable log level from config
- JSON and text output formats
- Context-based logger passing

## Input Dependencies
- Task 8: Configuration system with logging settings

## Output Artifacts
- `internal/logging/logging.go` - Logger initialization and configuration
- `internal/logging/logging_test.go` - Unit tests

## Implementation Notes

<details>
<summary>Detailed Implementation Guide</summary>

### Logging Package

Create `internal/logging/logging.go`:

```go
package logging

import (
    "io"
    "os"
    "strings"
    "time"

    "github.com/rs/zerolog"

    "github.com/deviantintegral/replica/internal/config"
)

// New creates a new logger based on configuration
func New(cfg config.LoggingConfig) zerolog.Logger {
    var output io.Writer = os.Stdout

    // Configure output format
    if strings.ToLower(cfg.Format) == "text" {
        output = zerolog.ConsoleWriter{
            Out:        os.Stdout,
            TimeFormat: time.RFC3339,
        }
    }

    // Parse log level
    level := parseLevel(cfg.Level)

    return zerolog.New(output).
        Level(level).
        With().
        Timestamp().
        Logger()
}

func parseLevel(level string) zerolog.Level {
    switch strings.ToLower(level) {
    case "debug":
        return zerolog.DebugLevel
    case "info":
        return zerolog.InfoLevel
    case "warn":
        return zerolog.WarnLevel
    case "error":
        return zerolog.ErrorLevel
    default:
        return zerolog.InfoLevel
    }
}

// WithComponent returns a logger with a component field
func WithComponent(logger zerolog.Logger, component string) zerolog.Logger {
    return logger.With().Str("component", component).Logger()
}
```

### Integration with Main

Update `cmd/replica/root.go` to initialize logger:

```go
package main

import (
    "os"

    "github.com/rs/zerolog"
    "github.com/rs/zerolog/log"
    "github.com/spf13/cobra"

    "github.com/deviantintegral/replica/internal/config"
    "github.com/deviantintegral/replica/internal/logging"
)

var (
    cfgFile string
    logger  zerolog.Logger
)

var rootCmd = &cobra.Command{
    Use:   "replica",
    Short: "Replica - A distributed CMS",
    Long:  `Replica is a distributed content management system...`,
    PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
        // Load configuration
        cfg, err := config.Load(cfgFile)
        if err != nil {
            return err
        }

        // Initialize logger
        logger = logging.New(cfg.Logging)
        log.Logger = logger // Set global logger

        return nil
    },
}

func init() {
    rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "replica.yaml", "config file path")
}

func Execute() {
    if err := rootCmd.Execute(); err != nil {
        os.Exit(1)
    }
}
```

### Usage Examples

```go
// Component-specific logger
dbLogger := logging.WithComponent(logger, "database")
dbLogger.Info().Str("driver", "sqlite").Msg("connecting to database")

// Structured logging
logger.Info().
    Str("method", "GET").
    Str("path", "/api/content").
    Int("status", 200).
    Dur("latency", duration).
    Msg("request completed")

// Error logging
logger.Error().
    Err(err).
    Str("operation", "migrate").
    Msg("migration failed")
```

### Unit Tests

Create `internal/logging/logging_test.go` with tests for:
- Logger creation with different levels
- JSON and text format output
- Component logger creation
- Level parsing edge cases

### Verification

```bash
go test ./internal/logging/...
```

</details>
