---
id: 14
group: "testing"
dependencies: [9]
status: "pending"
created: "2026-01-15"
skills:
  - go
  - unit-testing
---
# Test Utilities

## Objective
Create test helper utilities for database setup/teardown and common test patterns. This provides consistent, reusable testing infrastructure for all future tests.

## Skills Required
- go: Go testing patterns and best practices
- unit-testing: Test helper design

## Acceptance Criteria
- [ ] In-memory SQLite helper for fast unit tests
- [ ] Database cleanup utilities for test isolation
- [ ] Test fixtures helper for consistent test data
- [ ] Test helpers integrate with Go's testing.T

Use your internal Todo tool to track these and keep on track.

## Technical Requirements
- Go standard library testing package (no testify)
- In-memory SQLite for fast tests
- Helper functions that accept *testing.T

## Input Dependencies
- Task 10: SQLite driver for in-memory testing

## Output Artifacts
- `internal/testutil/database.go` - Database test helpers
- `internal/testutil/database_test.go` - Tests for test helpers

## Implementation Notes

<details>
<summary>Detailed Implementation Guide</summary>

### Test Database Helper

Create `internal/testutil/database.go`:

```go
package testutil

import (
    "testing"

    "github.com/rs/zerolog"

    "github.com/deviantintegral/replica/internal/database"
    "github.com/deviantintegral/replica/internal/database/sqlite"
)

// TestDB creates an in-memory SQLite database for testing
// The database is automatically cleaned up when the test completes
func TestDB(t *testing.T) database.Database {
    t.Helper()

    // Create silent logger for tests
    logger := zerolog.Nop()

    cfg := database.Config{
        Driver: "sqlite",
        URL:    ":memory:",
    }

    db, err := sqlite.New(cfg, logger)
    if err != nil {
        t.Fatalf("failed to create test database: %v", err)
    }

    // Run migrations
    if err := db.Migrate(); err != nil {
        db.Close()
        t.Fatalf("failed to migrate test database: %v", err)
    }

    // Register cleanup
    t.Cleanup(func() {
        db.Close()
    })

    return db
}

// TestDBWithLogger creates a test database with a custom logger
func TestDBWithLogger(t *testing.T, logger zerolog.Logger) database.Database {
    t.Helper()

    cfg := database.Config{
        Driver: "sqlite",
        URL:    ":memory:",
    }

    db, err := sqlite.New(cfg, logger)
    if err != nil {
        t.Fatalf("failed to create test database: %v", err)
    }

    if err := db.Migrate(); err != nil {
        db.Close()
        t.Fatalf("failed to migrate test database: %v", err)
    }

    t.Cleanup(func() {
        db.Close()
    })

    return db
}

// MustExec executes a SQL statement and fails the test on error
func MustExec(t *testing.T, db database.Database, query string, args ...interface{}) {
    t.Helper()

    _, err := db.DB().Exec(query, args...)
    if err != nil {
        t.Fatalf("failed to execute query %q: %v", query, err)
    }
}

// MustQuery executes a query and returns rows, failing the test on error
func MustQuery(t *testing.T, db database.Database, query string, args ...interface{}) *sql.Rows {
    t.Helper()

    rows, err := db.DB().Query(query, args...)
    if err != nil {
        t.Fatalf("failed to execute query %q: %v", query, err)
    }
    return rows
}
```

### Assertion Helpers

Create `internal/testutil/assert.go`:

```go
package testutil

import (
    "testing"
)

// AssertEqual fails the test if got != want
func AssertEqual[T comparable](t *testing.T, got, want T) {
    t.Helper()
    if got != want {
        t.Errorf("got %v, want %v", got, want)
    }
}

// AssertNoError fails the test if err is not nil
func AssertNoError(t *testing.T, err error) {
    t.Helper()
    if err != nil {
        t.Errorf("unexpected error: %v", err)
    }
}

// AssertError fails the test if err is nil
func AssertError(t *testing.T, err error) {
    t.Helper()
    if err == nil {
        t.Error("expected error, got nil")
    }
}

// AssertContains fails the test if s does not contain substr
func AssertContains(t *testing.T, s, substr string) {
    t.Helper()
    if !strings.Contains(s, substr) {
        t.Errorf("expected %q to contain %q", s, substr)
    }
}
```

### Usage Example

```go
func TestSomeDatabaseOperation(t *testing.T) {
    db := testutil.TestDB(t)

    // Test code here - db is automatically cleaned up
    testutil.MustExec(t, db, "INSERT INTO replica_metadata (key, value) VALUES (?, ?)", "test", "value")

    rows := testutil.MustQuery(t, db, "SELECT value FROM replica_metadata WHERE key = ?", "test")
    defer rows.Close()

    // Assertions...
}
```

### Unit Tests

Create `internal/testutil/database_test.go` with tests for:
- TestDB creates valid database
- Database is cleaned up after test
- MustExec handles errors correctly
- Multiple TestDB calls create isolated databases

### Verification

```bash
go test ./internal/testutil/...
```

</details>
