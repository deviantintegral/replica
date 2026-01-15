---
id: 9
group: "database"
dependencies: [7, 8]
status: "pending"
created: "2026-01-14"
skills:
  - go
  - database
---
# Database Interface and SQLite Driver

## Objective
Define the database abstraction interface and implement the SQLite driver using go-sqlite3 with connection pooling and health checks.

## Skills Required
- **go**: Interfaces, database/sql package
- **database**: SQLite, connection pooling, health checks

## Acceptance Criteria
- [ ] Database interface defined with DB(), Migrate(), Ping(), Close(), Begin()
- [ ] SQLite driver implements interface using go-sqlite3
- [ ] Connection pooling configurable
- [ ] Ping() health check works
- [ ] In-memory SQLite supported for testing

Use your internal Todo tool to track these and keep on track.

## Technical Requirements
- Use `github.com/mattn/go-sqlite3` (CGO required)
- Standard `database/sql` interface
- Support both file-based and in-memory (`:memory:`) SQLite

## Input Dependencies
- Task 7: Configuration system (database config)
- Task 8: Logging (for database operations)

## Output Artifacts
- `internal/database/database.go` - Interface definition
- `internal/database/sqlite/sqlite.go` - SQLite implementation
- `internal/database/sqlite/sqlite_test.go` - Tests

## Implementation Notes

<details>
<summary>Detailed Implementation Instructions</summary>

1. **Add go-sqlite3 dependency**:
   ```bash
   go get github.com/mattn/go-sqlite3
   ```

2. **Define database interface** (`internal/database/database.go`):
   ```go
   package database

   import (
       "context"
       "database/sql"
   )

   // Database is the interface for database operations
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
       Begin(ctx context.Context) (*sql.Tx, error)
   }

   // Config holds database configuration
   type Config struct {
       Driver          string
       URL             string
       MaxOpenConns    int
       MaxIdleConns    int
       ConnMaxLifetime time.Duration
   }
   ```

3. **Implement SQLite driver** (`internal/database/sqlite/sqlite.go`):
   ```go
   package sqlite

   import (
       "context"
       "database/sql"
       "fmt"

       _ "github.com/mattn/go-sqlite3"
       "github.com/deviantintegral/replica/internal/database"
   )

   type SQLite struct {
       db *sql.DB
   }

   // New creates a new SQLite database connection
   func New(cfg *database.Config) (*SQLite, error) {
       // SQLite URL is the file path, or ":memory:" for in-memory
       db, err := sql.Open("sqlite3", cfg.URL)
       if err != nil {
           return nil, fmt.Errorf("opening sqlite database: %w", err)
       }

       // Configure connection pool
       if cfg.MaxOpenConns > 0 {
           db.SetMaxOpenConns(cfg.MaxOpenConns)
       }
       if cfg.MaxIdleConns > 0 {
           db.SetMaxIdleConns(cfg.MaxIdleConns)
       }
       if cfg.ConnMaxLifetime > 0 {
           db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
       }

       // Enable WAL mode for better concurrency (file-based only)
       if cfg.URL != ":memory:" {
           _, err = db.Exec("PRAGMA journal_mode=WAL")
           if err != nil {
               db.Close()
               return nil, fmt.Errorf("enabling WAL mode: %w", err)
           }
       }

       // Enable foreign keys
       _, err = db.Exec("PRAGMA foreign_keys=ON")
       if err != nil {
           db.Close()
           return nil, fmt.Errorf("enabling foreign keys: %w", err)
       }

       return &SQLite{db: db}, nil
   }

   func (s *SQLite) DB() *sql.DB {
       return s.db
   }

   func (s *SQLite) Migrate() error {
       // Migration will be implemented in task 12
       return nil
   }

   func (s *SQLite) Ping(ctx context.Context) error {
       return s.db.PingContext(ctx)
   }

   func (s *SQLite) Close() error {
       return s.db.Close()
   }

   func (s *SQLite) Begin(ctx context.Context) (*sql.Tx, error) {
       return s.db.BeginTx(ctx, nil)
   }
   ```

4. **Write tests** (`internal/database/sqlite/sqlite_test.go`):
   Test cases:
   - Connect to in-memory database
   - Ping succeeds
   - Begin/Commit transaction
   - Begin/Rollback transaction
   - Close connection

</details>
