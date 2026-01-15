// Package sqlite provides a SQLite database driver implementation
// for the replica application.
package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/deviantintegral/replica/internal/database"
	_ "github.com/mattn/go-sqlite3" // SQLite driver
	"github.com/rs/zerolog"
)

const (
	// DefaultMaxOpenConns is the default maximum number of open connections.
	// SQLite performs best with a single writer connection.
	DefaultMaxOpenConns = 1

	// DefaultMaxIdleConns is the default maximum number of idle connections.
	DefaultMaxIdleConns = 1
)

// SQLiteDB implements the database.Database interface for SQLite.
type SQLiteDB struct {
	db     *sql.DB
	logger zerolog.Logger
}

// sqliteTx implements the database.Tx interface for SQLite transactions.
type sqliteTx struct {
	tx *sql.Tx
}

// New creates a new SQLite database connection with the given configuration.
// It configures WAL mode, busy timeout, and foreign keys by default.
// If MaxOpenConns is not set, it defaults to 1 (SQLite best practice).
func New(cfg database.Config, logger zerolog.Logger) (*SQLiteDB, error) {
	logger = logger.With().Str("component", "sqlite").Logger()

	// Build connection string with pragmas
	dsn := buildDSN(cfg.URL)
	logger.Debug().Str("dsn", dsn).Msg("opening SQLite database")

	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open SQLite database: %w", err)
	}

	// Configure connection pool
	maxOpen := cfg.MaxOpenConns
	if maxOpen <= 0 {
		maxOpen = DefaultMaxOpenConns
	}
	db.SetMaxOpenConns(maxOpen)

	maxIdle := cfg.MaxIdleConns
	if maxIdle <= 0 {
		maxIdle = DefaultMaxIdleConns
	}
	db.SetMaxIdleConns(maxIdle)

	if cfg.ConnMaxLifetime > 0 {
		db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	}

	// Verify connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping SQLite database: %w", err)
	}

	logger.Info().Str("url", cfg.URL).Msg("SQLite database connected")

	return &SQLiteDB{
		db:     db,
		logger: logger,
	}, nil
}

// buildDSN constructs the SQLite connection string with appropriate pragmas.
// It enables WAL mode, sets a busy timeout, and enables foreign key constraints.
func buildDSN(url string) string {
	// Check if URL already has query parameters
	separator := "?"
	if strings.Contains(url, "?") {
		separator = "&"
	}

	// Add pragmas for:
	// - WAL mode for better concurrent read performance
	// - Busy timeout of 5 seconds to wait for locks
	// - Foreign key enforcement
	pragmas := "_journal_mode=WAL&_busy_timeout=5000&_foreign_keys=on"

	return url + separator + pragmas
}

// DB returns the underlying *sql.DB connection pool.
func (s *SQLiteDB) DB() *sql.DB {
	return s.db
}

// Migrate runs database migrations using golang-migrate.
// It applies all pending migrations from the embedded SQL files.
func (s *SQLiteDB) Migrate() error {
	s.logger.Debug().Msg("running database migrations")
	if err := database.RunMigrations(s); err != nil {
		return err
	}
	s.logger.Info().Msg("database migrations completed successfully")
	return nil
}

// Ping verifies the database connection is alive.
func (s *SQLiteDB) Ping(ctx context.Context) error {
	return s.db.PingContext(ctx)
}

// Close closes the database connection and releases resources.
func (s *SQLiteDB) Close() error {
	s.logger.Debug().Msg("closing SQLite database connection")
	return s.db.Close()
}

// Begin starts a new transaction with the given context.
func (s *SQLiteDB) Begin(ctx context.Context) (database.Tx, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	return &sqliteTx{tx: tx}, nil
}

// Driver returns the name of the database driver.
func (s *SQLiteDB) Driver() string {
	return "sqlite"
}

// Commit commits the transaction.
func (t *sqliteTx) Commit() error {
	return t.tx.Commit()
}

// Rollback aborts the transaction.
func (t *sqliteTx) Rollback() error {
	return t.tx.Rollback()
}

// Tx returns the underlying *sql.Tx transaction.
func (t *sqliteTx) Tx() *sql.Tx {
	return t.tx
}
