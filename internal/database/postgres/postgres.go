// Package postgres provides a PostgreSQL database driver implementation
// for the replica application.
package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/deviantintegral/replica/internal/database"
	_ "github.com/jackc/pgx/v5/stdlib" // PostgreSQL driver
	"github.com/rs/zerolog"
)

const (
	// DefaultMaxOpenConns is the default maximum number of open connections.
	// PostgreSQL handles concurrent connections well, so we allow more connections.
	DefaultMaxOpenConns = 25

	// DefaultMaxIdleConns is the default maximum number of idle connections.
	DefaultMaxIdleConns = 5

	// DefaultConnMaxLifetime is the default maximum connection lifetime.
	DefaultConnMaxLifetime = 5 * time.Minute
)

// PostgresDB implements the database.Database interface for PostgreSQL.
type PostgresDB struct {
	db     *sql.DB
	logger zerolog.Logger
}

// postgresTx implements the database.Tx interface for PostgreSQL transactions.
type postgresTx struct {
	tx *sql.Tx
}

// New creates a new PostgreSQL database connection with the given configuration.
// It configures the connection pool with sensible defaults for PostgreSQL.
func New(cfg database.Config, logger zerolog.Logger) (*PostgresDB, error) {
	logger = logger.With().Str("component", "postgres").Logger()

	if cfg.URL == "" {
		return nil, fmt.Errorf("database URL is required")
	}

	logger.Debug().Str("url", cfg.URL).Msg("opening PostgreSQL database")

	db, err := sql.Open("pgx", cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to open PostgreSQL database: %w", err)
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

	logger.Info().
		Str("url", cfg.URL).
		Int("max_open_conns", maxOpen).
		Int("max_idle_conns", maxIdle).
		Dur("conn_max_lifetime", connMaxLifetime).
		Msg("PostgreSQL database configured")

	return &PostgresDB{
		db:     db,
		logger: logger,
	}, nil
}

// DB returns the underlying *sql.DB connection pool.
func (p *PostgresDB) DB() *sql.DB {
	return p.db
}

// Migrate runs database migrations using golang-migrate.
// It applies all pending migrations from the embedded SQL files.
func (p *PostgresDB) Migrate() error {
	p.logger.Debug().Msg("running database migrations")
	if err := database.RunMigrations(p); err != nil {
		return err
	}
	p.logger.Info().Msg("database migrations completed successfully")
	return nil
}

// Ping verifies the database connection is alive.
func (p *PostgresDB) Ping(ctx context.Context) error {
	return p.db.PingContext(ctx)
}

// Close closes the database connection and releases resources.
func (p *PostgresDB) Close() error {
	p.logger.Debug().Msg("closing PostgreSQL database connection")
	return p.db.Close()
}

// Begin starts a new transaction with the given context.
func (p *PostgresDB) Begin(ctx context.Context) (database.Tx, error) {
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	return &postgresTx{tx: tx}, nil
}

// Driver returns the name of the database driver.
func (p *PostgresDB) Driver() string {
	return "postgres"
}

// Commit commits the transaction.
func (t *postgresTx) Commit() error {
	return t.tx.Commit()
}

// Rollback aborts the transaction.
func (t *postgresTx) Rollback() error {
	return t.tx.Rollback()
}

// Tx returns the underlying *sql.Tx transaction.
func (t *postgresTx) Tx() *sql.Tx {
	return t.tx
}
