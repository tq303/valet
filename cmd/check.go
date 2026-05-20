package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Health check AI rule coverage across all packages",
	Long:  "Reports rule coverage per package and flags missing or conflicting rules.",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("val check — not yet implemented")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(checkCmd)
}
