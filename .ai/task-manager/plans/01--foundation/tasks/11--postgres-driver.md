---
id: 11
group: "database"
dependencies: [9]
status: "completed"
created: "2026-01-15"
skills:
  - go
  - database
---
# PostgreSQL Driver

## Objective
Implement the PostgreSQL driver for the database interface. This enables Replica to use PostgreSQL as a production database backend.

## Skills Required
- go: Go programming with database/sql
- database: PostgreSQL database operations

## Acceptance Criteria
- [ ] PostgreSQL driver implements the Database interface
- [ ] Connection string parsing works correctly
- [ ] Connection pooling is configurable
- [ ] Health check (Ping) works correctly
- [ ] Transaction support is implemented

Use your internal Todo tool to track these and keep on track.

## Technical Requirements
- PostgreSQL driver: `github.com/jackc/pgx/v5/stdlib` (database/sql compatible)
- Standard `database/sql` interface
- DSN format: `postgres://user:password@host:port/database?sslmode=disable`

## Input Dependencies
- Task 10: Database interface definition

## Output Artifacts
- `internal/database/postgres/postgres.go` - PostgreSQL implementation
- `internal/database/postgres/postgres_test.go` - Unit tests

## Implementation Notes

<details>
<summary>Detailed Implementation Guide</summary>

### PostgreSQL Implementation

Create `internal/database/postgres/postgres.go`:

```go
package postgres

import (
    "context"
    "database/sql"
    "fmt"
    "time"

    _ "github.com/jackc/pgx/v5/stdlib"
    "github.com/rs/zerolog"

    "github.com/deviantintegral/replica/internal/database"
)

type PostgresDB struct {
    db     *sql.DB
    logger zerolog.Logger
}

type postgresTx struct {
    tx *sql.Tx
}

// New creates a new PostgreSQL database connection
func New(cfg database.Config, logger zerolog.Logger) (*PostgresDB, error) {
    dsn := cfg.URL
    if dsn == "" {
        return nil, fmt.Errorf("database URL is required for postgres")
    }

    // pgx/stdlib uses "pgx" as the driver name for database/sql
    db, err := sql.Open("pgx", dsn)
    if err != nil {
        return nil, fmt.Errorf("opening postgres database: %w", err)
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
        return nil, fmt.Errorf("connecting to postgres database: %w", err)
    }

    logger.Info().Msg("connected to postgres database")

    return &PostgresDB{db: db, logger: logger}, nil
}

func (p *PostgresDB) DB() *sql.DB {
    return p.db
}

func (p *PostgresDB) Migrate() error {
    // Migration implementation will be added in Task 12
    return nil
}

func (p *PostgresDB) Ping(ctx context.Context) error {
    return p.db.PingContext(ctx)
}

func (p *PostgresDB) Close() error {
    p.logger.Info().Msg("closing postgres database connection")
    return p.db.Close()
}

func (p *PostgresDB) Begin(ctx context.Context) (database.Tx, error) {
    tx, err := p.db.BeginTx(ctx, nil)
    if err != nil {
        return nil, fmt.Errorf("beginning transaction: %w", err)
    }
    return &postgresTx{tx: tx}, nil
}

func (p *PostgresDB) Driver() string {
    return "postgres"
}

func (t *postgresTx) Commit() error {
    return t.tx.Commit()
}

func (t *postgresTx) Rollback() error {
    return t.tx.Rollback()
}

func (t *postgresTx) Tx() *sql.Tx {
    return t.tx
}
```

### Unit Tests

Create `internal/database/postgres/postgres_test.go`.

Note: Full integration tests require a running PostgreSQL instance. Unit tests should focus on:
- DSN validation
- Configuration handling
- Error handling for invalid connections

Integration tests against real PostgreSQL will be covered in Task 15 (CI Database Matrix).

### Test with Docker

For local testing:
```bash
docker compose up -d postgres
# Wait for health check
go test ./internal/database/postgres/... -tags=integration
```

### Verification

```bash
go test ./internal/database/postgres/...
```

</details>
