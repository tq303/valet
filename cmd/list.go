package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Show configured files and coverage per package",
	Long:  "Lists all files configured in valet.yaml and their coverage across packages.",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("val list — not yet implemented")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
