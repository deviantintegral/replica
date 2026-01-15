---
id: 10
group: "database"
dependencies: [9]
status: "pending"
created: "2026-01-14"
skills:
  - go
  - database
---
# MariaDB/MySQL Driver

## Objective
Implement the MariaDB/MySQL driver using go-sql-driver/mysql with connection pooling and health checks.

## Skills Required
- **go**: database/sql package, MySQL driver
- **database**: MariaDB/MySQL connection strings, pooling

## Acceptance Criteria
- [ ] MariaDB driver implements Database interface
- [ ] Connection string parsing works correctly
- [ ] Connection pooling configurable
- [ ] Ping() health check works
- [ ] Proper handling of MySQL-specific settings (parseTime, etc.)

Use your internal Todo tool to track these and keep on track.

## Technical Requirements
- Use `github.com/go-sql-driver/mysql`
- DSN format: `user:password@tcp(host:port)/database?parseTime=true`
- Support MariaDB 11.8 (MySQL-compatible)

## Input Dependencies
- Task 9: Database interface definition

## Output Artifacts
- `internal/database/mysql/mysql.go` - MariaDB/MySQL implementation
- `internal/database/mysql/mysql_test.go` - Tests (require Docker or skip)

## Implementation Notes

<details>
<summary>Detailed Implementation Instructions</summary>

1. **Add MySQL driver dependency**:
   ```bash
   go get github.com/go-sql-driver/mysql
   ```

2. **Implement MariaDB driver** (`internal/database/mysql/mysql.go`):
   ```go
   package mysql

   import (
       "context"
       "database/sql"
       "fmt"

       _ "github.com/go-sql-driver/mysql"
       "github.com/deviantintegral/replica/internal/database"
   )

   type MySQL struct {
       db *sql.DB
   }

   // New creates a new MariaDB/MySQL database connection
   func New(cfg *database.Config) (*MySQL, error) {
       // DSN format: user:password@tcp(host:port)/database?parseTime=true
       db, err := sql.Open("mysql", cfg.URL)
       if err != nil {
           return nil, fmt.Errorf("opening mysql database: %w", err)
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
           return nil, fmt.Errorf("connecting to mysql: %w", err)
       }

       return &MySQL{db: db}, nil
   }

   func (m *MySQL) DB() *sql.DB {
       return m.db
   }

   func (m *MySQL) Migrate() error {
       // Migration will be implemented in task 12
       return nil
   }

   func (m *MySQL) Ping(ctx context.Context) error {
       return m.db.PingContext(ctx)
   }

   func (m *MySQL) Close() error {
       return m.db.Close()
   }

   func (m *MySQL) Begin(ctx context.Context) (*sql.Tx, error) {
       return m.db.BeginTx(ctx, nil)
   }
   ```

3. **Write tests** (`internal/database/mysql/mysql_test.go`):
   ```go
   package mysql_test

   import (
       "context"
       "os"
       "testing"
       "time"

       "github.com/deviantintegral/replica/internal/database"
       "github.com/deviantintegral/replica/internal/database/mysql"
   )

   func TestMySQL(t *testing.T) {
       // Skip if no MySQL available
       dsn := os.Getenv("TEST_MYSQL_DSN")
       if dsn == "" {
           t.Skip("TEST_MYSQL_DSN not set, skipping MySQL tests")
       }

       cfg := &database.Config{
           Driver:       "mariadb",
           URL:          dsn,
           MaxOpenConns: 5,
           MaxIdleConns: 2,
       }

       db, err := mysql.New(cfg)
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
   docker-compose up -d mariadb
   TEST_MYSQL_DSN="replica:replica_pass@tcp(localhost:3306)/replica?parseTime=true" go test ./internal/database/mysql/...
   ```

</details>
