package main

import (
	"fmt"

	"github.com/deviantintegral/replica/internal/version"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		info := version.Get()
		fmt.Printf("replica version %s\n", info.Version)
		fmt.Printf("  commit: %s\n", info.Commit)
		fmt.Printf("  built:  %s\n", info.BuildDate)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
