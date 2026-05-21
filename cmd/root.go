package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "valet",
	Short: "Valet — sync config files across your monorepo",
	Long:  "Valet manages and syncs files across locations. Add any file, URL or repo location once or cache to local file.",
}

func Execute() {
	rootCmd.CompletionOptions.HiddenDefaultCmd = true
	rootCmd.SetHelpCommand(&cobra.Command{Hidden: true})
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
