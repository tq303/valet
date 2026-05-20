package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialise Valiant in a monorepo",
	Long:  "Detects monorepo structure and creates a valiant.yaml config file.",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("val init — not yet implemented")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
