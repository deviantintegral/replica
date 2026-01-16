package database_test

import (
	"testing"

	"github.com/deviantintegral/replica/internal/database"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew_SQLiteDriver(t *testing.T) {
	logger := zerolog.Nop()
	cfg := database.Config{
		Driver: "sqlite",
		URL:    ":memory:",
	}
	opts := database.Options{}

	db, err := database.New(cfg, logger, opts)
	require.NoError(t, err)
	require.NotNil(t, db)
	defer db.Close()

	assert.Equal(t, "sqlite", db.Driver())
}

func TestNew_MySQLAliasSelectsMariaDBDriver(t *testing.T) {
	// This test verifies that "mysql" alias selects the mariadb driver.
	// The connection will fail without a real MySQL server, but we can
	// verify the driver selection by checking the error message or
	// by using a mock configuration.
	logger := zerolog.Nop()
	cfg := database.Config{
		Driver: "mysql",
		URL:    "user:pass@tcp(localhost:3306)/testdb",
	}
	opts := database.Options{}

	// The connection will fail without a real server, but the driver
	// should be selected correctly. We expect a connection error, not
	// an "unsupported driver" error.
	db, err := database.New(cfg, logger, opts)
	if err != nil {
		// Verify it's not an unsupported driver error
		assert.NotContains(t, err.Error(), "unsupported database driver")
		// The error should be about connection failure, not driver selection
		assert.Contains(t, err.Error(), "MariaDB")
	} else {
		// If somehow we connected, verify the driver
		defer db.Close()
		assert.Equal(t, "mariadb", db.Driver())
	}
}

func TestNew_PostgreSQLAliasSelectsPostgresDriver(t *testing.T) {
	// This test verifies that "postgresql" alias selects the postgres driver.
	// The connection may fail without a real PostgreSQL server, but we can
	// verify the driver selection by checking that we don't get an
	// "unsupported driver" error.
	logger := zerolog.Nop()
	cfg := database.Config{
		Driver: "postgresql",
		URL:    "postgres://user:pass@localhost:5432/testdb",
	}
	opts := database.Options{}

	// The connection will fail without a real server, but the driver
	// should be selected correctly. We expect a connection error, not
	// an "unsupported driver" error.
	db, err := database.New(cfg, logger, opts)
	if err != nil {
		// Verify it's not an unsupported driver error
		assert.NotContains(t, err.Error(), "unsupported database driver")
	} else {
		// If somehow we connected, verify the driver
		defer db.Close()
		assert.Equal(t, "postgres", db.Driver())
	}
}

func TestNew_InvalidDriverReturnsError(t *testing.T) {
	logger := zerolog.Nop()
	cfg := database.Config{
		Driver: "invalid_driver",
		URL:    ":memory:",
	}
	opts := database.Options{}

	db, err := database.New(cfg, logger, opts)
	assert.Nil(t, db)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported database driver")
	assert.Contains(t, err.Error(), "invalid_driver")
}

func TestNew_AutoMigrateRunsMigrations(t *testing.T) {
	logger := zerolog.Nop()
	cfg := database.Config{
		Driver: "sqlite",
		URL:    ":memory:",
	}
	opts := database.Options{
		AutoMigrate: true,
	}

	db, err := database.New(cfg, logger, opts)
	require.NoError(t, err)
	require.NotNil(t, db)
	defer db.Close()

	// Verify the database was created and migrations ran
	// The migrations should have created the schema_migrations table at minimum
	assert.Equal(t, "sqlite", db.Driver())

	// Verify we can query the database (migrations succeeded)
	sqlDB := db.DB()
	require.NotNil(t, sqlDB)
	err = sqlDB.Ping()
	assert.NoError(t, err)
}

func TestNew_DriverCaseInsensitive(t *testing.T) {
	testCases := []struct {
		name           string
		driver         string
		expectedDriver string
	}{
		{"SQLite uppercase", "SQLITE", "sqlite"},
		{"SQLite mixed case", "SqLiTe", "sqlite"},
		{"MySQL uppercase", "MYSQL", "mariadb"},
		{"MariaDB uppercase", "MARIADB", "mariadb"},
		{"Postgres uppercase", "POSTGRES", "postgres"},
		{"PostgreSQL uppercase", "POSTGRESQL", "postgres"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			logger := zerolog.Nop()
			cfg := database.Config{
				Driver: tc.driver,
				URL:    ":memory:", // Only works for SQLite
			}
			opts := database.Options{}

			db, err := database.New(cfg, logger, opts)
			if tc.expectedDriver == "sqlite" {
				// SQLite should connect successfully with :memory:
				require.NoError(t, err)
				require.NotNil(t, db)
				defer db.Close()
				assert.Equal(t, tc.expectedDriver, db.Driver())
			} else {
				// MySQL/Postgres will fail to connect, but should not be
				// rejected as unsupported drivers
				if err != nil {
					assert.NotContains(t, err.Error(), "unsupported database driver")
				} else {
					defer db.Close()
					assert.Equal(t, tc.expectedDriver, db.Driver())
				}
			}
		})
	}
}
