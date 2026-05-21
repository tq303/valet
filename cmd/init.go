package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/tq303/val/internal/config"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialise Valet in the current directory",
	Long:  "Creates a valet.yaml config file in the current directory.",
	RunE: func(cmd *cobra.Command, args []string) error {
		root, err := os.Getwd()
		if err != nil {
			return err
		}

		cfg := config.Config{Version: 1}
		if err := config.Save(root, &cfg); err != nil {
			return fmt.Errorf("could not write valet.yaml: %w", err)
		}

		fmt.Println("Created valet.yaml — use `valet add <file>` to start syncing files.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
