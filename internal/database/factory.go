// Package database provides database interface abstractions and common types
// for the replica application. It defines the contract that all database
// drivers must implement.
package database

import (
	"fmt"
	"strings"

	"github.com/deviantintegral/replica/internal/database/mysql"
	"github.com/deviantintegral/replica/internal/database/postgres"
	"github.com/deviantintegral/replica/internal/database/sqlite"
	"github.com/rs/zerolog"
)

// Options contains optional settings for database initialization.
type Options struct {
	// AutoMigrate runs database migrations automatically after connection.
	AutoMigrate bool
}

// New creates a new database connection based on the provided configuration.
// It supports the following drivers:
//   - "sqlite": SQLite database
//   - "mysql" or "mariadb": MariaDB/MySQL database
//   - "postgres" or "postgresql": PostgreSQL database
//
// If opts.AutoMigrate is true, database migrations will be run after connection.
// On migration failure, the database connection is closed and an error is returned.
func New(cfg Config, logger zerolog.Logger, opts Options) (Database, error) {
	driver := strings.ToLower(cfg.Driver)

	var db Database
	var err error

	switch driver {
	case "sqlite":
		db, err = sqlite.New(cfg, logger)
	case "mysql", "mariadb":
		db, err = mysql.New(cfg, logger)
	case "postgres", "postgresql":
		db, err = postgres.New(cfg, logger)
	default:
		return nil, fmt.Errorf("unsupported database driver: %q", cfg.Driver)
	}

	if err != nil {
		return nil, err
	}

	if opts.AutoMigrate {
		if err := db.Migrate(); err != nil {
			db.Close()
			return nil, fmt.Errorf("failed to run migrations: %w", err)
		}
		logger.Info().Str("driver", driver).Msg("database migrations completed")
	}

	return db, nil
}
