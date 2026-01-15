---
id: 11
group: "database"
dependencies: [9]
status: "pending"
created: "2026-01-14"
skills:
  - go
  - database
---
# PostgreSQL Driver

## Objective
Implement the PostgreSQL driver using pgx/v5/stdlib with connection pooling and health checks.

## Skills Required
- **go**: database/sql package, pgx driver
- **database**: PostgreSQL connection strings, pooling

## Acceptance Criteria
- [ ] PostgreSQL driver implements Database interface
- [ ] Connection string parsing works correctly
- [ ] Connection pooling configurable
- [ ] Ping() health check works
- [ ] Uses pgx/v5/stdlib for database/sql compatibility

Use your internal Todo tool to track these and keep on track.

## Technical Requirements
- Use `github.com/jackc/pgx/v5/stdlib`
- DSN format: `postgres://user:password@host:port/database?sslmode=disable`
- Support PostgreSQL 18

## Input Dependencies
- Task 9: Database interface definition

## Output Artifacts
- `internal/database/postgres/postgres.go` - PostgreSQL implementation
- `internal/database/postgres/postgres_test.go` - Tests (require Docker or skip)

## Implementation Notes

<details>
<summary>Detailed Implementation Instructions</summary>

1. **Add pgx dependency**:
   ```bash
   go get github.com/jackc/pgx/v5
   ```

2. **Implement PostgreSQL driver** (`internal/database/postgres/postgres.go`):
   ```go
   package postgres

   import (
       "context"
       "database/sql"
       "fmt"

       _ "github.com/jackc/pgx/v5/stdlib"
       "github.com/deviantintegral/replica/internal/database"
   )

   type Postgres struct {
       db *sql.DB
   }

   // New creates a new PostgreSQL database connection
   func New(cfg *database.Config) (*Postgres, error) {
       // DSN format: postgres://user:password@host:port/database?sslmode=disable
       db, err := sql.Open("pgx", cfg.URL)
       if err != nil {
           return nil, fmt.Errorf("opening postgres database: %w", err)
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

       // Verify connection
       if err := db.Ping(); err != nil {
           db.Close()
           return nil, fmt.Errorf("connecting to postgres: %w", err)
       }

       return &Postgres{db: db}, nil
   }

   func (p *Postgres) DB() *sql.DB {
       return p.db
   }

   func (p *Postgres) Migrate() error {
       // Migration will be implemented in task 12
       return nil
   }

   func (p *Postgres) Ping(ctx context.Context) error {
       return p.db.PingContext(ctx)
   }

   func (p *Postgres) Close() error {
       return p.db.Close()
   }

   func (p *Postgres) Begin(ctx context.Context) (*sql.Tx, error) {
       return p.db.BeginTx(ctx, nil)
   }
   ```

3. **Write tests** (`internal/database/postgres/postgres_test.go`):
   ```go
   package postgres_test

   import (
       "context"
       "os"
       "testing"
       "time"

       "github.com/deviantintegral/replica/internal/database"
       "github.com/deviantintegral/replica/internal/database/postgres"
   )

   func TestPostgres(t *testing.T) {
       // Skip if no PostgreSQL available
       dsn := os.Getenv("TEST_POSTGRES_DSN")
       if dsn == "" {
           t.Skip("TEST_POSTGRES_DSN not set, skipping PostgreSQL tests")
       }

       cfg := &database.Config{
           Driver:       "postgres",
           URL:          dsn,
           MaxOpenConns: 5,
           MaxIdleConns: 2,
       }

       db, err := postgres.New(cfg)
       if err != nil {
           t.Fatalf("failed to connect: %v", err)
       }
       defer db.Close()

       ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
       defer cancel()

       if err := db.Ping(ctx); err != nil {
           t.Errorf("ping failed: %v", err)
       }
   }
   ```

4. **Test with docker-compose**:
   ```bash
   docker-compose up -d postgres
   TEST_POSTGRES_DSN="postgres://replica:replica_pass@localhost:5432/replica?sslmode=disable" go test ./internal/database/postgres/...
   ```

</details>
