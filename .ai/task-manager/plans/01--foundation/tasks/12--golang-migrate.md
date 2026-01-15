---
id: 12
group: "database"
dependencies: [9, 10, 11]
status: "pending"
created: "2026-01-15"
skills:
  - go
  - database
---
# golang-migrate Integration

## Objective
Integrate golang-migrate for database schema migrations across all database backends. This provides version-controlled schema management with support for up/down migrations.

## Skills Required
- go: Go programming with golang-migrate
- database: SQL DDL and migration patterns

## Acceptance Criteria
- [ ] golang-migrate is integrated with all three database drivers
- [ ] Migrations are embedded in the binary using Go embed
- [ ] `Migrate()` method applies all pending migrations
- [ ] Initial migration creates a simple schema for testing
- [ ] Migrations work identically across SQLite, MariaDB, and PostgreSQL

Use your internal Todo tool to track these and keep on track.

## Technical Requirements
- golang-migrate: `github.com/golang-migrate/migrate/v4`
- Go embed for migration files
- Database-agnostic SQL where possible
- Separate migration files per database when needed

## Input Dependencies
- Task 10: SQLite driver
- Task 11: MariaDB driver
- Task 12: PostgreSQL driver

## Output Artifacts
- `internal/database/migrations/` - Migration SQL files
- `internal/database/migrate.go` - Migration helper functions
- Updated driver implementations with Migrate() support

## Implementation Notes

<details>
<summary>Detailed Implementation Guide</summary>

### Migration Files Structure

Create migration files in `internal/database/migrations/`:

```
migrations/
├── 000001_create_schema_version.up.sql
├── 000001_create_schema_version.down.sql
```

### Initial Migration

Create `internal/database/migrations/000001_create_schema_version.up.sql`:

```sql
-- Schema version tracking table (used by golang-migrate internally)
-- This migration serves as a placeholder to verify migrations work
-- Actual schema tables will be added in future releases

CREATE TABLE IF NOT EXISTS replica_metadata (
    key VARCHAR(255) PRIMARY KEY,
    value TEXT NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO replica_metadata (key, value) VALUES ('schema_version', '1');
```

Create `internal/database/migrations/000001_create_schema_version.down.sql`:

```sql
DROP TABLE IF EXISTS replica_metadata;
```

### Migration Embedding

Create `internal/database/migrations/embed.go`:

```go
package migrations

import "embed"

//go:embed *.sql
var FS embed.FS
```

### Migration Helper

Create `internal/database/migrate.go`:

```go
package database

import (
    "fmt"

    "github.com/golang-migrate/migrate/v4"
    "github.com/golang-migrate/migrate/v4/database"
    "github.com/golang-migrate/migrate/v4/database/mysql"
    "github.com/golang-migrate/migrate/v4/database/postgres"
    "github.com/golang-migrate/migrate/v4/database/sqlite3"
    "github.com/golang-migrate/migrate/v4/source/iofs"

    "github.com/deviantintegral/replica/internal/database/migrations"
)

// RunMigrations applies all pending migrations for the given database
func RunMigrations(db Database) error {
    sqlDB := db.DB()
    driverName := db.Driver()

    var driver database.Driver
    var err error

    switch driverName {
    case "sqlite":
        driver, err = sqlite3.WithInstance(sqlDB, &sqlite3.Config{})
    case "mariadb":
        driver, err = mysql.WithInstance(sqlDB, &mysql.Config{})
    case "postgres":
        driver, err = postgres.WithInstance(sqlDB, &postgres.Config{})
    default:
        return fmt.Errorf("unsupported database driver: %s", driverName)
    }

    if err != nil {
        return fmt.Errorf("creating migration driver: %w", err)
    }

    // Create source from embedded filesystem
    source, err := iofs.New(migrations.FS, ".")
    if err != nil {
        return fmt.Errorf("creating migration source: %w", err)
    }

    m, err := migrate.NewWithInstance("iofs", source, driverName, driver)
    if err != nil {
        return fmt.Errorf("creating migrator: %w", err)
    }

    if err := m.Up(); err != nil && err != migrate.ErrNoChange {
        return fmt.Errorf("running migrations: %w", err)
    }

    return nil
}
```

### Update Driver Implementations

Update each driver's `Migrate()` method to call `RunMigrations`:

```go
// In sqlite/sqlite.go, mysql/mysql.go, postgres/postgres.go
func (s *SQLiteDB) Migrate() error {
    return database.RunMigrations(s)
}
```

### Dependencies

Add required dependencies:
```bash
go get github.com/golang-migrate/migrate/v4
go get github.com/golang-migrate/migrate/v4/database/sqlite3
go get github.com/golang-migrate/migrate/v4/database/mysql
go get github.com/golang-migrate/migrate/v4/database/postgres
go get github.com/golang-migrate/migrate/v4/source/iofs
```

### Unit Tests

Test migration functionality:
- Migrations apply successfully on fresh database
- Migrations are idempotent (running twice doesn't error)
- Down migrations work correctly
- Version tracking is accurate

### Verification

```bash
go test ./internal/database/...
```

Test with SQLite in-memory:
```go
db, _ := sqlite.New(database.Config{URL: ":memory:"}, logger)
err := db.Migrate()
// Should complete without error
```

</details>
