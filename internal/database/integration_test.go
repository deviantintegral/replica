//go:build integration

// Package database_test provides integration tests for database functionality.
package database_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/deviantintegral/replica/internal/database"
	"github.com/deviantintegral/replica/internal/database/mysql"
	"github.com/deviantintegral/replica/internal/database/postgres"
	"github.com/deviantintegral/replica/internal/database/sqlite"
	"github.com/rs/zerolog"
)

// createTestDatabase creates a database connection based on environment variables.
// It returns the database and a boolean indicating if the test should be skipped.
func createTestDatabase(t *testing.T) (database.Database, bool) {
	t.Helper()

	driver := os.Getenv("TEST_DB_DRIVER")
	url := os.Getenv("TEST_DB_URL")

	if driver == "" || url == "" {
		return nil, true
	}

	logger := zerolog.Nop()
	cfg := database.Config{
		Driver: driver,
		URL:    url,
	}

	var db database.Database
	var err error

	switch driver {
	case "sqlite":
		db, err = sqlite.New(cfg, logger)
	case "mariadb":
		db, err = mysql.New(cfg, logger)
	case "postgres":
		db, err = postgres.New(cfg, logger)
	default:
		t.Fatalf("unsupported database driver: %s", driver)
	}

	if err != nil {
		t.Fatalf("failed to create database: %v", err)
	}

	// Run migrations (AutoMigrate equivalent)
	if err := db.Migrate(); err != nil {
		db.Close()
		t.Fatalf("failed to run migrations: %v", err)
	}

	return db, false
}

func TestIntegration_DatabaseConnection(t *testing.T) {
	db, skip := createTestDatabase(t)
	if skip {
		t.Skip("TEST_DB_DRIVER and TEST_DB_URL environment variables not set")
	}
	defer db.Close()

	// Test Ping with context timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.Ping(ctx); err != nil {
		t.Fatalf("Ping failed: %v", err)
	}

	// Test Begin transaction and Rollback
	tx, err := db.Begin(ctx)
	if err != nil {
		t.Fatalf("Begin transaction failed: %v", err)
	}

	if err := tx.Rollback(); err != nil {
		t.Fatalf("Rollback failed: %v", err)
	}
}

func TestIntegration_MigrationCreatesTable(t *testing.T) {
	db, skip := createTestDatabase(t)
	if skip {
		t.Skip("TEST_DB_DRIVER and TEST_DB_URL environment variables not set")
	}
	defer db.Close()

	// Query replica_metadata table to verify migration worked
	var value string
	err := db.DB().QueryRow("SELECT value FROM replica_metadata WHERE key = 'schema_version'").Scan(&value)
	if err != nil {
		t.Fatalf("failed to query replica_metadata table: %v", err)
	}

	if value != "1" {
		t.Errorf("expected schema_version to be '1', got '%s'", value)
	}
}
