---
id: 12
group: "database"
dependencies: [9, 10, 11]
status: "pending"
created: "2026-01-14"
skills:
  - go
  - database
---
# golang-migrate Integration

## Objective
Integrate golang-migrate for database schema migrations across all three database backends with embedded migration files.

## Skills Required
- **go**: Go embed, migrations
- **database**: SQL DDL, cross-database compatibility

## Acceptance Criteria
- [ ] golang-migrate integrated with all three drivers
- [ ] Migration files embedded in binary using go:embed
- [ ] Migrate() method works for SQLite, MariaDB, PostgreSQL
- [ ] Initial migration creates schema_migrations tracking table
- [ ] Migrations are database-agnostic where possible

Use your internal Todo tool to track these and keep on track.

## Technical Requirements
- Use `github.com/golang-migrate/migrate/v4`
- Embed migrations with `//go:embed`
- Migration file naming: `NNNNNN_description.up.sql`, `NNNNNN_description.down.sql`
- Support database-specific migrations when needed

## Input Dependencies
- Task 9, 10, 11: All database drivers implemented

## Output Artifacts
- `internal/database/migrations/` - Migration SQL files
- Updated `Migrate()` method in all drivers
- `internal/database/migrate.go` - Shared migration helper

## Implementation Notes

<details>
<summary>Detailed Implementation Instructions</summary>

1. **Add golang-migrate dependency**:
   ```bash
   go get github.com/golang-migrate/migrate/v4
   go get github.com/golang-migrate/migrate/v4/database/sqlite3
   go get github.com/golang-migrate/migrate/v4/database/mysql
   go get github.com/golang-migrate/migrate/v4/database/postgres
   go get github.com/golang-migrate/migrate/v4/source/iofs
   ```

2. **Create migration embedding** (`internal/database/migrations/migrations.go`):
   ```go
   package migrations

   import "embed"

   //go:embed *.sql
   var FS embed.FS
   ```

3. **Create initial migration** (`internal/database/migrations/000001_init.up.sql`):
   ```sql
   -- Initial schema migration
   -- This file intentionally empty for now
   -- Future releases will add content tables
   ```

4. **Create down migration** (`internal/database/migrations/000001_init.down.sql`):
   ```sql
   -- Rollback initial migration
   -- This file intentionally empty for now
   ```

5. **Create migration helper** (`internal/database/migrate.go`):
   ```go
   package database

   import (
       "database/sql"
       "fmt"

       "github.com/golang-migrate/migrate/v4"
       "github.com/golang-migrate/migrate/v4/database"
       "github.com/golang-migrate/migrate/v4/database/mysql"
       "github.com/golang-migrate/migrate/v4/database/postgres"
       "github.com/golang-migrate/migrate/v4/database/sqlite3"
       "github.com/golang-migrate/migrate/v4/source/iofs"
       "github.com/deviantintegral/replica/internal/database/migrations"
   )

   // RunMigrations executes all pending migrations
   func RunMigrations(db *sql.DB, driver string) error {
       // Create source from embedded files
       source, err := iofs.New(migrations.FS, ".")
       if err != nil {
           return fmt.Errorf("creating migration source: %w", err)
       }

       // Create database driver
       var dbDriver database.Driver
       switch driver {
       case "sqlite":
           dbDriver, err = sqlite3.WithInstance(db, &sqlite3.Config{})
       case "mariadb":
           dbDriver, err = mysql.WithInstance(db, &mysql.Config{})
       case "postgres":
           dbDriver, err = postgres.WithInstance(db, &postgres.Config{})
       default:
           return fmt.Errorf("unknown database driver: %s", driver)
       }
       if err != nil {
           return fmt.Errorf("creating database driver: %w", err)
       }

       // Create migrator
       m, err := migrate.NewWithInstance("iofs", source, driver, dbDriver)
       if err != nil {
           return fmt.Errorf("creating migrator: %w", err)
       }

       // Run migrations
       if err := m.Up(); err != nil && err != migrate.ErrNoChange {
           return fmt.Errorf("running migrations: %w", err)
       }

       return nil
   }
   ```

6. **Update SQLite Migrate()** (`internal/database/sqlite/sqlite.go`):
   ```go
   func (s *SQLite) Migrate() error {
       return database.RunMigrations(s.db, "sqlite")
   }
   ```

7. **Update MySQL Migrate()** (`internal/database/mysql/mysql.go`):
   ```go
   func (m *MySQL) Migrate() error {
       return database.RunMigrations(m.db, "mariadb")
   }
   ```

8. **Update PostgreSQL Migrate()** (`internal/database/postgres/postgres.go`):
   ```go
   func (p *Postgres) Migrate() error {
       return database.RunMigrations(p.db, "postgres")
   }
   ```

9. **Test migrations**:
   ```go
   func TestMigrations(t *testing.T) {
       db, _ := sqlite.New(&database.Config{URL: ":memory:"})
       defer db.Close()

       if err := db.Migrate(); err != nil {
           t.Fatalf("migration failed: %v", err)
       }
   }
   ```

</details>
