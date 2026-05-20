package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var dryRun bool

var installCmd = &cobra.Command{
	Use:   "install [package]",
	Short: "Install AI rules into each package",
	Long:  "Generates tool-specific config files (.claude/, .cursor/) in each package directory.",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("val install — not yet implemented")
		return nil
	},
}

func init() {
	installCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview changes without applying them")
	rootCmd.AddCommand(installCmd)
}
