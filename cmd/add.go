package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/huh"
	"github.com/tq303/val/internal/config"
	"github.com/tq303/val/internal/discovery"
	"github.com/tq303/val/internal/installer"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add [file]",
	Short: "Add a file to be synced across packages",
	Long:  "Ad-hoc add any file (e.g. .eslintrc.js) to be tracked and synced across packages.",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		root, err := os.Getwd()
		if err != nil {
			return err
		}

		// Resolve file path
		var file string
		if len(args) == 1 {
			file = args[0]
		} else {
			if err := huh.NewForm(
				huh.NewGroup(
					huh.NewInput().
						Title("File to add").
						Placeholder(".eslintrc.js").
						Value(&file),
				),
			).Run(); err != nil {
				return err
			}
		}

		// Verify file exists
		if _, err := os.Stat(filepath.Join(root, file)); err != nil {
			return fmt.Errorf("file not found: %s", file)
		}

		// Default dest to same filename
		dest := filepath.Base(file)
		if err := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Destination path in each package").
					Value(&dest),
			),
		).Run(); err != nil {
			return err
		}

		// Discover packages
		packages, err := discovery.Discover(root)
		if err != nil {
			return err
		}

		var pkgOptions []huh.Option[string]
		for _, p := range packages {
			pkgOptions = append(pkgOptions, huh.NewOption(p.Path, p.Path))
		}

		var selectedPkgs []string
		if err := huh.NewForm(
			huh.NewGroup(
				huh.NewMultiSelect[string]().
					Title("Which packages should this apply to?").
					Options(pkgOptions...).
					Value(&selectedPkgs),
			),
		).Run(); err != nil {
			return err
		}

		if len(selectedPkgs) == 0 {
			fmt.Println("No packages selected, nothing to do.")
			return nil
		}

		// Update valet.yaml
		cfg, err := config.Load(root)
		if err != nil {
			return err
		}

		rule := config.Rule{File: file, Dest: dest}
		cfg.Rules = append(cfg.Rules, rule)

		// Ensure all selected packages are in config
		existing := map[string]bool{}
		for _, p := range cfg.Packages {
			existing[p.Path] = true
		}
		for _, p := range selectedPkgs {
			if !existing[p] {
				cfg.Packages = append(cfg.Packages, config.Package{Path: p})
			}
		}

		if err := config.Save(root, cfg); err != nil {
			return err
		}

		// Install immediately
		results, err := installer.InstallRule(root, rule, selectedPkgs, false)
		if err != nil {
			return err
		}

		fmt.Printf("\nInstalled %s:\n\n", file)
		for _, r := range results {
			fmt.Printf("  %s → %s\n", r.Package, r.Dest)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
}
