package cmd

import (
	"os"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "val",
	Short: "Valet — sync config files across your monorepo",
	Long:  "Valet manages and syncs config files across your monorepo packages. Add any file once, install it everywhere.",
}

func Execute() {
	rootCmd.CompletionOptions.HiddenDefaultCmd = true
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
