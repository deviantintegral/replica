---
id: 9
group: "database"
dependencies: [7, 8]
status: "pending"
created: "2026-01-15"
skills:
  - go
  - database
---
# Database Interface and SQLite Driver

## Objective
Define the database interface abstraction and implement the SQLite driver. This establishes the foundation for multi-database support with SQLite as the default backend.

## Skills Required
- go: Go programming with database/sql
- database: SQLite database operations

## Acceptance Criteria
- [ ] Database interface is defined with core operations
- [ ] SQLite driver implements the interface
- [ ] Connection pooling is configurable
- [ ] Health check (Ping) works correctly
- [ ] Transaction support is implemented
- [ ] In-memory SQLite works for testing

Use your internal Todo tool to track these and keep on track.

## Technical Requirements
- SQLite driver: `github.com/mattn/go-sqlite3` (CGO required)
- Standard `database/sql` interface
- Context support for all operations
- Configurable connection pool settings

## Input Dependencies
- Task 8: Configuration with database settings
- Task 9: Logging for database operations

## Output Artifacts
- `internal/database/database.go` - Interface definitions
- `internal/database/sqlite/sqlite.go` - SQLite implementation
- `internal/database/sqlite/sqlite_test.go` - Unit tests

## Implementation Notes

<details>
<summary>Detailed Implementation Guide</summary>

### Database Interface

Create `internal/database/database.go`:

```go
package database

import (
    "context"
    "database/sql"
)

// Database defines the interface for database operations
type Database interface {
    // DB returns the underlying *sql.DB for direct queries
    DB() *sql.DB

    // Migrate runs database migrations
    Migrate() error

    // Ping checks database connectivity
    Ping(ctx context.Context) error

    // Close closes the database connection
    Close() error

    // Begin starts a new transaction
    Begin(ctx context.Context) (Tx, error)

    // Driver returns the driver name (sqlite, mariadb, postgres)
    Driver() string
}

// Tx represents a database transaction
type Tx interface {
    // Commit commits the transaction
    Commit() error

    // Rollback aborts the transaction
    Rollback() error

    // Tx returns the underlying *sql.Tx
    Tx() *sql.Tx
}

// Config holds database configuration
type Config struct {
    Driver          string
    URL             string
    MaxOpenConns    int
    MaxIdleConns    int
    ConnMaxLifetime int // seconds
}
```

### SQLite Implementation

Create `internal/database/sqlite/sqlite.go`:

```go
package sqlite

import (
    "context"
    "database/sql"
    "fmt"

    _ "github.com/mattn/go-sqlite3"
    "github.com/rs/zerolog"

    "github.com/deviantintegral/replica/internal/database"
)

type SQLiteDB struct {
    db     *sql.DB
    logger zerolog.Logger
}

type sqliteTx struct {
    tx *sql.Tx
}

// New creates a new SQLite database connection
func New(cfg database.Config, logger zerolog.Logger) (*SQLiteDB, error) {
    // For in-memory: use ":memory:" or "file::memory:?cache=shared"
    dsn := cfg.URL
    if dsn == "" {
        dsn = "replica.db"
    }

    // Add SQLite options for better concurrency
    if dsn != ":memory:" && !containsOptions(dsn) {
        dsn += "?_journal_mode=WAL&_busy_timeout=5000&_foreign_keys=ON"
    }

    db, err := sql.Open("sqlite3", dsn)
    if err != nil {
        return nil, fmt.Errorf("opening sqlite database: %w", err)
    }

    // Configure connection pool
    if cfg.MaxOpenConns > 0 {
        db.SetMaxOpenConns(cfg.MaxOpenConns)
    } else {
        // SQLite works best with a single connection for writes
        db.SetMaxOpenConns(1)
    }
    if cfg.MaxIdleConns > 0 {
        db.SetMaxIdleConns(cfg.MaxIdleConns)
    }

    // Verify connection
    if err := db.PingContext(context.Background()); err != nil {
        db.Close()
        return nil, fmt.Errorf("connecting to sqlite database: %w", err)
    }

    logger.Info().Str("url", cfg.URL).Msg("connected to sqlite database")

    return &SQLiteDB{db: db, logger: logger}, nil
}

func containsOptions(dsn string) bool {
    return len(dsn) > 0 && (dsn[0] == ':' || containsRune(dsn, '?'))
}

func containsRune(s string, r rune) bool {
    for _, c := range s {
        if c == r {
            return true
        }
    }
    return false
}

func (s *SQLiteDB) DB() *sql.DB {
    return s.db
}

func (s *SQLiteDB) Migrate() error {
    // Migration implementation will be added in Task 12
    return nil
}

func (s *SQLiteDB) Ping(ctx context.Context) error {
    return s.db.PingContext(ctx)
}

func (s *SQLiteDB) Close() error {
    s.logger.Info().Msg("closing sqlite database connection")
    return s.db.Close()
}

func (s *SQLiteDB) Begin(ctx context.Context) (database.Tx, error) {
    tx, err := s.db.BeginTx(ctx, nil)
    if err != nil {
        return nil, fmt.Errorf("beginning transaction: %w", err)
    }
    return &sqliteTx{tx: tx}, nil
}

func (s *SQLiteDB) Driver() string {
    return "sqlite"
}

func (t *sqliteTx) Commit() error {
    return t.tx.Commit()
}

func (t *sqliteTx) Rollback() error {
    return t.tx.Rollback()
}

func (t *sqliteTx) Tx() *sql.Tx {
    return t.tx
}
```

### Unit Tests

Create `internal/database/sqlite/sqlite_test.go` with tests for:
- In-memory database creation
- File-based database creation
- Ping/health check
- Transaction begin/commit/rollback
- Connection closure

Use in-memory SQLite (`":memory:"`) for fast tests.

### Verification

```bash
go test ./internal/database/sqlite/...
```

</details>
