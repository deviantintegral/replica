package testutil

import (
	"context"
	"testing"
)

func TestTestDB_CreatesValidDatabase(t *testing.T) {
	db := TestDB(t)

	// Verify database is valid by pinging it
	err := db.Ping(context.Background())
	if err != nil {
		t.Fatalf("ping failed: %v", err)
	}

	// Verify driver is correct
	if got := db.Driver(); got != "sqlite" {
		t.Errorf("got driver %q, want %q", got, "sqlite")
	}
}

func TestMustExec_ExecutesQueries(t *testing.T) {
	db := TestDB(t)

	// Create a test table
	MustExec(t, db, "CREATE TABLE test_table (id INTEGER PRIMARY KEY, name TEXT)")

	// Insert a row
	MustExec(t, db, "INSERT INTO test_table (name) VALUES (?)", "test_value")

	// Verify the row was inserted
	rows := MustQuery(t, db, "SELECT name FROM test_table WHERE id = 1")
	defer rows.Close()

	if !rows.Next() {
		t.Fatal("expected a row, got none")
	}

	var name string
	if err := rows.Scan(&name); err != nil {
		t.Fatalf("failed to scan row: %v", err)
	}

	if name != "test_value" {
		t.Errorf("got name %q, want %q", name, "test_value")
	}
}

func TestMultipleTestDB_CreatesIsolatedDatabases(t *testing.T) {
	// Create first database and add a table
	db1 := TestDB(t)
	MustExec(t, db1, "CREATE TABLE isolation_test (id INTEGER PRIMARY KEY)")

	// Create second database
	db2 := TestDB(t)

	// Verify the table does NOT exist in db2 (isolation check)
	rows := MustQuery(t, db2, "SELECT name FROM sqlite_master WHERE type='table' AND name='isolation_test'")
	defer rows.Close()

	if rows.Next() {
		t.Error("expected databases to be isolated, but table exists in second database")
	}

	// Verify the table DOES exist in db1
	rows1 := MustQuery(t, db1, "SELECT name FROM sqlite_master WHERE type='table' AND name='isolation_test'")
	defer rows1.Close()

	if !rows1.Next() {
		t.Error("expected table to exist in first database")
	}
}
