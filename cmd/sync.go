package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/tq303/valet/internal/config"
	"github.com/tq303/valet/internal/installer"
)

var dryRun bool

var syncCmd = &cobra.Command{
	Use:   "sync [file]",
	Short: "Sync all configured files into their locations",
	Long:  "Reads valet.yaml and syncs every configured file into its target locations. Pass a file path to promote that version as the source before syncing.",
	Args:  cobra.MaximumNArgs(1),
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

		if len(args) == 1 {
			return syncFromPath(root, cfg, args[0], dryRun)
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
			fmt.Printf("  %s → %s\n", installer.FileName(r.File), r.Dest)
		}
		return nil
	},
}

func syncFromPath(root string, cfg *config.Config, path string, dryRun bool) error {
	abs := installer.ResolvePath(root, path)
	filename := filepath.Base(abs)
	dir := filepath.Dir(abs)

	for _, rule := range cfg.Rules {
		for _, f := range rule.Files {
			if filepath.Base(f) != filename {
				continue
			}
			for _, loc := range rule.Locations {
				locAbs := installer.ResolvePath(root, loc)
				expectedDir := locAbs
				if rule.Dest != "" {
					expectedDir = filepath.Join(locAbs, rule.Dest)
				}
				if dir != expectedDir {
					continue
				}
				// Promote this version to the source
				src := installer.ResolvePath(root, f)
				if !dryRun {
					if err := installer.CopyFile(abs, src); err != nil {
						return fmt.Errorf("failed to promote %s: %w", filename, err)
					}
				}
				results, err := installer.SyncRule(root, rule, rule.Locations, dryRun)
				if err != nil {
					return err
				}
				if dryRun {
					fmt.Printf("Dry run — would promote %s from %s and sync to:\n\n", filename, loc)
				} else {
					fmt.Printf("Promoted %s from %s and synced to:\n\n", filename, loc)
				}
				for _, r := range results {
					fmt.Printf("  %s → %s\n", r.Package, r.Dest)
				}
				return nil
			}
		}
	}

	return fmt.Errorf("%s is not inside a tracked location", path)
}

func init() {
	syncCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview changes without applying them")
	rootCmd.AddCommand(syncCmd)
}
