---
id: 7
group: "configuration"
dependencies: [2]
status: "completed"
created: "2026-01-15"
skills:
  - go
---
# Configuration System

## Objective
Implement YAML configuration loading with environment variable overrides. This provides flexible configuration management for different deployment environments.

## Skills Required
- go: Go programming with YAML parsing and validation

## Acceptance Criteria
- [ ] YAML configuration loads from `./replica.yaml` by default
- [ ] Environment variables override YAML values (e.g., `REPLICA_DATABASE_DRIVER`)
- [ ] Invalid configuration produces clear error messages
- [ ] Default configuration works out of the box
- [ ] Configuration struct covers all required categories
- [ ] Example configuration file `replica.yaml.example` exists

Use your internal Todo tool to track these and keep on track.

## Technical Requirements
- YAML parsing: `gopkg.in/yaml.v3`
- Environment variable overrides using struct tags
- Validation for required fields and valid values
- Sensible defaults for all optional fields

## Input Dependencies
- Task 3: Cobra CLI foundation for config flag integration

## Output Artifacts
- `internal/config/config.go` - Configuration struct and loader
- `internal/config/config_test.go` - Unit tests
- `replica.yaml.example` - Example configuration file

## Implementation Notes

<details>
<summary>Detailed Implementation Guide</summary>

### Configuration Struct

Create `internal/config/config.go`:

```go
package config

import (
    "fmt"
    "os"
    "strings"

    "gopkg.in/yaml.v3"
)

type Config struct {
    Instance InstanceConfig  `yaml:"instance"`
    Database DatabaseConfig  `yaml:"database"`
    Server   ServerConfig    `yaml:"server"`
    Logging  LoggingConfig   `yaml:"logging"`
}

type InstanceConfig struct {
    Name string `yaml:"name"`
}

type DatabaseConfig struct {
    Driver string `yaml:"driver"` // sqlite, mariadb, postgres
    URL    string `yaml:"url"`
}

type ServerConfig struct {
    Host string `yaml:"host"`
    Port int    `yaml:"port"`
}

type LoggingConfig struct {
    Level  string `yaml:"level"`  // debug, info, warn, error
    Format string `yaml:"format"` // json, text
}

// Default returns configuration with sensible defaults
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

// Load reads configuration from file and applies environment overrides
func Load(path string) (*Config, error) {
    cfg := Default()

    // Load from file if it exists
    if path != "" {
        data, err := os.ReadFile(path)
        if err != nil {
            if !os.IsNotExist(err) {
                return nil, fmt.Errorf("reading config file: %w", err)
            }
            // File doesn't exist, use defaults
        } else {
            if err := yaml.Unmarshal(data, cfg); err != nil {
                return nil, fmt.Errorf("parsing config file: %w", err)
            }
        }
    }

    // Apply environment variable overrides
    cfg.applyEnvOverrides()

    // Validate configuration
    if err := cfg.Validate(); err != nil {
        return nil, fmt.Errorf("invalid configuration: %w", err)
    }

    return cfg, nil
}

func (c *Config) applyEnvOverrides() {
    if v := os.Getenv("REPLICA_INSTANCE_NAME"); v != "" {
        c.Instance.Name = v
    }
    if v := os.Getenv("REPLICA_DATABASE_DRIVER"); v != "" {
        c.Database.Driver = v
    }
    if v := os.Getenv("REPLICA_DATABASE_URL"); v != "" {
        c.Database.URL = v
    }
    if v := os.Getenv("REPLICA_SERVER_HOST"); v != "" {
        c.Server.Host = v
    }
    if v := os.Getenv("REPLICA_SERVER_PORT"); v != "" {
        // Parse port from string
        var port int
        fmt.Sscanf(v, "%d", &port)
        if port > 0 {
            c.Server.Port = port
        }
    }
    if v := os.Getenv("REPLICA_LOGGING_LEVEL"); v != "" {
        c.Logging.Level = v
    }
    if v := os.Getenv("REPLICA_LOGGING_FORMAT"); v != "" {
        c.Logging.Format = v
    }
}

func (c *Config) Validate() error {
    // Validate database driver
    validDrivers := map[string]bool{
        "sqlite": true, "mariadb": true, "postgres": true,
    }
    if !validDrivers[strings.ToLower(c.Database.Driver)] {
        return fmt.Errorf("invalid database driver %q: must be sqlite, mariadb, or postgres", c.Database.Driver)
    }

    // Validate logging level
    validLevels := map[string]bool{
        "debug": true, "info": true, "warn": true, "error": true,
    }
    if !validLevels[strings.ToLower(c.Logging.Level)] {
        return fmt.Errorf("invalid logging level %q: must be debug, info, warn, or error", c.Logging.Level)
    }

    // Validate logging format
    validFormats := map[string]bool{"json": true, "text": true}
    if !validFormats[strings.ToLower(c.Logging.Format)] {
        return fmt.Errorf("invalid logging format %q: must be json or text", c.Logging.Format)
    }

    // Validate port range
    if c.Server.Port < 1 || c.Server.Port > 65535 {
        return fmt.Errorf("invalid port %d: must be between 1 and 65535", c.Server.Port)
    }

    return nil
}
```

### Example Configuration

Create `replica.yaml.example`:

```yaml
# Replica Configuration
# Copy to replica.yaml and modify as needed

instance:
  name: "production-us-east"

database:
  # Supported drivers: sqlite, mariadb, postgres
  driver: sqlite
  # For SQLite: path to database file
  # For MariaDB: user:password@tcp(host:port)/database
  # For PostgreSQL: postgres://user:password@host:port/database?sslmode=disable
  url: "replica.db"

server:
  host: "0.0.0.0"
  port: 8080

logging:
  # Levels: debug, info, warn, error
  level: info
  # Formats: json, text
  format: json
```

### Unit Tests

Create `internal/config/config_test.go` with tests for:
- Default configuration values
- YAML loading
- Environment variable overrides
- Validation errors for invalid values
- Override precedence (env > yaml > default)

### Verification

```bash
go test ./internal/config/...
```

</details>
