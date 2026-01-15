---
id: 7
group: "configuration"
dependencies: [2]
status: "pending"
created: "2026-01-14"
skills:
  - go
  - yaml
---
# Configuration System

## Objective
Implement YAML configuration loading with environment variable overrides, validation, and default values for all Replica settings.

## Skills Required
- **go**: Go structs, struct tags, error handling
- **yaml**: YAML parsing with gopkg.in/yaml.v3

## Acceptance Criteria
- [ ] Configuration struct definitions cover instance, database, server, logging
- [ ] YAML file loads from `./replica.yaml` (configurable path)
- [ ] Environment variables override YAML values (e.g., `REPLICA_DATABASE_DRIVER`)
- [ ] Invalid configuration produces clear error messages
- [ ] Default configuration values work out of the box
- [ ] `replica.yaml.example` template file created

Use your internal Todo tool to track these and keep on track.

## Technical Requirements
- Use `gopkg.in/yaml.v3` for YAML parsing
- Environment variable prefix: `REPLICA_`
- Nested keys use underscore: `REPLICA_DATABASE_DRIVER`
- Config file path default: `./replica.yaml`

## Input Dependencies
- Task 2: Cobra CLI (config loads during command initialization)

## Output Artifacts
- `internal/config/config.go` - Configuration structs and loader
- `internal/config/config_test.go` - Unit tests
- `replica.yaml.example` - Example configuration file

## Implementation Notes

<details>
<summary>Detailed Implementation Instructions</summary>

1. **Add YAML dependency**:
   ```bash
   go get gopkg.in/yaml.v3
   ```

2. **Create config structs** (`internal/config/config.go`):
   ```go
   package config

   import (
       "fmt"
       "os"
       "strings"

       "gopkg.in/yaml.v3"
   )

   // Config holds all configuration for Replica
   type Config struct {
       Instance InstanceConfig `yaml:"instance"`
       Database DatabaseConfig `yaml:"database"`
       Server   ServerConfig   `yaml:"server"`
       Logging  LoggingConfig  `yaml:"logging"`
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

   // Default returns a Config with default values
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

   // Load loads configuration from file and environment
   func Load(path string) (*Config, error) {
       cfg := Default()

       // Try to load from file
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

       // Validate
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

   // Validate checks configuration for errors
   func (c *Config) Validate() error {
       var errs []string

       // Validate database driver
       validDrivers := map[string]bool{"sqlite": true, "mariadb": true, "postgres": true}
       if !validDrivers[c.Database.Driver] {
           errs = append(errs, fmt.Sprintf("invalid database driver: %s (must be sqlite, mariadb, or postgres)", c.Database.Driver))
       }

       // Validate database URL
       if c.Database.URL == "" {
           errs = append(errs, "database URL is required")
       }

       // Validate server port
       if c.Server.Port < 1 || c.Server.Port > 65535 {
           errs = append(errs, fmt.Sprintf("invalid server port: %d (must be 1-65535)", c.Server.Port))
       }

       // Validate logging level
       validLevels := map[string]bool{"debug": true, "info": true, "warn": true, "error": true}
       if !validLevels[c.Logging.Level] {
           errs = append(errs, fmt.Sprintf("invalid logging level: %s (must be debug, info, warn, or error)", c.Logging.Level))
       }

       // Validate logging format
       validFormats := map[string]bool{"json": true, "text": true}
       if !validFormats[c.Logging.Format] {
           errs = append(errs, fmt.Sprintf("invalid logging format: %s (must be json or text)", c.Logging.Format))
       }

       if len(errs) > 0 {
           return fmt.Errorf("%s", strings.Join(errs, "; "))
       }
       return nil
   }
   ```

3. **Create example config** (`replica.yaml.example`):
   ```yaml
   # Replica Configuration File
   # Copy this file to replica.yaml and modify as needed

   instance:
     name: "production-us-east"

   database:
     driver: sqlite  # sqlite, mariadb, postgres
     url: "replica.db"
     # For MariaDB: "replica:replica_pass@tcp(localhost:3306)/replica?parseTime=true"
     # For PostgreSQL: "postgres://replica:replica_pass@localhost:5432/replica?sslmode=disable"

   server:
     host: "0.0.0.0"
     port: 8080

   logging:
     level: info  # debug, info, warn, error
     format: json # json, text
   ```

4. **Write tests** (`internal/config/config_test.go`):
   Test cases:
   - Default values load correctly
   - YAML file parses correctly
   - Environment variables override YAML
   - Invalid driver produces error
   - Invalid port produces error
   - Missing URL produces error

</details>
