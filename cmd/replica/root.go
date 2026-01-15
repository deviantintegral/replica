package main

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "replica",
	Short: "Replica - A distributed CMS",
	Long: `Replica is a distributed content management system supporting
multiple database backends and designed for offline-first workflows.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
