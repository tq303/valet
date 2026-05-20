package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "val",
	Short: "Valiant — AI rule coverage for monorepos",
	Long:  "Valiant discovers your monorepo structure, validates AI rule coverage across packages, and installs rules into each package automatically.",
}

func Execute() {
	rootCmd.CompletionOptions.HiddenDefaultCmd = true
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
