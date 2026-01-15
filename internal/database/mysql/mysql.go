// Package mysql provides a MariaDB/MySQL database driver implementation
// for the replica application.
package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/deviantintegral/replica/internal/database"
	_ "github.com/go-sql-driver/mysql" // MySQL/MariaDB driver
	"github.com/rs/zerolog"
)

const (
	// DefaultMaxOpenConns is the default maximum number of open connections.
	DefaultMaxOpenConns = 25

	// DefaultMaxIdleConns is the default maximum number of idle connections.
	DefaultMaxIdleConns = 5

	// DefaultConnMaxLifetime is the default maximum connection lifetime.
	DefaultConnMaxLifetime = 5 * time.Minute
)

// MariaDB implements the database.Database interface for MariaDB/MySQL.
type MariaDB struct {
	db     *sql.DB
	logger zerolog.Logger
}

// mariadbTx implements the database.Tx interface for MariaDB transactions.
type mariadbTx struct {
	tx *sql.Tx
}

// New creates a new MariaDB database connection with the given configuration.
// It automatically appends parseTime=true to the DSN if not present for proper
// time handling. Returns an error if the URL is empty.
func New(cfg database.Config, logger zerolog.Logger) (*MariaDB, error) {
	logger = logger.With().Str("component", "mariadb").Logger()

	if cfg.URL == "" {
		return nil, fmt.Errorf("database URL is required")
	}

	// Build connection string with parseTime parameter
	dsn := buildDSN(cfg.URL)
	logger.Debug().Str("dsn", maskDSN(dsn)).Msg("opening MariaDB database")

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open MariaDB database: %w", err)
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

	connMaxLifetime := cfg.ConnMaxLifetime
	if connMaxLifetime <= 0 {
		connMaxLifetime = DefaultConnMaxLifetime
	}
	db.SetConnMaxLifetime(connMaxLifetime)

	// Verify connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping MariaDB database: %w", err)
	}

	logger.Info().Str("url", maskDSN(cfg.URL)).Msg("MariaDB database connected")

	return &MariaDB{
		db:     db,
		logger: logger,
	}, nil
}

// buildDSN constructs the MySQL connection string, ensuring parseTime=true is set.
// This is required for proper time.Time scanning from DATE/DATETIME columns.
func buildDSN(url string) string {
	// Check if parseTime is already set
	if containsParam(url, "parseTime") {
		return url
	}

	// Determine the separator based on whether URL already has parameters
	separator := "?"
	if containsRune(url, '?') {
		separator = "&"
	}

	return url + separator + "parseTime=true"
}

// containsParam checks if a DSN contains a specific parameter.
func containsParam(dsn, param string) bool {
	// Check for param at start of query string: ?param= or ?param&
	if strings.Contains(dsn, "?"+param+"=") || strings.Contains(dsn, "?"+param+"&") {
		return true
	}
	// Check for param in middle/end of query string: &param= or &param&
	if strings.Contains(dsn, "&"+param+"=") || strings.Contains(dsn, "&"+param+"&") {
		return true
	}
	// Check for param at end without value: ?param or &param (at end of string)
	if strings.HasSuffix(dsn, "?"+param) || strings.HasSuffix(dsn, "&"+param) {
		return true
	}
	return false
}

// containsRune checks if a string contains a specific rune.
func containsRune(s string, r rune) bool {
	return strings.ContainsRune(s, r)
}

// maskDSN masks sensitive information in DSN for logging.
func maskDSN(dsn string) string {
	// Look for password in format user:password@
	atIdx := strings.Index(dsn, "@")
	if atIdx == -1 {
		return dsn
	}

	colonIdx := strings.Index(dsn, ":")
	if colonIdx == -1 || colonIdx > atIdx {
		return dsn
	}

	// Mask password between first colon and @ symbol
	return dsn[:colonIdx+1] + "***" + dsn[atIdx:]
}

// DB returns the underlying *sql.DB connection pool.
func (m *MariaDB) DB() *sql.DB {
	return m.db
}

// Migrate runs database migrations using golang-migrate.
// It applies all pending migrations from the embedded SQL files.
func (m *MariaDB) Migrate() error {
	m.logger.Debug().Msg("running database migrations")
	if err := database.RunMigrations(m); err != nil {
		return err
	}
	m.logger.Info().Msg("database migrations completed successfully")
	return nil
}

// Ping verifies the database connection is alive.
func (m *MariaDB) Ping(ctx context.Context) error {
	return m.db.PingContext(ctx)
}

// Close closes the database connection and releases resources.
func (m *MariaDB) Close() error {
	m.logger.Debug().Msg("closing MariaDB database connection")
	return m.db.Close()
}

// Begin starts a new transaction with the given context.
func (m *MariaDB) Begin(ctx context.Context) (database.Tx, error) {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	return &mariadbTx{tx: tx}, nil
}

// Driver returns the name of the database driver.
func (m *MariaDB) Driver() string {
	return "mariadb"
}

// Commit commits the transaction.
func (t *mariadbTx) Commit() error {
	return t.tx.Commit()
}

// Rollback aborts the transaction.
func (t *mariadbTx) Rollback() error {
	return t.tx.Rollback()
}

// Tx returns the underlying *sql.Tx transaction.
func (t *mariadbTx) Tx() *sql.Tx {
	return t.tx
}
