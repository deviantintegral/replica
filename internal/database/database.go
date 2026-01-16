// Package database provides database interface abstractions and common types
// for the replica application. It defines the contract that all database
// drivers must implement.
package database

import (
	"context"
	"database/sql"
	"time"
)

// Database defines the interface for database operations.
// All database drivers must implement this interface.
type Database interface {
	// DB returns the underlying *sql.DB connection pool.
	DB() *sql.DB

	// Migrate runs database migrations. Returns nil until migrations are implemented.
	Migrate() error

	// Ping verifies the database connection is alive.
	Ping(ctx context.Context) error

	// Close closes the database connection and releases resources.
	Close() error

	// Begin starts a new transaction with the given context.
	Begin(ctx context.Context) (Tx, error)

	// Driver returns the name of the database driver (e.g., "sqlite", "postgres", "mysql").
	Driver() string
}

// Tx defines the interface for database transactions.
type Tx interface {
	// Commit commits the transaction.
	Commit() error

	// Rollback aborts the transaction.
	Rollback() error

	// Tx returns the underlying *sql.Tx transaction.
	Tx() *sql.Tx
}

// Config holds database connection configuration.
type Config struct {
	// Driver specifies the database driver to use (e.g., "sqlite", "postgres", "mysql").
	Driver string

	// URL is the database connection URL or file path.
	// For SQLite, this can be a file path or ":memory:" for in-memory databases.
	URL string

	// MaxOpenConns sets the maximum number of open connections to the database.
	// For SQLite, this should typically be 1.
	MaxOpenConns int

	// MaxIdleConns sets the maximum number of connections in the idle connection pool.
	MaxIdleConns int

	// ConnMaxLifetime sets the maximum amount of time a connection may be reused.
	ConnMaxLifetime time.Duration
}
