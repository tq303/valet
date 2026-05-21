package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/tq303/val/internal/config"
	"github.com/tq303/val/internal/discovery"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialise Valet in a monorepo",
	Long:  "Detects monorepo structure and creates a valet.yaml config file.",
	RunE: func(cmd *cobra.Command, args []string) error {
		root, err := os.Getwd()
		if err != nil {
			return err
		}

		fmt.Println("Scanning monorepo...")
		packages, err := discovery.Discover(root)
		if err != nil {
			return err
		}

		fmt.Printf("Found %d package(s):\n\n", len(packages))
		for _, p := range packages {
			fmt.Printf("  %s\n", p.Path)
		}

		cfg := config.Config{Version: 1}
		if err := config.Save(root, &cfg); err != nil {
			return fmt.Errorf("could not write valet.yaml: %w", err)
		}

		fmt.Println("\nCreated valet.yaml — use `valet add <file>` to start syncing files.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
