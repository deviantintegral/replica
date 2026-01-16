package sqlite

import (
	"context"
	"testing"
	"time"

	"github.com/deviantintegral/replica/internal/database"
	"github.com/rs/zerolog"
)

// testLogger returns a disabled logger for testing.
func testLogger() zerolog.Logger {
	return zerolog.Nop()
}

// testConfig returns a test configuration using an in-memory database.
func testConfig() database.Config {
	return database.Config{
		Driver: "sqlite",
		URL:    ":memory:",
	}
}

func TestNew_InMemoryDatabase(t *testing.T) {
	cfg := testConfig()
	db, err := New(cfg, testLogger())
	if err != nil {
		t.Fatalf("New() error = %v, want nil", err)
	}
	defer db.Close()

	if db.db == nil {
		t.Error("New() db.db = nil, want non-nil")
	}
}

func TestNew_WithCustomConfig(t *testing.T) {
	cfg := database.Config{
		Driver:          "sqlite",
		URL:             ":memory:",
		MaxOpenConns:    5,
		MaxIdleConns:    3,
		ConnMaxLifetime: time.Minute,
	}

	db, err := New(cfg, testLogger())
	if err != nil {
		t.Fatalf("New() error = %v, want nil", err)
	}
	defer db.Close()

	// Verify the database is accessible
	if err := db.Ping(context.Background()); err != nil {
		t.Errorf("Ping() error = %v, want nil", err)
	}
}

func TestSQLiteDB_Ping(t *testing.T) {
	db, err := New(testConfig(), testLogger())
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	if err := db.Ping(ctx); err != nil {
		t.Errorf("Ping() error = %v, want nil", err)
	}
}

func TestSQLiteDB_Ping_WithTimeout(t *testing.T) {
	db, err := New(testConfig(), testLogger())
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.Ping(ctx); err != nil {
		t.Errorf("Ping() with timeout error = %v, want nil", err)
	}
}

func TestSQLiteDB_Driver(t *testing.T) {
	db, err := New(testConfig(), testLogger())
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer db.Close()

	got := db.Driver()
	want := "sqlite"
	if got != want {
		t.Errorf("Driver() = %q, want %q", got, want)
	}
}

func TestSQLiteDB_DB(t *testing.T) {
	db, err := New(testConfig(), testLogger())
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer db.Close()

	sqlDB := db.DB()
	if sqlDB == nil {
		t.Error("DB() = nil, want non-nil")
	}
}

func TestSQLiteDB_Close(t *testing.T) {
	db, err := New(testConfig(), testLogger())
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	if err := db.Close(); err != nil {
		t.Errorf("Close() error = %v, want nil", err)
	}

	// After close, ping should fail
	if err := db.Ping(context.Background()); err == nil {
		t.Error("Ping() after Close() should fail, but got nil error")
	}
}

func TestSQLiteDB_Migrate(t *testing.T) {
	db, err := New(testConfig(), testLogger())
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer db.Close()

	// Migrate should return nil for now
	if err := db.Migrate(); err != nil {
		t.Errorf("Migrate() error = %v, want nil", err)
	}
}

func TestSQLiteDB_Begin(t *testing.T) {
	db, err := New(testConfig(), testLogger())
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	tx, err := db.Begin(ctx)
	if err != nil {
		t.Fatalf("Begin() error = %v, want nil", err)
	}

	if tx == nil {
		t.Fatal("Begin() returned nil transaction")
	}

	// Clean up
	if err := tx.Rollback(); err != nil {
		t.Errorf("Rollback() error = %v", err)
	}
}

func TestSQLiteDB_Transaction_Commit(t *testing.T) {
	db, err := New(testConfig(), testLogger())
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer db.Close()

	// Create a test table
	_, err = db.DB().Exec("CREATE TABLE test_commit (id INTEGER PRIMARY KEY, value TEXT)")
	if err != nil {
		t.Fatalf("failed to create test table: %v", err)
	}

	// Start transaction
	ctx := context.Background()
	tx, err := db.Begin(ctx)
	if err != nil {
		t.Fatalf("Begin() error = %v", err)
	}

	// Insert data in transaction
	_, err = tx.Tx().Exec("INSERT INTO test_commit (value) VALUES (?)", "test_value")
	if err != nil {
		t.Fatalf("failed to insert: %v", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		t.Fatalf("Commit() error = %v, want nil", err)
	}

	// Verify data persisted
	var value string
	err = db.DB().QueryRow("SELECT value FROM test_commit WHERE id = 1").Scan(&value)
	if err != nil {
		t.Fatalf("failed to query: %v", err)
	}
	if value != "test_value" {
		t.Errorf("value = %q, want %q", value, "test_value")
	}
}

func TestSQLiteDB_Transaction_Rollback(t *testing.T) {
	db, err := New(testConfig(), testLogger())
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer db.Close()

	// Create a test table
	_, err = db.DB().Exec("CREATE TABLE test_rollback (id INTEGER PRIMARY KEY, value TEXT)")
	if err != nil {
		t.Fatalf("failed to create test table: %v", err)
	}

	// Start transaction
	ctx := context.Background()
	tx, err := db.Begin(ctx)
	if err != nil {
		t.Fatalf("Begin() error = %v", err)
	}

	// Insert data in transaction
	_, err = tx.Tx().Exec("INSERT INTO test_rollback (value) VALUES (?)", "test_value")
	if err != nil {
		t.Fatalf("failed to insert: %v", err)
	}

	// Rollback transaction
	if err := tx.Rollback(); err != nil {
		t.Fatalf("Rollback() error = %v, want nil", err)
	}

	// Verify data was NOT persisted
	var count int
	err = db.DB().QueryRow("SELECT COUNT(*) FROM test_rollback").Scan(&count)
	if err != nil {
		t.Fatalf("failed to query: %v", err)
	}
	if count != 0 {
		t.Errorf("count = %d, want 0 (rollback should have prevented insert)", count)
	}
}

func TestSqliteTx_Tx(t *testing.T) {
	db, err := New(testConfig(), testLogger())
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	tx, err := db.Begin(ctx)
	if err != nil {
		t.Fatalf("Begin() error = %v", err)
	}
	defer tx.Rollback()

	if tx.Tx() == nil {
		t.Error("Tx() = nil, want non-nil")
	}
}

func TestBuildDSN(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		wantPart string
	}{
		{
			name:     "memory database",
			url:      ":memory:",
			wantPart: ":memory:?_journal_mode=WAL&_busy_timeout=5000&_foreign_keys=on",
		},
		{
			name:     "file database",
			url:      "/tmp/test.db",
			wantPart: "/tmp/test.db?_journal_mode=WAL&_busy_timeout=5000&_foreign_keys=on",
		},
		{
			name:     "url with existing params",
			url:      "/tmp/test.db?mode=rw",
			wantPart: "/tmp/test.db?mode=rw&_journal_mode=WAL&_busy_timeout=5000&_foreign_keys=on",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildDSN(tt.url)
			if got != tt.wantPart {
				t.Errorf("buildDSN(%q) = %q, want %q", tt.url, got, tt.wantPart)
			}
		})
	}
}

func TestSQLiteDB_ImplementsDatabaseInterface(t *testing.T) {
	db, err := New(testConfig(), testLogger())
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer db.Close()

	// This test verifies at compile time that SQLiteDB implements database.Database
	var _ database.Database = db
}

func TestSqliteTx_ImplementsTxInterface(t *testing.T) {
	db, err := New(testConfig(), testLogger())
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer db.Close()

	tx, err := db.Begin(context.Background())
	if err != nil {
		t.Fatalf("Begin() error = %v", err)
	}
	defer tx.Rollback()

	// This test verifies at compile time that sqliteTx implements database.Tx
	var _ database.Tx = tx
}
