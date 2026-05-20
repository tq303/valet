package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/tq303/val/internal/discovery"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialise Valet in a monorepo",
	Long:  "Detects monorepo structure and creates a valiant.yaml config file.",
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
			tools := "no AI tools detected"
			if len(p.Tools) > 0 {
				tools = ""
				for i, t := range p.Tools {
					if i > 0 {
						tools += ", "
					}
					tools += string(t)
				}
			}
			fmt.Printf("  %-30s %s\n", p.Path, tools)
		}

		rulesDir := filepath.Join(root, ".rules")
		if _, err := os.Stat(rulesDir); os.IsNotExist(err) {
			if err := os.Mkdir(rulesDir, 0755); err != nil {
				return fmt.Errorf("could not create .rules folder: %w", err)
			}
			fmt.Println("\nCreated .rules/ — add your rule files there:")
			fmt.Println("  CLAUDE.md       → Claude Code")
			fmt.Println("  rules.mdc       → Cursor")
		} else {
			fmt.Println("\n.rules/ already exists, skipping.")
		}

		fmt.Println("\nvalet.yaml generation coming soon.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
