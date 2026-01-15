---
id: 14
group: "testing"
dependencies: [9]
status: "pending"
created: "2026-01-14"
skills:
  - go
  - unit-testing
---
# Test Utilities

## Objective
Create test helper utilities for database setup/teardown, providing easy-to-use functions for unit and integration tests.

## Skills Required
- **go**: Go testing patterns, test helpers
- **unit-testing**: Test fixtures, setup/teardown patterns

## Acceptance Criteria
- [ ] In-memory SQLite helper for fast unit tests
- [ ] Test database cleanup functions
- [ ] Helper functions for common test patterns
- [ ] Tests run in isolation (no shared state)

Use your internal Todo tool to track these and keep on track.

## Technical Requirements
- Use Go's `testing` package (no testify)
- In-memory SQLite for unit tests
- Each test gets fresh database

## Input Dependencies
- Task 9: SQLite driver (for in-memory testing)

## Output Artifacts
- `internal/testutil/database.go` - Database test helpers
- `internal/testutil/database_test.go` - Tests for helpers

## Implementation Notes

<details>
<summary>Detailed Implementation Instructions</summary>

1. **Create test helpers** (`internal/testutil/database.go`):
   ```go
   package testutil

   import (
       "testing"

       "github.com/deviantintegral/replica/internal/database"
       "github.com/deviantintegral/replica/internal/database/sqlite"
   )

   // TestDB creates an in-memory SQLite database for testing.
   // The database is automatically closed when the test completes.
   func TestDB(t *testing.T) database.Database {
       t.Helper()

       cfg := &database.Config{
           Driver: "sqlite",
           URL:    ":memory:",
       }

       db, err := sqlite.New(cfg)
       if err != nil {
           t.Fatalf("failed to create test database: %v", err)
       }

       // Run migrations
       if err := db.Migrate(); err != nil {
           db.Close()
           t.Fatalf("failed to run migrations: %v", err)
       }

       // Register cleanup
       t.Cleanup(func() {
           db.Close()
       })

       return db
   }

   // MustExec executes a SQL statement and fails the test if it errors.
   func MustExec(t *testing.T, db database.Database, query string, args ...interface{}) {
       t.Helper()

       _, err := db.DB().Exec(query, args...)
       if err != nil {
           t.Fatalf("failed to execute query %q: %v", query, err)
       }
   }

   // MustQuery executes a query and fails the test if it errors.
   // The caller is responsible for closing the rows.
   func MustQuery(t *testing.T, db database.Database, query string, args ...interface{}) *sql.Rows {
       t.Helper()

       rows, err := db.DB().Query(query, args...)
       if err != nil {
           t.Fatalf("failed to query %q: %v", query, err)
       }
       return rows
   }
   ```

2. **Write tests for helpers** (`internal/testutil/database_test.go`):
   ```go
   package testutil_test

   import (
       "context"
       "testing"
       "time"

       "github.com/deviantintegral/replica/internal/testutil"
   )

   func TestTestDB(t *testing.T) {
       db := testutil.TestDB(t)

       ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
       defer cancel()

       // Verify database is working
       if err := db.Ping(ctx); err != nil {
           t.Errorf("ping failed: %v", err)
       }

       // Verify SQL works
       _, err := db.DB().Exec("CREATE TABLE test (id INTEGER PRIMARY KEY)")
       if err != nil {
           t.Errorf("create table failed: %v", err)
       }
   }

   func TestTestDB_Isolation(t *testing.T) {
       // Create two test databases
       db1 := testutil.TestDB(t)
       db2 := testutil.TestDB(t)

       // Create table in db1
       _, err := db1.DB().Exec("CREATE TABLE test (id INTEGER)")
       if err != nil {
           t.Fatalf("create table in db1 failed: %v", err)
       }

       // Table should not exist in db2
       _, err = db2.DB().Exec("INSERT INTO test VALUES (1)")
       if err == nil {
           t.Error("expected error inserting into non-existent table in db2")
       }
   }
   ```

3. **Usage example in other tests**:
   ```go
   func TestSomeFeature(t *testing.T) {
       db := testutil.TestDB(t)

       // Test code using db
       // No cleanup needed - handled automatically
   }
   ```

</details>
