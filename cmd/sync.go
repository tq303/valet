package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/tq303/val/internal/config"
	"github.com/tq303/val/internal/installer"
)

var dryRun bool

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync all configured files into their packages",
	Long:  "Reads valet.yaml and syncs every configured file into its target packages.",
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

		results, err := installer.SyncAll(root, cfg, dryRun)
		if err != nil {
			return err
		}

		if dryRun {
			fmt.Println("Dry run — no files written:")
		} else {
			fmt.Println("Synced:")
		}
		fmt.Println()
		for _, r := range results {
			fmt.Printf("  %s → %s\n", filepath.Base(r.File), r.Dest)
		}
		return nil
	},
}

func init() {
	syncCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview changes without applying them")
	rootCmd.AddCommand(syncCmd)
}
