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
	Short: "Add a file to be synced across locations",
	Long:  "Add any file to be tracked and synced. If already tracked, promotes or extends it.",
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
		src := installer.ResolvePath(root, file)
		if _, err := os.Stat(src); err != nil {
			return fmt.Errorf("file not found: %s", file)
		}

		filename := filepath.Base(file)
		dir := filepath.Dir(file)

		cfg, err := config.Load(root)
		if err != nil {
			return err
		}

		// Check if file is already tracked in any rule
		for i, rule := range cfg.Rules {
			for _, f := range rule.Files {
				if filepath.Base(f) == filename {
					// Check if passed path is from inside a known location — promote flow
					for _, loc := range rule.Locations {
						resolvedLoc := installer.ResolvePath(root, loc)
						resolvedDir := installer.ResolvePath(root, dir)
						if resolvedLoc == resolvedDir {
							// Promote: copy this version over the source, sync everywhere
							ruleSrc := installer.ResolvePath(root, f)
							if err := installer.CopyFile(src, ruleSrc); err != nil {
								return fmt.Errorf("failed to promote %s: %w", filename, err)
							}
							results, err := installer.SyncRule(root, cfg.Rules[i], rule.Locations, false)
							if err != nil {
								return err
							}
							fmt.Printf("\nPromoted %s from %s and synced to:\n\n", filename, dir)
							for _, r := range results {
								fmt.Printf("  %s → %s\n", r.Package, r.Dest)
							}
							return nil
						}
					}

					// Not a promote — show locations that don't have it yet
					existing := map[string]bool{}
					for _, rp := range rule.Locations {
						existing[rp] = true
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
						fmt.Printf("%s is already tracked for all locations.\n", filename)
						return nil
					}

					var selectedLocs []string
					if err := huh.NewForm(
						huh.NewGroup(
							huh.NewMultiSelect[string]().
								Title(fmt.Sprintf("Add %s to which locations?", filename)).
								Options(locOptions...).
								Value(&selectedLocs),
						),
					).Run(); err != nil {
						return err
					}

					if len(selectedLocs) == 0 {
						fmt.Println("No locations selected, nothing to do.")
						return nil
					}

					cfg.Rules[i].Locations = append(cfg.Rules[i].Locations, selectedLocs...)
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

		// New file — prompt for dest folder and locations
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
					Title("Which locations should this apply to?").
					Options(locOptions...).
					Value(&selectedLocs),
				huh.NewInput().
					Title("Destination folder in each location (leave blank for root)").
					Placeholder(".claude").
					Value(&dest),
			),
		).Run(); err != nil {
			return err
		}

		if len(selectedLocs) == 0 {
			fmt.Println("No locations selected, nothing to do.")
			return nil
		}

		// Find an existing rule with same dest and locations to group into
		for i, rule := range cfg.Rules {
			if rule.Dest == dest && sameLocations(rule.Locations, selectedLocs) {
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
			Dest:      dest,
			Files:     []string{filename},
			Locations: selectedLocs,
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

func sameLocations(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	m := map[string]bool{}
	for _, p := range a {
		m[p] = true
	}
	for _, p := range b {
		if !m[p] {
			return false
		}
	}
	return true
}

func init() {
	rootCmd.AddCommand(addCmd)
}
