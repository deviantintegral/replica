---
id: 15
group: "testing"
dependencies: [5, 10, 11, 14]
status: "pending"
created: "2026-01-14"
skills:
  - github-actions
  - ci-cd
---
# CI Database Matrix Testing

## Objective
Extend the GitHub Actions CI workflow to run integration tests against all three database backends using a job matrix.

## Skills Required
- **github-actions**: Matrix builds, service containers
- **ci-cd**: Multi-database testing strategies

## Acceptance Criteria
- [ ] CI matrix tests SQLite, MariaDB 11.8, PostgreSQL 18
- [ ] Database services start as GitHub Actions service containers
- [ ] Integration tests tagged and run separately from unit tests
- [ ] All database tests pass before merge allowed

Use your internal Todo tool to track these and keep on track.

## Technical Requirements
- GitHub Actions service containers for MariaDB and PostgreSQL
- Build matrix for database selection
- Integration tests use build tags or environment variables

## Input Dependencies
- Task 5: GitHub Actions CI workflow
- Task 10, 11: MariaDB and PostgreSQL drivers
- Task 14: Test utilities

## Output Artifacts
- Updated `.github/workflows/ci.yml` with database matrix
- Integration test files with proper tagging

## Implementation Notes

<details>
<summary>Detailed Implementation Instructions</summary>

1. **Update CI workflow** (`.github/workflows/ci.yml`):
   Add integration test job with matrix:
   ```yaml
   integration-test:
     name: Integration Test (${{ matrix.database }})
     runs-on: ubuntu-latest
     needs: [lint]
     strategy:
       fail-fast: false
       matrix:
         database: [sqlite, mariadb, postgres]
         include:
           - database: sqlite
             test_dsn: ":memory:"
           - database: mariadb
             test_dsn: "replica:replica_pass@tcp(localhost:3306)/replica?parseTime=true"
           - database: postgres
             test_dsn: "postgres://replica:replica_pass@localhost:5432/replica?sslmode=disable"

     services:
       mariadb:
         image: mariadb:11.8
         env:
           MYSQL_ROOT_PASSWORD: replica_root
           MYSQL_DATABASE: replica
           MYSQL_USER: replica
           MYSQL_PASSWORD: replica_pass
         ports:
           - 3306:3306
         options: >-
           --health-cmd="healthcheck.sh --connect --innodb_initialized"
           --health-interval=10s
           --health-timeout=5s
           --health-retries=5
         # Only start for mariadb matrix entry
         if: ${{ matrix.database == 'mariadb' }}

       postgres:
         image: postgres:18
         env:
           POSTGRES_DB: replica
           POSTGRES_USER: replica
           POSTGRES_PASSWORD: replica_pass
         ports:
           - 5432:5432
         options: >-
           --health-cmd="pg_isready -U replica -d replica"
           --health-interval=10s
           --health-timeout=5s
           --health-retries=5
         # Only start for postgres matrix entry
         if: ${{ matrix.database == 'postgres' }}

     steps:
       - uses: actions/checkout@v4

       - uses: actions/setup-go@v5
         with:
           go-version: ${{ env.GO_VERSION }}

       - name: Run integration tests
         env:
           TEST_DATABASE_DRIVER: ${{ matrix.database }}
           TEST_DATABASE_DSN: ${{ matrix.test_dsn }}
         run: go test -race -tags=integration ./...
   ```

2. **Create integration test file** (`internal/database/integration_test.go`):
   ```go
   //go:build integration

   package database_test

   import (
       "context"
       "os"
       "testing"
       "time"

       "github.com/deviantintegral/replica/internal/database"
       "github.com/deviantintegral/replica/internal/database/mysql"
       "github.com/deviantintegral/replica/internal/database/postgres"
       "github.com/deviantintegral/replica/internal/database/sqlite"
   )

   func TestDatabaseIntegration(t *testing.T) {
       driver := os.Getenv("TEST_DATABASE_DRIVER")
       dsn := os.Getenv("TEST_DATABASE_DSN")

       if driver == "" || dsn == "" {
           t.Skip("TEST_DATABASE_DRIVER and TEST_DATABASE_DSN required")
       }

       cfg := &database.Config{
           Driver: driver,
           URL:    dsn,
       }

       var db database.Database
       var err error

       switch driver {
       case "sqlite":
           db, err = sqlite.New(cfg)
       case "mariadb":
           db, err = mysql.New(cfg)
       case "postgres":
           db, err = postgres.New(cfg)
       default:
           t.Fatalf("unknown driver: %s", driver)
       }

       if err != nil {
           t.Fatalf("failed to connect: %v", err)
       }
       defer db.Close()

       t.Run("Ping", func(t *testing.T) {
           ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
           defer cancel()

           if err := db.Ping(ctx); err != nil {
               t.Errorf("ping failed: %v", err)
           }
       })

       t.Run("Migrate", func(t *testing.T) {
           if err := db.Migrate(); err != nil {
               t.Errorf("migrate failed: %v", err)
           }
       })

       t.Run("Transaction", func(t *testing.T) {
           ctx := context.Background()
           tx, err := db.Begin(ctx)
           if err != nil {
               t.Fatalf("begin failed: %v", err)
           }
           if err := tx.Rollback(); err != nil {
               t.Errorf("rollback failed: %v", err)
           }
       })
   }
   ```

3. **Update Makefile** with integration test target:
   ```makefile
   ## test-integration: Run integration tests (requires database)
   test-integration:
   	go test -race -tags=integration ./...
   ```

</details>
