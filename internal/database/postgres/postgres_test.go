package postgres

import (
	"testing"

	"github.com/deviantintegral/replica/internal/database"
	"github.com/rs/zerolog"
)

// testLogger returns a disabled logger for testing.
func testLogger() zerolog.Logger {
	return zerolog.Nop()
}

func TestNew_EmptyURL_ReturnsError(t *testing.T) {
	cfg := database.Config{
		Driver: "postgres",
		URL:    "",
	}

	_, err := New(cfg, testLogger())
	if err == nil {
		t.Error("New() with empty URL should return error, got nil")
	}
}

func TestPostgresDB_Driver(t *testing.T) {
	// Create a PostgresDB directly to test Driver() without needing a real connection
	db := &PostgresDB{
		db:     nil,
		logger: testLogger(),
	}

	got := db.Driver()
	want := "postgres"
	if got != want {
		t.Errorf("Driver() = %q, want %q", got, want)
	}
}

func TestPostgresDB_ImplementsDatabaseInterface(t *testing.T) {
	// This test verifies at compile time that PostgresDB implements database.Database
	var _ database.Database = (*PostgresDB)(nil)
}

func TestPostgresTx_ImplementsTxInterface(t *testing.T) {
	// This test verifies at compile time that postgresTx implements database.Tx
	var _ database.Tx = (*postgresTx)(nil)
}

func TestDefaultConstants(t *testing.T) {
	// Verify default values are set as expected
	if DefaultMaxOpenConns != 25 {
		t.Errorf("DefaultMaxOpenConns = %d, want 25", DefaultMaxOpenConns)
	}
	if DefaultMaxIdleConns != 5 {
		t.Errorf("DefaultMaxIdleConns = %d, want 5", DefaultMaxIdleConns)
	}
	// 5 minutes in nanoseconds
	expectedLifetime := 5 * 60 * 1000000000
	if int(DefaultConnMaxLifetime) != expectedLifetime {
		t.Errorf("DefaultConnMaxLifetime = %v, want 5 minutes", DefaultConnMaxLifetime)
	}
}
