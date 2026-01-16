// Package testutil provides testing utilities for the replica application.
package testutil

import (
	"database/sql"
	"testing"

	"github.com/deviantintegral/replica/internal/database"
	"github.com/deviantintegral/replica/internal/database/sqlite"
	"github.com/rs/zerolog"
)

// TestDB creates an in-memory SQLite database for testing.
// It runs migrations and registers cleanup to close the database when the test completes.
func TestDB(t *testing.T) database.Database {
	t.Helper()
	return TestDBWithLogger(t, zerolog.Nop())
}

// TestDBWithLogger creates an in-memory SQLite database for testing with a custom logger.
// It runs migrations and registers cleanup to close the database when the test completes.
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
		t.Fatalf("failed to run migrations: %v", err)
	}

	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Errorf("failed to close test database: %v", err)
		}
	})

	return db
}

// MustExec executes a SQL query and fails the test if an error occurs.
func MustExec(t *testing.T, db database.Database, query string, args ...interface{}) {
	t.Helper()

	_, err := db.DB().Exec(query, args...)
	if err != nil {
		t.Fatalf("MustExec failed: %v\nQuery: %s", err, query)
	}
}

// MustQuery executes a SQL query and returns the rows, failing the test if an error occurs.
// The caller is responsible for closing the returned rows.
func MustQuery(t *testing.T, db database.Database, query string, args ...interface{}) *sql.Rows {
	t.Helper()

	rows, err := db.DB().Query(query, args...)
	if err != nil {
		t.Fatalf("MustQuery failed: %v\nQuery: %s", err, query)
	}

	return rows
}
