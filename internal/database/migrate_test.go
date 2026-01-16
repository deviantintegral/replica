package database_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/deviantintegral/replica/internal/database"
	"github.com/deviantintegral/replica/internal/database/sqlite"
	"github.com/rs/zerolog"
)

func TestRunMigrations(t *testing.T) {
	logger := zerolog.Nop()

	// Create an in-memory SQLite database
	db, err := sqlite.New(database.Config{
		Driver: "sqlite",
		URL:    ":memory:",
	}, logger)
	if err != nil {
		t.Fatalf("failed to create SQLite database: %v", err)
	}
	defer db.Close()

	// Run migrations
	err = database.RunMigrations(db)
	if err != nil {
		t.Fatalf("RunMigrations failed: %v", err)
	}

	// Verify replica_metadata table exists by querying it
	var value string
	err = db.DB().QueryRow("SELECT value FROM replica_metadata WHERE key = 'schema_version'").Scan(&value)
	if err != nil {
		t.Fatalf("failed to query replica_metadata table: %v", err)
	}
	if value != "1" {
		t.Errorf("expected schema_version to be '1', got '%s'", value)
	}
}

func TestRunMigrations_Idempotent(t *testing.T) {
	logger := zerolog.Nop()

	// Create an in-memory SQLite database
	db, err := sqlite.New(database.Config{
		Driver: "sqlite",
		URL:    ":memory:",
	}, logger)
	if err != nil {
		t.Fatalf("failed to create SQLite database: %v", err)
	}
	defer db.Close()

	// Run migrations first time
	err = database.RunMigrations(db)
	if err != nil {
		t.Fatalf("first RunMigrations failed: %v", err)
	}

	// Run migrations second time - should succeed (idempotent)
	err = database.RunMigrations(db)
	if err != nil {
		t.Fatalf("second RunMigrations failed (not idempotent): %v", err)
	}
}

func TestRunMigrations_ReplicaMetadataTableExists(t *testing.T) {
	logger := zerolog.Nop()

	// Create an in-memory SQLite database
	db, err := sqlite.New(database.Config{
		Driver: "sqlite",
		URL:    ":memory:",
	}, logger)
	if err != nil {
		t.Fatalf("failed to create SQLite database: %v", err)
	}
	defer db.Close()

	// Run migrations
	err = database.RunMigrations(db)
	if err != nil {
		t.Fatalf("RunMigrations failed: %v", err)
	}

	// Verify replica_metadata table exists by checking sqlite_master
	var tableName string
	err = db.DB().QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='replica_metadata'").Scan(&tableName)
	if err != nil {
		t.Fatalf("replica_metadata table does not exist: %v", err)
	}
	if tableName != "replica_metadata" {
		t.Errorf("expected table name 'replica_metadata', got '%s'", tableName)
	}
}

func TestRunMigrations_UnsupportedDriver(t *testing.T) {
	// Create a mock database with an unsupported driver
	mock := &mockDatabase{driver: "unsupported"}

	err := database.RunMigrations(mock)
	if err == nil {
		t.Fatal("expected error for unsupported driver, got nil")
	}
}

// mockDatabase is a minimal implementation for testing unsupported drivers
type mockDatabase struct {
	driver string
}

func (m *mockDatabase) DB() *sql.DB                            { return nil }
func (m *mockDatabase) Migrate() error                         { return nil }
func (m *mockDatabase) Ping(_ context.Context) error           { return nil }
func (m *mockDatabase) Close() error                           { return nil }
func (m *mockDatabase) Begin(_ context.Context) (database.Tx, error) { return nil, nil }
func (m *mockDatabase) Driver() string                         { return m.driver }
