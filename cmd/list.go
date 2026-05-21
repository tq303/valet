package cmd

import (
	"fmt"
	"os"

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
			fmt.Println("No files configured — run `val init` or `val add` first.")
			return nil
		}

		pkgPaths := make([]string, len(cfg.Packages))
		for i, p := range cfg.Packages {
			pkgPaths[i] = p.Path
		}

		fmt.Printf("%-35s %-20s %s\n", "FILE", "PACKAGE", "STATUS")
		fmt.Printf("%-35s %-20s %s\n", "----", "-------", "------")

		exitCode := 0
		for _, rule := range cfg.Rules {
			results, err := installer.InstallFile(root, rule, pkgPaths, true)
			if err != nil {
				return err
			}
			for _, r := range results {
				if r.Skipped {
					fmt.Printf("%-35s %-20s excluded\n", rule.File, r.Package)
					continue
				}
				if _, err := os.Stat(r.Dest); os.IsNotExist(err) {
					fmt.Printf("%-35s %-20s MISSING\n", rule.File, r.Package)
					exitCode = 1
				} else {
					fmt.Printf("%-35s %-20s ok\n", rule.File, r.Package)
				}
			}
		}

		if exitCode != 0 {
			fmt.Println("\nRun `val install` to fix missing files.")
			os.Exit(exitCode)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
