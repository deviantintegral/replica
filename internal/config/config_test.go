package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefault(t *testing.T) {
	cfg := Default()

	// Test instance defaults
	if cfg.Instance.Name != "replica" {
		t.Errorf("Instance.Name = %q, want %q", cfg.Instance.Name, "replica")
	}

	// Test database defaults
	if cfg.Database.Driver != "sqlite" {
		t.Errorf("Database.Driver = %q, want %q", cfg.Database.Driver, "sqlite")
	}
	if cfg.Database.URL != "replica.db" {
		t.Errorf("Database.URL = %q, want %q", cfg.Database.URL, "replica.db")
	}

	// Test server defaults
	if cfg.Server.Host != "0.0.0.0" {
		t.Errorf("Server.Host = %q, want %q", cfg.Server.Host, "0.0.0.0")
	}
	if cfg.Server.Port != 8080 {
		t.Errorf("Server.Port = %d, want %d", cfg.Server.Port, 8080)
	}

	// Test logging defaults
	if cfg.Logging.Level != "info" {
		t.Errorf("Logging.Level = %q, want %q", cfg.Logging.Level, "info")
	}
	if cfg.Logging.Format != "json" {
		t.Errorf("Logging.Format = %q, want %q", cfg.Logging.Format, "json")
	}
}

func TestLoad_YAMLFile(t *testing.T) {
	// Create a temporary YAML config file
	yamlContent := `
instance:
  name: "test-instance"

database:
  driver: postgres
  url: "postgres://localhost/testdb"

server:
  host: "127.0.0.1"
  port: 9090

logging:
  level: debug
  format: text
`
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to write temp config file: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Verify YAML values were loaded
	if cfg.Instance.Name != "test-instance" {
		t.Errorf("Instance.Name = %q, want %q", cfg.Instance.Name, "test-instance")
	}
	if cfg.Database.Driver != "postgres" {
		t.Errorf("Database.Driver = %q, want %q", cfg.Database.Driver, "postgres")
	}
	if cfg.Database.URL != "postgres://localhost/testdb" {
		t.Errorf("Database.URL = %q, want %q", cfg.Database.URL, "postgres://localhost/testdb")
	}
	if cfg.Server.Host != "127.0.0.1" {
		t.Errorf("Server.Host = %q, want %q", cfg.Server.Host, "127.0.0.1")
	}
	if cfg.Server.Port != 9090 {
		t.Errorf("Server.Port = %d, want %d", cfg.Server.Port, 9090)
	}
	if cfg.Logging.Level != "debug" {
		t.Errorf("Logging.Level = %q, want %q", cfg.Logging.Level, "debug")
	}
	if cfg.Logging.Format != "text" {
		t.Errorf("Logging.Format = %q, want %q", cfg.Logging.Format, "text")
	}
}

func TestLoad_PartialYAML(t *testing.T) {
	// Test that partial YAML uses defaults for missing values
	yamlContent := `
instance:
  name: "partial-test"

database:
  driver: mariadb
`
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to write temp config file: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Verify YAML values were loaded
	if cfg.Instance.Name != "partial-test" {
		t.Errorf("Instance.Name = %q, want %q", cfg.Instance.Name, "partial-test")
	}
	if cfg.Database.Driver != "mariadb" {
		t.Errorf("Database.Driver = %q, want %q", cfg.Database.Driver, "mariadb")
	}

	// Verify defaults for missing values
	if cfg.Database.URL != "replica.db" {
		t.Errorf("Database.URL = %q, want default %q", cfg.Database.URL, "replica.db")
	}
	if cfg.Server.Port != 8080 {
		t.Errorf("Server.Port = %d, want default %d", cfg.Server.Port, 8080)
	}
}

func TestLoad_EnvOverrides(t *testing.T) {
	// Create a YAML config file with values that will be overridden
	yamlContent := `
instance:
  name: "yaml-instance"

database:
  driver: sqlite
  url: "yaml.db"

server:
  host: "0.0.0.0"
  port: 8080

logging:
  level: info
  format: json
`
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to write temp config file: %v", err)
	}

	// Set environment variables
	envVars := map[string]string{
		"REPLICA_INSTANCE_NAME":   "env-instance",
		"REPLICA_DATABASE_DRIVER": "postgres",
		"REPLICA_DATABASE_URL":    "postgres://env/db",
		"REPLICA_SERVER_HOST":     "localhost",
		"REPLICA_SERVER_PORT":     "3000",
		"REPLICA_LOGGING_LEVEL":   "debug",
		"REPLICA_LOGGING_FORMAT":  "text",
	}

	for k, v := range envVars {
		os.Setenv(k, v)
	}
	defer func() {
		for k := range envVars {
			os.Unsetenv(k)
		}
	}()

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Verify environment variables override YAML values
	if cfg.Instance.Name != "env-instance" {
		t.Errorf("Instance.Name = %q, want %q (env override)", cfg.Instance.Name, "env-instance")
	}
	if cfg.Database.Driver != "postgres" {
		t.Errorf("Database.Driver = %q, want %q (env override)", cfg.Database.Driver, "postgres")
	}
	if cfg.Database.URL != "postgres://env/db" {
		t.Errorf("Database.URL = %q, want %q (env override)", cfg.Database.URL, "postgres://env/db")
	}
	if cfg.Server.Host != "localhost" {
		t.Errorf("Server.Host = %q, want %q (env override)", cfg.Server.Host, "localhost")
	}
	if cfg.Server.Port != 3000 {
		t.Errorf("Server.Port = %d, want %d (env override)", cfg.Server.Port, 3000)
	}
	if cfg.Logging.Level != "debug" {
		t.Errorf("Logging.Level = %q, want %q (env override)", cfg.Logging.Level, "debug")
	}
	if cfg.Logging.Format != "text" {
		t.Errorf("Logging.Format = %q, want %q (env override)", cfg.Logging.Format, "text")
	}
}

func TestLoad_EnvOverridesDefaults(t *testing.T) {
	// Test that env vars override defaults when no YAML file is provided
	envVars := map[string]string{
		"REPLICA_INSTANCE_NAME": "env-only-instance",
		"REPLICA_SERVER_PORT":   "5000",
	}

	for k, v := range envVars {
		os.Setenv(k, v)
	}
	defer func() {
		for k := range envVars {
			os.Unsetenv(k)
		}
	}()

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Instance.Name != "env-only-instance" {
		t.Errorf("Instance.Name = %q, want %q (env override of default)", cfg.Instance.Name, "env-only-instance")
	}
	if cfg.Server.Port != 5000 {
		t.Errorf("Server.Port = %d, want %d (env override of default)", cfg.Server.Port, 5000)
	}

	// Verify defaults are still used for non-overridden values
	if cfg.Database.Driver != "sqlite" {
		t.Errorf("Database.Driver = %q, want default %q", cfg.Database.Driver, "sqlite")
	}
}

func TestLoad_InvalidYAMLFile(t *testing.T) {
	yamlContent := `
invalid: yaml: content: [
`
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to write temp config file: %v", err)
	}

	_, err := Load(configPath)
	if err == nil {
		t.Error("Load() expected error for invalid YAML, got nil")
	}
}

func TestLoad_NonexistentFile(t *testing.T) {
	_, err := Load("/nonexistent/path/config.yaml")
	if err == nil {
		t.Error("Load() expected error for nonexistent file, got nil")
	}
}

func TestValidate_InvalidDatabaseDriver(t *testing.T) {
	cfg := Default()
	cfg.Database.Driver = "mysql" // Invalid, should be mariadb

	err := cfg.Validate()
	if err == nil {
		t.Error("Validate() expected error for invalid database driver, got nil")
	}
	if err != nil && !contains(err.Error(), "invalid database driver") {
		t.Errorf("Validate() error = %q, want error containing 'invalid database driver'", err.Error())
	}
}

func TestValidate_ValidDatabaseDrivers(t *testing.T) {
	drivers := []string{"sqlite", "mariadb", "postgres"}

	for _, driver := range drivers {
		t.Run(driver, func(t *testing.T) {
			cfg := Default()
			cfg.Database.Driver = driver

			err := cfg.Validate()
			if err != nil {
				t.Errorf("Validate() error = %v for valid driver %q", err, driver)
			}
		})
	}
}

func TestValidate_InvalidLoggingLevel(t *testing.T) {
	cfg := Default()
	cfg.Logging.Level = "verbose"

	err := cfg.Validate()
	if err == nil {
		t.Error("Validate() expected error for invalid logging level, got nil")
	}
	if err != nil && !contains(err.Error(), "invalid logging level") {
		t.Errorf("Validate() error = %q, want error containing 'invalid logging level'", err.Error())
	}
}

func TestValidate_ValidLoggingLevels(t *testing.T) {
	levels := []string{"debug", "info", "warn", "error"}

	for _, level := range levels {
		t.Run(level, func(t *testing.T) {
			cfg := Default()
			cfg.Logging.Level = level

			err := cfg.Validate()
			if err != nil {
				t.Errorf("Validate() error = %v for valid level %q", err, level)
			}
		})
	}
}

func TestValidate_InvalidLoggingFormat(t *testing.T) {
	cfg := Default()
	cfg.Logging.Format = "xml"

	err := cfg.Validate()
	if err == nil {
		t.Error("Validate() expected error for invalid logging format, got nil")
	}
	if err != nil && !contains(err.Error(), "invalid logging format") {
		t.Errorf("Validate() error = %q, want error containing 'invalid logging format'", err.Error())
	}
}

func TestValidate_ValidLoggingFormats(t *testing.T) {
	formats := []string{"json", "text"}

	for _, format := range formats {
		t.Run(format, func(t *testing.T) {
			cfg := Default()
			cfg.Logging.Format = format

			err := cfg.Validate()
			if err != nil {
				t.Errorf("Validate() error = %v for valid format %q", err, format)
			}
		})
	}
}

func TestValidate_InvalidPortRange(t *testing.T) {
	tests := []struct {
		name string
		port int
	}{
		{"port zero", 0},
		{"port negative", -1},
		{"port too high", 65536},
		{"port way too high", 100000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Default()
			cfg.Server.Port = tt.port

			err := cfg.Validate()
			if err == nil {
				t.Errorf("Validate() expected error for port %d, got nil", tt.port)
			}
			if err != nil && !contains(err.Error(), "invalid server port") {
				t.Errorf("Validate() error = %q, want error containing 'invalid server port'", err.Error())
			}
		})
	}
}

func TestValidate_ValidPortRange(t *testing.T) {
	tests := []struct {
		name string
		port int
	}{
		{"port 1", 1},
		{"port 80", 80},
		{"port 443", 443},
		{"port 8080", 8080},
		{"port 65535", 65535},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Default()
			cfg.Server.Port = tt.port

			err := cfg.Validate()
			if err != nil {
				t.Errorf("Validate() error = %v for valid port %d", err, tt.port)
			}
		})
	}
}

func TestValidate_MultipleErrors(t *testing.T) {
	cfg := Default()
	cfg.Database.Driver = "invalid"
	cfg.Logging.Level = "invalid"
	cfg.Logging.Format = "invalid"
	cfg.Server.Port = 0

	err := cfg.Validate()
	if err == nil {
		t.Error("Validate() expected error for multiple invalid values, got nil")
	}

	// Verify all errors are reported
	errStr := err.Error()
	if !contains(errStr, "invalid database driver") {
		t.Error("Validate() error should contain 'invalid database driver'")
	}
	if !contains(errStr, "invalid logging level") {
		t.Error("Validate() error should contain 'invalid logging level'")
	}
	if !contains(errStr, "invalid logging format") {
		t.Error("Validate() error should contain 'invalid logging format'")
	}
	if !contains(errStr, "invalid server port") {
		t.Error("Validate() error should contain 'invalid server port'")
	}
}

func TestLoad_Precedence(t *testing.T) {
	// Test the full precedence chain: env > yaml > default

	// YAML sets some values
	yamlContent := `
instance:
  name: "yaml-name"

database:
  driver: mariadb
  url: "yaml-url"

server:
  port: 9000
`
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to write temp config file: %v", err)
	}

	// Env overrides only some YAML values
	os.Setenv("REPLICA_INSTANCE_NAME", "env-name")
	os.Setenv("REPLICA_DATABASE_DRIVER", "postgres")
	defer func() {
		os.Unsetenv("REPLICA_INSTANCE_NAME")
		os.Unsetenv("REPLICA_DATABASE_DRIVER")
	}()

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Env overrides YAML
	if cfg.Instance.Name != "env-name" {
		t.Errorf("Instance.Name = %q, want %q (env > yaml)", cfg.Instance.Name, "env-name")
	}
	if cfg.Database.Driver != "postgres" {
		t.Errorf("Database.Driver = %q, want %q (env > yaml)", cfg.Database.Driver, "postgres")
	}

	// YAML overrides default (no env set)
	if cfg.Database.URL != "yaml-url" {
		t.Errorf("Database.URL = %q, want %q (yaml > default)", cfg.Database.URL, "yaml-url")
	}
	if cfg.Server.Port != 9000 {
		t.Errorf("Server.Port = %d, want %d (yaml > default)", cfg.Server.Port, 9000)
	}

	// Default used when neither env nor YAML set
	if cfg.Server.Host != "0.0.0.0" {
		t.Errorf("Server.Host = %q, want %q (default)", cfg.Server.Host, "0.0.0.0")
	}
	if cfg.Logging.Level != "info" {
		t.Errorf("Logging.Level = %q, want %q (default)", cfg.Logging.Level, "info")
	}
}

func TestLoad_InvalidPortEnvVar(t *testing.T) {
	// Test that invalid port env var is ignored (keeps previous value)
	yamlContent := `
server:
  port: 9000
`
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to write temp config file: %v", err)
	}

	os.Setenv("REPLICA_SERVER_PORT", "not-a-number")
	defer os.Unsetenv("REPLICA_SERVER_PORT")

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Port should remain at YAML value since env var is invalid
	if cfg.Server.Port != 9000 {
		t.Errorf("Server.Port = %d, want %d (invalid env ignored)", cfg.Server.Port, 9000)
	}
}

// contains checks if substr is in s
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
