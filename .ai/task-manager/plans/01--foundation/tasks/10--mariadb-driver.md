---
id: 10
group: "database"
dependencies: [9]
status: "pending"
created: "2026-01-15"
skills:
  - go
  - database
---
# MariaDB/MySQL Driver

## Objective
Implement the MariaDB/MySQL driver for the database interface. This enables Replica to use MariaDB (MySQL-compatible) as a production database backend.

## Skills Required
- go: Go programming with database/sql
- database: MariaDB/MySQL database operations

## Acceptance Criteria
- [ ] MariaDB driver implements the Database interface
- [ ] Connection string parsing works correctly
- [ ] Connection pooling is configurable
- [ ] Health check (Ping) works correctly
- [ ] Transaction support is implemented

Use your internal Todo tool to track these and keep on track.

## Technical Requirements
- MariaDB driver: `github.com/go-sql-driver/mysql`
- Standard `database/sql` interface
- DSN format: `user:password@tcp(host:port)/database`

## Input Dependencies
- Task 10: Database interface definition

## Output Artifacts
- `internal/database/mysql/mysql.go` - MariaDB implementation
- `internal/database/mysql/mysql_test.go` - Unit tests

## Implementation Notes

<details>
<summary>Detailed Implementation Guide</summary>

### MariaDB Implementation

Create `internal/database/mysql/mysql.go`:

```go
package mysql

import (
    "context"
    "database/sql"
    "fmt"
    "time"

    _ "github.com/go-sql-driver/mysql"
    "github.com/rs/zerolog"

    "github.com/deviantintegral/replica/internal/database"
)

type MariaDB struct {
    db     *sql.DB
    logger zerolog.Logger
}

type mariadbTx struct {
    tx *sql.Tx
}

// New creates a new MariaDB database connection
func New(cfg database.Config, logger zerolog.Logger) (*MariaDB, error) {
    dsn := cfg.URL
    if dsn == "" {
        return nil, fmt.Errorf("database URL is required for mariadb")
    }

    // Append required parameters if not present
    // parseTime=true for time.Time scanning
    if !containsParam(dsn, "parseTime") {
        if containsRune(dsn, '?') {
            dsn += "&parseTime=true"
        } else {
            dsn += "?parseTime=true"
        }
    }

    db, err := sql.Open("mysql", dsn)
    if err != nil {
        return nil, fmt.Errorf("opening mariadb database: %w", err)
    }

    // Configure connection pool
    if cfg.MaxOpenConns > 0 {
        db.SetMaxOpenConns(cfg.MaxOpenConns)
    } else {
        db.SetMaxOpenConns(25)
    }
    if cfg.MaxIdleConns > 0 {
        db.SetMaxIdleConns(cfg.MaxIdleConns)
    } else {
        db.SetMaxIdleConns(5)
    }
    if cfg.ConnMaxLifetime > 0 {
        db.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Second)
    } else {
        db.SetConnMaxLifetime(5 * time.Minute)
    }

    // Verify connection
    if err := db.PingContext(context.Background()); err != nil {
        db.Close()
        return nil, fmt.Errorf("connecting to mariadb database: %w", err)
    }

    logger.Info().Msg("connected to mariadb database")

    return &MariaDB{db: db, logger: logger}, nil
}

func containsParam(dsn, param string) bool {
    return containsSubstring(dsn, param+"=")
}

func containsSubstring(s, substr string) bool {
    return len(s) >= len(substr) && findSubstring(s, substr) >= 0
}

func findSubstring(s, substr string) int {
    for i := 0; i <= len(s)-len(substr); i++ {
        if s[i:i+len(substr)] == substr {
            return i
        }
    }
    return -1
}

func containsRune(s string, r rune) bool {
    for _, c := range s {
        if c == r {
            return true
        }
    }
    return false
}

func (m *MariaDB) DB() *sql.DB {
    return m.db
}

func (m *MariaDB) Migrate() error {
    // Migration implementation will be added in Task 12
    return nil
}

func (m *MariaDB) Ping(ctx context.Context) error {
    return m.db.PingContext(ctx)
}

func (m *MariaDB) Close() error {
    m.logger.Info().Msg("closing mariadb database connection")
    return m.db.Close()
}

func (m *MariaDB) Begin(ctx context.Context) (database.Tx, error) {
    tx, err := m.db.BeginTx(ctx, nil)
    if err != nil {
        return nil, fmt.Errorf("beginning transaction: %w", err)
    }
    return &mariadbTx{tx: tx}, nil
}

func (m *MariaDB) Driver() string {
    return "mariadb"
}

func (t *mariadbTx) Commit() error {
    return t.tx.Commit()
}

func (t *mariadbTx) Rollback() error {
    return t.tx.Rollback()
}

func (t *mariadbTx) Tx() *sql.Tx {
    return t.tx
}
```

### Unit Tests

Create `internal/database/mysql/mysql_test.go`.

Note: Full integration tests require a running MariaDB instance. Unit tests should focus on:
- DSN parameter handling
- Configuration validation
- Error handling for invalid connections

Integration tests against real MariaDB will be covered in Task 15 (CI Database Matrix).

### Test with Docker

For local testing:
```bash
docker compose up -d mariadb
# Wait for health check
go test ./internal/database/mysql/... -tags=integration
```

### Verification

```bash
go test ./internal/database/mysql/...
```

</details>
