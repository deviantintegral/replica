// Package database provides database interface abstractions and common types
// for the replica application.
package database

import (
	"errors"
	"fmt"

	"github.com/deviantintegral/replica/internal/database/migrations"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

// RunMigrations runs database migrations for the given database connection.
// It detects the driver type and creates the appropriate migrate driver.
// Returns nil if migrations complete successfully or if there are no changes.
func RunMigrations(db Database) error {
	// Create the iofs source from embedded migrations
	source, err := iofs.New(migrations.FS, ".")
	if err != nil {
		return fmt.Errorf("failed to create migration source: %w", err)
	}

	// Create the appropriate database driver based on the driver type
	var m *migrate.Migrate
	switch db.Driver() {
	case "sqlite":
		driver, err := sqlite3.WithInstance(db.DB(), &sqlite3.Config{})
		if err != nil {
			return fmt.Errorf("failed to create sqlite migration driver: %w", err)
		}
		m, err = migrate.NewWithInstance("iofs", source, "sqlite3", driver)
		if err != nil {
			return fmt.Errorf("failed to create migrate instance for sqlite: %w", err)
		}

	case "mariadb":
		driver, err := mysql.WithInstance(db.DB(), &mysql.Config{})
		if err != nil {
			return fmt.Errorf("failed to create mysql migration driver: %w", err)
		}
		m, err = migrate.NewWithInstance("iofs", source, "mysql", driver)
		if err != nil {
			return fmt.Errorf("failed to create migrate instance for mysql: %w", err)
		}

	case "postgres":
		driver, err := postgres.WithInstance(db.DB(), &postgres.Config{})
		if err != nil {
			return fmt.Errorf("failed to create postgres migration driver: %w", err)
		}
		m, err = migrate.NewWithInstance("iofs", source, "postgres", driver)
		if err != nil {
			return fmt.Errorf("failed to create migrate instance for postgres: %w", err)
		}

	default:
		return fmt.Errorf("unsupported database driver: %s", db.Driver())
	}

	// Run migrations
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}
