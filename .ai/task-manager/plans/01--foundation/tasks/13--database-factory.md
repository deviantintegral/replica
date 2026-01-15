---
id: 13
group: "database"
dependencies: [12]
status: "pending"
created: "2026-01-15"
skills:
  - go
---
# Database Factory

## Objective
Create a factory function that instantiates the appropriate database driver based on configuration. This provides a single entry point for database initialization regardless of backend.

## Skills Required
- go: Go programming with factory patterns

## Acceptance Criteria
- [ ] Factory function creates correct driver based on config
- [ ] Factory applies migrations after connection
- [ ] Error handling is consistent across all backends
- [ ] Factory is integrated with CLI startup

Use your internal Todo tool to track these and keep on track.

## Technical Requirements
- Factory pattern for database instantiation
- Configuration-driven driver selection
- Automatic migration on startup (optional flag)

## Input Dependencies
- Task 13: golang-migrate integration for all drivers

## Output Artifacts
- `internal/database/factory.go` - Database factory implementation
- `internal/database/factory_test.go` - Unit tests

## Implementation Notes

<details>
<summary>Detailed Implementation Guide</summary>

### Database Factory

Create `internal/database/factory.go`:

```go
package database

import (
    "fmt"
    "strings"

    "github.com/rs/zerolog"

    "github.com/deviantintegral/replica/internal/database/mysql"
    "github.com/deviantintegral/replica/internal/database/postgres"
    "github.com/deviantintegral/replica/internal/database/sqlite"
)

// Options configures database initialization
type Options struct {
    // AutoMigrate runs migrations after connection
    AutoMigrate bool
}

// New creates a new database connection based on configuration
func New(cfg Config, logger zerolog.Logger, opts Options) (Database, error) {
    var db Database
    var err error

    driver := strings.ToLower(cfg.Driver)

    switch driver {
    case "sqlite":
        db, err = sqlite.New(cfg, logger)
    case "mariadb", "mysql":
        db, err = mysql.New(cfg, logger)
    case "postgres", "postgresql":
        db, err = postgres.New(cfg, logger)
    default:
        return nil, fmt.Errorf("unsupported database driver: %s", cfg.Driver)
    }

    if err != nil {
        return nil, fmt.Errorf("creating %s database: %w", driver, err)
    }

    if opts.AutoMigrate {
        if err := db.Migrate(); err != nil {
            db.Close()
            return nil, fmt.Errorf("running migrations: %w", err)
        }
        logger.Info().Str("driver", driver).Msg("database migrations completed")
    }

    return db, nil
}

// NewFromConfig creates a database from application config
func NewFromConfig(cfg *config.Config, logger zerolog.Logger, opts Options) (Database, error) {
    dbCfg := Config{
        Driver: cfg.Database.Driver,
        URL:    cfg.Database.URL,
    }
    return New(dbCfg, logger, opts)
}
```

Note: The `NewFromConfig` function requires importing the config package. If this creates a circular dependency, keep only the `New` function and have the caller convert config types.

### Integration Example

In `cmd/replica/root.go` or a startup command:

```go
func initDatabase(cfg *config.Config, logger zerolog.Logger) (database.Database, error) {
    dbCfg := database.Config{
        Driver: cfg.Database.Driver,
        URL:    cfg.Database.URL,
    }

    return database.New(dbCfg, logger, database.Options{
        AutoMigrate: true,
    })
}
```

### Unit Tests

Create `internal/database/factory_test.go` with tests for:
- SQLite driver selection
- MariaDB driver selection (also test "mysql" alias)
- PostgreSQL driver selection (also test "postgresql" alias)
- Invalid driver error
- AutoMigrate option behavior

### Verification

```bash
go test ./internal/database/...
```

</details>
