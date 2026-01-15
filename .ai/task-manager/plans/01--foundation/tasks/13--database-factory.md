---
id: 13
group: "database"
dependencies: [12]
status: "pending"
created: "2026-01-14"
skills:
  - go
  - database
---
# Database Factory

## Objective
Create a factory function that instantiates the correct database driver based on configuration, providing a unified entry point for database initialization.

## Skills Required
- **go**: Factory pattern, interface types
- **database**: Multi-driver abstraction

## Acceptance Criteria
- [ ] Factory function creates correct driver from config
- [ ] Factory validates driver selection
- [ ] Returns error for unknown drivers
- [ ] Integrates with CLI for database initialization

Use your internal Todo tool to track these and keep on track.

## Technical Requirements
- Factory function in `internal/database/factory.go`
- CLI integration for startup database connection

## Input Dependencies
- Task 12: All drivers with migration support

## Output Artifacts
- `internal/database/factory.go` - Factory function
- Updated CLI to initialize database on startup

## Implementation Notes

<details>
<summary>Detailed Implementation Instructions</summary>

1. **Create factory function** (`internal/database/factory.go`):
   ```go
   package database

   import (
       "fmt"
       "time"

       "github.com/deviantintegral/replica/internal/config"
       "github.com/deviantintegral/replica/internal/database/mysql"
       "github.com/deviantintegral/replica/internal/database/postgres"
       "github.com/deviantintegral/replica/internal/database/sqlite"
   )

   // New creates a new database connection based on configuration
   func New(cfg *config.DatabaseConfig) (Database, error) {
       dbCfg := &Config{
           Driver:          cfg.Driver,
           URL:             cfg.URL,
           MaxOpenConns:    25,
           MaxIdleConns:    5,
           ConnMaxLifetime: 5 * time.Minute,
       }

       switch cfg.Driver {
       case "sqlite":
           return sqlite.New(dbCfg)
       case "mariadb":
           return mysql.New(dbCfg)
       case "postgres":
           return postgres.New(dbCfg)
       default:
           return nil, fmt.Errorf("unknown database driver: %s", cfg.Driver)
       }
   }
   ```

2. **Update root command** (`cmd/replica/root.go`):
   ```go
   var (
       cfgFile string
       cfg     *config.Config
       logger  zerolog.Logger
       db      database.Database
   )

   func init() {
       rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "replica.yaml", "config file path")
       rootCmd.PersistentPreRunE = initApp
   }

   func initApp(cmd *cobra.Command, args []string) error {
       // Skip init for version command
       if cmd.Name() == "version" {
           return nil
       }

       var err error
       cfg, err = config.Load(cfgFile)
       if err != nil {
           return fmt.Errorf("loading config: %w", err)
       }

       logger = logging.Setup(&cfg.Logging)
       logger.Info().Str("config", cfgFile).Msg("configuration loaded")

       db, err = database.New(&cfg.Database)
       if err != nil {
           return fmt.Errorf("connecting to database: %w", err)
       }

       if err := db.Migrate(); err != nil {
           return fmt.Errorf("running migrations: %w", err)
       }

       logger.Info().
           Str("driver", cfg.Database.Driver).
           Msg("database connected and migrated")

       return nil
   }
   ```

3. **Add cleanup** (`cmd/replica/root.go`):
   ```go
   func Execute() {
       defer func() {
           if db != nil {
               db.Close()
           }
       }()
       if err := rootCmd.Execute(); err != nil {
           os.Exit(1)
       }
   }
   ```

4. **Write factory tests** (`internal/database/factory_test.go`):
   Test cases:
   - SQLite driver created correctly
   - Unknown driver returns error
   - Factory passes config correctly

</details>
