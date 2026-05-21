package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/tq303/val/internal/config"
	"github.com/tq303/val/internal/installer"
)

var dryRun bool

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Apply all configured files into each package",
	Long:  "Reads valet.yaml and installs every configured file into its target packages.",
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
			fmt.Println("No rules configured — run `val init` or `val add` first.")
			return nil
		}

		results, err := installer.InstallAll(root, cfg, dryRun)
		if err != nil {
			return err
		}

		if dryRun {
			fmt.Println("Dry run — no files written:\n")
		} else {
			fmt.Println("Installed:\n")
		}
		for _, r := range results {
			fmt.Printf("  %s → %s\n", r.File, r.Dest)
		}
		return nil
	},
}

func init() {
	installCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview changes without applying them")
	rootCmd.AddCommand(installCmd)
}
