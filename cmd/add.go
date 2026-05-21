package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
	"github.com/tq303/valet/internal/config"
	"github.com/tq303/valet/internal/installer"
)

var addCmd = &cobra.Command{
	Use:   "add [file]",
	Short: "Add a file to be synced across locations",
	Long:  "Add any file or URL to be tracked and synced. If already tracked, promotes or extends it.",
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
						Title("File or URL to add").
						Placeholder("CLAUDE.md").
						Value(&file),
				),
			).Run(); err != nil {
				return err
			}
		}

		cfg, err := config.Load(root)
		if err != nil {
			return err
		}

		if installer.IsGitRepo(file) {
			return addFromRepo(root, cfg, file)
		}

		isURL := installer.IsURL(file)
		filename := installer.FileName(file)
		dir := ""

		if !isURL {
			src := installer.ResolvePath(root, file)
			if _, err := os.Stat(src); err != nil {
				return fmt.Errorf("file not found: %s", file)
			}
			dir = filepath.Dir(file)
		}

		// Check if file is already tracked in any rule
		for i, rule := range cfg.Rules {
			for _, f := range rule.Files {
				if installer.FileName(f) != filename {
					continue
				}

				// Promote flow — only for local files whose path is inside a known location
				if !isURL {
					abs := installer.ResolvePath(root, file)
					absDir := filepath.Dir(abs)
					for _, loc := range rule.Locations {
						if installer.ResolvePath(root, loc) == absDir {
							ruleSrc := installer.ResolvePath(root, f)
							if err := installer.CopyFile(abs, ruleSrc); err != nil {
								return fmt.Errorf("failed to promote %s: %w", filename, err)
							}
							results, err := installer.SyncRule(root, cfg.Rules[i], rule.Locations, false, true)
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
				}

				// Extend — prompt for an additional location
				var loc string
				if err := huh.NewForm(
					huh.NewGroup(
						huh.NewInput().
							Title(fmt.Sprintf("Add %s to which location?", filename)).
							Placeholder("packages/web").
							Value(&loc),
					),
				).Run(); err != nil {
					return err
				}

				if loc == "" {
					fmt.Println("No location entered, nothing to do.")
					return nil
				}

				for _, existing := range rule.Locations {
					if existing == loc {
						fmt.Printf("%s is already synced to %s.\n", filename, loc)
						return nil
					}
				}

				cfg.Rules[i].Locations = append(cfg.Rules[i].Locations, loc)
				if err := config.Save(root, cfg); err != nil {
					return err
				}
				results, err := installer.SyncRule(root, cfg.Rules[i], []string{loc}, false, true)
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

		// New file — prompt for location and dest folder
		var loc, dest string
		if err := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Location to sync to").
					Placeholder("packages/auth").
					Value(&loc),
				huh.NewInput().
					Title("Destination folder in each location (leave blank for root)").
					Placeholder(".claude").
					Value(&dest),
			),
		).Run(); err != nil {
			return err
		}

		if loc == "" {
			fmt.Println("No location entered, nothing to do.")
			return nil
		}

		// Group into an existing rule with the same dest and location if possible
		for i, rule := range cfg.Rules {
			if rule.Dest == dest && len(rule.Locations) == 1 && rule.Locations[0] == loc {
				cfg.Rules[i].Files = append(cfg.Rules[i].Files, file)
				if err := config.Save(root, cfg); err != nil {
					return err
				}
				results, err := installer.SyncRule(root, cfg.Rules[i], []string{loc}, false, true)
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
			Files:     []string{file},
			Locations: []string{loc},
		}
		cfg.Rules = append(cfg.Rules, rule)
		if err := config.Save(root, cfg); err != nil {
			return err
		}

		results, err := installer.SyncRule(root, rule, []string{loc}, false, true)
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

func addFromRepo(root string, cfg *config.Config, repoURL string) error {
	fmt.Printf("Cloning %s...\n", repoURL)
	cacheDir, err := installer.EnsureRepo(repoURL)
	if err != nil {
		return fmt.Errorf("failed to clone repo: %w", err)
	}

	var repoFile string
	if err := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("File or folder within repo").
				Placeholder("valyu-best-practices/").
				Value(&repoFile),
		),
	).Run(); err != nil {
		return err
	}
	if repoFile == "" {
		fmt.Println("No file entered, nothing to do.")
		return nil
	}

	src := filepath.Join(cacheDir, strings.TrimSuffix(repoFile, "/"))
	if _, err := os.Stat(src); err != nil {
		return fmt.Errorf("not found in repo: %s", repoFile)
	}

	filename := installer.FileName(repoFile)

	// Check if already tracked in this repo
	for i, rule := range cfg.Rules {
		if rule.Repo != repoURL {
			continue
		}
		for _, f := range rule.Files {
			if f != repoFile {
				continue
			}
			// Extend with a new location
			var loc string
			if err := huh.NewForm(
				huh.NewGroup(
					huh.NewInput().
						Title(fmt.Sprintf("Add %s to which location?", filename)).
						Placeholder("packages/web").
						Value(&loc),
				),
			).Run(); err != nil {
				return err
			}
			if loc == "" {
				fmt.Println("No location entered, nothing to do.")
				return nil
			}
			for _, existing := range rule.Locations {
				if existing == loc {
					fmt.Printf("%s is already synced to %s.\n", filename, loc)
					return nil
				}
			}
			cfg.Rules[i].Locations = append(cfg.Rules[i].Locations, loc)
			if err := config.Save(root, cfg); err != nil {
				return err
			}
			results, err := installer.SyncRule(root, cfg.Rules[i], []string{loc}, false, true)
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

	// New repo file — prompt for location and dest
	var loc, dest string
	if err := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Location to sync to").
				Placeholder("packages/auth").
				Value(&loc),
			huh.NewInput().
				Title("Destination folder in each location (leave blank for root)").
				Placeholder(".claude/commands").
				Value(&dest),
		),
	).Run(); err != nil {
		return err
	}
	if loc == "" {
		fmt.Println("No location entered, nothing to do.")
		return nil
	}

	// Group into existing rule with same repo, dest, and location if possible
	for i, rule := range cfg.Rules {
		if rule.Repo == repoURL && rule.Dest == dest && len(rule.Locations) == 1 && rule.Locations[0] == loc {
			cfg.Rules[i].Files = append(cfg.Rules[i].Files, repoFile)
			if err := config.Save(root, cfg); err != nil {
				return err
			}
			results, err := installer.SyncRule(root, cfg.Rules[i], []string{loc}, false, true)
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

	rule := config.Rule{
		Repo:      repoURL,
		Dest:      dest,
		Files:     []string{repoFile},
		Locations: []string{loc},
	}
	cfg.Rules = append(cfg.Rules, rule)
	if err := config.Save(root, cfg); err != nil {
		return err
	}
	results, err := installer.SyncRule(root, rule, []string{loc}, false, true)
	if err != nil {
		return err
	}
	fmt.Printf("\nAdded %s:\n\n", filename)
	for _, r := range results {
		fmt.Printf("  %s → %s\n", r.Package, r.Dest)
	}
	return nil
}

func init() {
	rootCmd.AddCommand(addCmd)
}
