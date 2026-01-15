package main

import (
	"fmt"
	"os"

	"github.com/deviantintegral/replica/internal/config"
	"github.com/deviantintegral/replica/internal/logging"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var (
	cfgFile string
	cfg     *config.Config
)

var rootCmd = &cobra.Command{
	Use:   "replica",
	Short: "Replica - A distributed CMS",
	Long: `Replica is a distributed content management system supporting
multiple database backends and designed for offline-first workflows.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Load configuration
		var err error
		cfg, err = config.Load(cfgFile)
		if err != nil {
			// If config file was explicitly specified and doesn't exist, error
			if cfgFile != "" {
				return fmt.Errorf("loading config: %w", err)
			}
			// Otherwise use defaults
			cfg = config.Default()
		}

		// Initialize logger
		logger := logging.New(cfg.Logging)

		// Set global logger
		log.Logger = logger
		zerolog.DefaultContextLogger = &logger

		return nil
	},
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "replica.yaml", "config file path")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
