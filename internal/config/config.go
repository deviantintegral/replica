// Package config provides configuration loading and validation for the replica application.
package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config represents the complete application configuration.
type Config struct {
	Instance InstanceConfig `yaml:"instance"`
	Database DatabaseConfig `yaml:"database"`
	Server   ServerConfig   `yaml:"server"`
	Logging  LoggingConfig  `yaml:"logging"`
}

// InstanceConfig contains instance identification settings.
type InstanceConfig struct {
	Name string `yaml:"name"`
}

// DatabaseConfig contains database connection settings.
type DatabaseConfig struct {
	Driver string `yaml:"driver"`
	URL    string `yaml:"url"`
}

// ServerConfig contains HTTP server settings.
type ServerConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

// LoggingConfig contains logging settings.
type LoggingConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
}

// validDatabaseDrivers is the set of supported database drivers.
var validDatabaseDrivers = map[string]bool{
	"sqlite":  true,
	"mariadb": true,
	"postgres": true,
}

// validLogLevels is the set of supported log levels.
var validLogLevels = map[string]bool{
	"debug": true,
	"info":  true,
	"warn":  true,
	"error": true,
}

// validLogFormats is the set of supported log formats.
var validLogFormats = map[string]bool{
	"json": true,
	"text": true,
}

// Default returns a Config with sensible default values.
func Default() *Config {
	return &Config{
		Instance: InstanceConfig{
			Name: "replica",
		},
		Database: DatabaseConfig{
			Driver: "sqlite",
			URL:    "replica.db",
		},
		Server: ServerConfig{
			Host: "0.0.0.0",
			Port: 8080,
		},
		Logging: LoggingConfig{
			Level:  "info",
			Format: "json",
		},
	}
}

// Load loads configuration from a YAML file, applies environment variable
// overrides, and validates the result. The precedence order is:
// environment variables > YAML file > defaults.
func Load(path string) (*Config, error) {
	cfg := Default()

	// Load YAML file if it exists
	if path != "" {
		data, err := os.ReadFile(path)
		if err != nil {
			if !os.IsNotExist(err) {
				return nil, fmt.Errorf("reading config file: %w", err)
			}
			// File doesn't exist, continue with defaults
		} else {
			if err := yaml.Unmarshal(data, cfg); err != nil {
				return nil, fmt.Errorf("parsing config file: %w", err)
			}
		}
	}

	// Apply environment variable overrides
	applyEnvOverrides(cfg)

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("validating config: %w", err)
	}

	return cfg, nil
}

// applyEnvOverrides applies environment variable overrides to the configuration.
// Environment variables use the REPLICA_ prefix and underscores to separate
// nested keys (e.g., REPLICA_DATABASE_DRIVER).
func applyEnvOverrides(cfg *Config) {
	if v := os.Getenv("REPLICA_INSTANCE_NAME"); v != "" {
		cfg.Instance.Name = v
	}

	if v := os.Getenv("REPLICA_DATABASE_DRIVER"); v != "" {
		cfg.Database.Driver = v
	}

	if v := os.Getenv("REPLICA_DATABASE_URL"); v != "" {
		cfg.Database.URL = v
	}

	if v := os.Getenv("REPLICA_SERVER_HOST"); v != "" {
		cfg.Server.Host = v
	}

	if v := os.Getenv("REPLICA_SERVER_PORT"); v != "" {
		if port, err := strconv.Atoi(v); err == nil {
			cfg.Server.Port = port
		}
	}

	if v := os.Getenv("REPLICA_LOGGING_LEVEL"); v != "" {
		cfg.Logging.Level = v
	}

	if v := os.Getenv("REPLICA_LOGGING_FORMAT"); v != "" {
		cfg.Logging.Format = v
	}
}

// Validate checks that all configuration values are valid.
func (c *Config) Validate() error {
	var errs []string

	// Validate database driver
	driver := strings.ToLower(c.Database.Driver)
	if !validDatabaseDrivers[driver] {
		errs = append(errs, fmt.Sprintf("invalid database driver %q: must be one of sqlite, mariadb, postgres", c.Database.Driver))
	}

	// Validate logging level
	level := strings.ToLower(c.Logging.Level)
	if !validLogLevels[level] {
		errs = append(errs, fmt.Sprintf("invalid logging level %q: must be one of debug, info, warn, error", c.Logging.Level))
	}

	// Validate logging format
	format := strings.ToLower(c.Logging.Format)
	if !validLogFormats[format] {
		errs = append(errs, fmt.Sprintf("invalid logging format %q: must be one of json, text", c.Logging.Format))
	}

	// Validate server port
	if c.Server.Port < 1 || c.Server.Port > 65535 {
		errs = append(errs, fmt.Sprintf("invalid server port %d: must be between 1 and 65535", c.Server.Port))
	}

	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "; "))
	}

	return nil
}
