package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
	"github.com/tq303/val/internal/config"
	"github.com/tq303/val/internal/discovery"
	"github.com/tq303/val/internal/installer"
)

var addCmd = &cobra.Command{
	Use:   "add [file]",
	Short: "Add a file to be synced across packages",
	Long:  "Add any file to be tracked and synced. If a matching rule exists, the file is added to it.",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		root, err := os.Getwd()
		if err != nil {
			return err
		}

		var file string
		if len(args) == 1 {
			file = args[0]
		} else {
			fmt.Println("Tip: use `valet add <file>` for tab completion.")
			if err := huh.NewForm(
				huh.NewGroup(
					huh.NewInput().
						Title("File to add").
						Placeholder("CLAUDE.md").
						Value(&file),
				),
			).Run(); err != nil {
				return err
			}
		}

		// Verify file exists
		src := file
		if !filepath.IsAbs(src) {
			src = filepath.Join(root, src)
		}
		if _, err := os.Stat(src); err != nil {
			return fmt.Errorf("file not found: %s", file)
		}

		filename := filepath.Base(file)

		cfg, err := config.Load(root)
		if err != nil {
			return err
		}

		// Check if file is already tracked in any rule
		for i, rule := range cfg.Rules {
			for _, f := range rule.Files {
				if filepath.Base(f) == filename {
					// File already tracked — show packages that don't have it yet
					existing := map[string]bool{}
					for _, rp := range rule.Locations {
						existing[rp.Path] = true
					}

					allPkgs, err := discovery.Discover(root)
					if err != nil {
						return err
					}

					var locOptions []huh.Option[string]
					for _, p := range allPkgs {
						if !existing[p.Path] {
							locOptions = append(locOptions, huh.NewOption(p.Path, p.Path))
						}
					}

					if len(locOptions) == 0 {
						fmt.Printf("%s is already tracked for all packages.\n", filename)
						return nil
					}

					var selectedLocs []string
					if err := huh.NewForm(
						huh.NewGroup(
							huh.NewMultiSelect[string]().
								Title(fmt.Sprintf("Add %s to which packages?", filename)).
								Options(locOptions...).
								Value(&selectedLocs),
						),
					).Run(); err != nil {
						return err
					}

					if len(selectedLocs) == 0 {
						fmt.Println("No packages selected, nothing to do.")
						return nil
					}

					for _, p := range selectedLocs {
						cfg.Rules[i].Locations = append(cfg.Rules[i].Locations, config.Location{Path: p})
					}
					if err := config.Save(root, cfg); err != nil {
						return err
					}
					results, err := installer.SyncRule(root, cfg.Rules[i], selectedLocs, false)
					if err != nil {
						return err
					}
					fmt.Printf("\nAdded %s to:\n\n", filename)
					for _, r := range results {
						fmt.Printf("  %s → %s\n", r.Package, r.Dest)
					}
					return nil
				}
			}
		}

		// New file — prompt for dest folder and packages
		packages, err := discovery.Discover(root)
		if err != nil {
			return err
		}

		var locOptions []huh.Option[string]
		for _, p := range packages {
			locOptions = append(locOptions, huh.NewOption(p.Path, p.Path))
		}

		var selectedLocs []string
		var dest string

		if err := huh.NewForm(
			huh.NewGroup(
				huh.NewMultiSelect[string]().
					Title("Which packages should this apply to?").
					Options(locOptions...).
					Value(&selectedLocs),
				huh.NewInput().
					Title("Destination folder in each package (leave blank for root)").
					Placeholder(".claude").
					Value(&dest),
			),
		).Run(); err != nil {
			return err
		}

		if len(selectedLocs) == 0 {
			fmt.Println("No packages selected, nothing to do.")
			return nil
		}

		var locs []config.Location
		for _, p := range selectedLocs {
			locs = append(locs, config.Location{Path: p})
		}

		// Find an existing rule with the same dest and packages to group into
		for i, rule := range cfg.Rules {
			if rule.Dest == dest && samePackages(rule.Locations, locs) {
				cfg.Rules[i].Files = append(cfg.Rules[i].Files, filename)
				if err := config.Save(root, cfg); err != nil {
					return err
				}
				results, err := installer.SyncRule(root, cfg.Rules[i], selectedLocs, false)
				if err != nil {
					return err
				}
				fmt.Printf("\nAdded %s:\n\n", filename)
				for _, r := range results {
					fmt.Printf("  %s → %s\n", r.Package, r.Dest)
				}
				return nil
			}
		}

		// No matching rule — create a new one
		rule := config.Rule{
			Dest:     dest,
			Files:    []string{filename},
			Locations: locs,
		}
		cfg.Rules = append(cfg.Rules, rule)
		if err := config.Save(root, cfg); err != nil {
			return err
		}

		results, err := installer.SyncRule(root, rule, selectedLocs, false)
		if err != nil {
			return err
		}
		fmt.Printf("\nAdded %s:\n\n", filename)
		for _, r := range results {
			fmt.Printf("  %s → %s\n", r.Package, r.Dest)
		}
		return nil
	},
}

func samePackages(a, b []config.Location) bool {
	if len(a) != len(b) {
		return false
	}
	m := map[string]bool{}
	for _, p := range a {
		m[p.Path] = true
	}
	for _, p := range b {
		if !m[p.Path] {
			return false
		}
	}
	return true
}

func init() {
	rootCmd.AddCommand(addCmd)
}
