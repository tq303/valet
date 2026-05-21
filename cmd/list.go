package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/tq303/val/internal/config"
	"github.com/tq303/val/internal/installer"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Show configured files and coverage per package",
	Long:  "Lists all files configured in valet.yaml and their coverage across packages.",
	RunE: func(cmd *cobra.Command, args []string) error {
		root, err := os.Getwd()
		if err != nil {
			return err
		}

		cfg, err := config.Load(root)
		if err != nil {
			return err
		}
		if len(cfg.Rules) == 0 {
			fmt.Println("No files configured — run `valet add <file>` to get started.")
			return nil
		}

		fmt.Printf("%-30s %-20s %s\n", "FILE", "PACKAGE", "STATUS")
		fmt.Printf("%-30s %-20s %s\n", "----", "-------", "------")

		exitCode := 0
		for _, rule := range cfg.Rules {
			results, err := installer.SyncRule(root, rule, rule.Locations, true)
			if err != nil {
				return err
			}
			for _, r := range results {
				if _, err := os.Stat(r.Dest); os.IsNotExist(err) {
					fmt.Printf("%-30s %-20s MISSING\n", filepath.Base(r.File), r.Package)
					exitCode = 1
				} else {
					fmt.Printf("%-30s %-20s ok\n", filepath.Base(r.File), r.Package)
				}
			}
		}

		if exitCode != 0 {
			fmt.Println("\nRun `valet sync` to fix missing files.")
			os.Exit(exitCode)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
