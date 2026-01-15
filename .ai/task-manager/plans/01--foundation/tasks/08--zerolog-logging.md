---
id: 8
group: "configuration"
dependencies: [7]
status: "pending"
created: "2026-01-14"
skills:
  - go
  - logging
---
# Zerolog Logging Integration

## Objective
Integrate zerolog for structured logging with configuration-driven log levels and output formats.

## Skills Required
- **go**: Go packages, context handling
- **logging**: Structured logging patterns, zerolog library

## Acceptance Criteria
- [ ] zerolog initialized based on config (level, format)
- [ ] JSON format outputs structured logs
- [ ] Text format outputs human-readable logs
- [ ] Logger accessible via package function or context
- [ ] Log levels (debug, info, warn, error) work correctly

Use your internal Todo tool to track these and keep on track.

## Technical Requirements
- Use `github.com/rs/zerolog`
- Support JSON and console (text) output formats
- Log level configuration from config file/env vars

## Input Dependencies
- Task 7: Configuration system (provides logging config)

## Output Artifacts
- `internal/logging/logging.go` - Logger initialization
- Updated CLI to initialize logging early

## Implementation Notes

<details>
<summary>Detailed Implementation Instructions</summary>

1. **Add zerolog dependency**:
   ```bash
   go get github.com/rs/zerolog
   ```

2. **Create logging package** (`internal/logging/logging.go`):
   ```go
   package logging

   import (
       "io"
       "os"
       "time"

       "github.com/rs/zerolog"
       "github.com/deviantintegral/replica/internal/config"
   )

   // Setup initializes the global logger based on configuration
   func Setup(cfg *config.LoggingConfig) zerolog.Logger {
       var output io.Writer = os.Stdout

       // Set output format
       if cfg.Format == "text" {
           output = zerolog.ConsoleWriter{
               Out:        os.Stdout,
               TimeFormat: time.RFC3339,
           }
       }

       // Set log level
       level := parseLevel(cfg.Level)

       logger := zerolog.New(output).
           With().
           Timestamp().
           Logger().
           Level(level)

       return logger
   }

   func parseLevel(level string) zerolog.Level {
       switch level {
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
   ```

3. **Update root command** to initialize logging:
   ```go
   // In cmd/replica/root.go, add PersistentPreRunE
   var cfgFile string
   var logger zerolog.Logger

   func init() {
       rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "replica.yaml", "config file path")
       rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
           cfg, err := config.Load(cfgFile)
           if err != nil {
               return err
           }
           logger = logging.Setup(&cfg.Logging)
           return nil
       }
   }
   ```

4. **Usage example**:
   ```go
   logger.Info().
       Str("component", "database").
       Str("driver", cfg.Database.Driver).
       Msg("connecting to database")

   logger.Error().
       Err(err).
       Msg("failed to connect")
   ```

</details>
