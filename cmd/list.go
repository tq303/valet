package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tq303/valet/internal/config"
	"github.com/tq303/valet/internal/installer"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Show configured files and coverage per package",
	Long:  "Lists all files configured in valet.yaml and their coverage across packages.",
	RunE: func(cmd *cobra.Command, args []string) error {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		root, err := config.FindRoot(cwd)
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

		fmt.Printf("%-30s %-20s %-20s %s\n", "FILE", "LOCATION", "PLATFORM", "STATUS")
		fmt.Printf("%-30s %-20s %-20s %s\n", "----", "--------", "--------", "------")

		exitCode := 0
		for _, rule := range cfg.Rules {
			if !installer.MatchesPlatform(rule, listPlatform) {
				continue
			}
			platform := "-"
			if len(rule.Platforms) > 0 {
				platform = strings.Join(rule.Platforms, ",")
			}
			results, err := installer.SyncRule(root, rule, rule.Locations, true, false)
			if err != nil {
				return err
			}
			for _, r := range results {
				if _, err := os.Stat(r.Dest); os.IsNotExist(err) {
					fmt.Printf("%-30s %-20s %-20s MISSING\n", installer.FileName(r.File), r.Package, platform)
					exitCode = 1
				} else {
					fmt.Printf("%-30s %-20s %-20s ok\n", installer.FileName(r.File), r.Package, platform)
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

var listPlatform string

func init() {
	listCmd.Flags().StringVar(&listPlatform, "platform", "", "Only show rules matching this platform tag")
	rootCmd.AddCommand(listCmd)
}
