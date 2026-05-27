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

var once bool
var locationFlags []string
var destFlag string
var platformFlags []string

var addCmd = &cobra.Command{
	Use:   "add [files...]",
	Short: "Add a file to be synced across locations",
	Long:  "Add any file or URL to be tracked and synced. If already tracked, promotes or extends it.",
	Args:  cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		root, err := findRoot()
		if err != nil {
			return err
		}

		// Non-interactive mode: files + -l flags provided
		if len(locationFlags) > 0 {
			if len(args) == 0 {
				return fmt.Errorf("at least one file is required with --location")
			}
			cfg, err := loadConfig(root)
			if err != nil {
				return err
			}
			return addWithFlags(root, cfg, args, locationFlags, destFlag, platformFlags, once)
		}

		// Collect single file (prompt if not provided)
		var file string
		if len(args) >= 1 {
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

		// Multiple files without -l: prompt once, apply to all
		if len(args) > 1 {
			cfg, err := loadConfig(root)
			if err != nil {
				return err
			}
			return addMultipleFiles(root, cfg, args, once)
		}

		if once {
			return addOnce(root, file)
		}

		cfg, err := loadConfig(root)
		if err != nil {
			return err
		}

		if installer.IsGitRepo(file) {
			return addFromRepo(root, cfg, file)
		}

		if installer.IsArchiveURL(file) {
			return addFromArchive(root, cfg, file)
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

				// Extend — prompt for additional locations
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

				newLocs, err := promptAdditionalLocations([]string{loc})
				if err != nil {
					return err
				}

				var added []string
				for _, l := range newLocs {
					already := false
					for _, existing := range rule.Locations {
						if existing == l {
							already = true
							break
						}
					}
					if !already {
						cfg.Rules[i].Locations = append(cfg.Rules[i].Locations, l)
						added = append(added, l)
					}
				}

				if len(added) == 0 {
					fmt.Println("All locations already tracked.")
					return nil
				}

				if err := saveConfig(root, cfg); err != nil {
					return err
				}
				results, err := installer.SyncRule(root, cfg.Rules[i], added, false, true)
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

		locs, err := promptAdditionalLocations([]string{loc})
		if err != nil {
			return err
		}

		// Group into an existing rule with the same dest and locations if possible
		for i, rule := range cfg.Rules {
			if rule.Dest == dest && len(rule.Locations) == 1 && rule.Locations[0] == loc {
				cfg.Rules[i].Files = append(cfg.Rules[i].Files, file)
				cfg.Rules[i].Locations = locs
				if err := saveConfig(root, cfg); err != nil {
					return err
				}
				results, err := installer.SyncRule(root, cfg.Rules[i], locs, false, true)
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
			Locations: locs,
			Platforms: platformFlags,
		}
		cfg.Rules = append(cfg.Rules, rule)
		if err := saveConfig(root, cfg); err != nil {
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

func addWithFlags(root string, cfg *config.Config, files, locs []string, dest string, platforms []string, ignoreConfig bool) error {
	for _, f := range files {
		if !installer.IsURL(f) && !installer.IsGitRepo(f) {
			src := installer.ResolvePath(root, f)
			if _, err := os.Stat(src); err != nil {
				return fmt.Errorf("file not found: %s", f)
			}
		}
	}

	rule := config.Rule{Dest: dest, Files: files, Locations: locs, Platforms: platforms}

	if !ignoreConfig {
		for i, r := range cfg.Rules {
			if r.Dest == dest && locationsMatch(r.Locations, locs) {
				cfg.Rules[i].Files = append(cfg.Rules[i].Files, files...)
				if err := saveConfig(root, cfg); err != nil {
					return err
				}
				rule = cfg.Rules[i]
				break
			}
		}
		if len(rule.Files) == len(files) {
			cfg.Rules = append(cfg.Rules, rule)
			if err := saveConfig(root, cfg); err != nil {
				return err
			}
		}
	}

	results, err := installer.SyncRule(root, rule, locs, false, true)
	if err != nil {
		return err
	}
	fmt.Println()
	for _, r := range results {
		fmt.Printf("  %s → %s\n", installer.FileName(r.File), r.Dest)
	}
	return nil
}

func addMultipleFiles(root string, cfg *config.Config, files []string, ignoreConfig bool) error {
	for _, f := range files {
		if !installer.IsURL(f) && !installer.IsGitRepo(f) {
			src := installer.ResolvePath(root, f)
			if _, err := os.Stat(src); err != nil {
				return fmt.Errorf("file not found: %s", f)
			}
		}
	}

	var loc, dest string
	if err := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Location to sync to").
				Placeholder("packages/auth").
				Value(&loc),
			huh.NewInput().
				Title("Destination folder (leave blank for root)").
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

	locs, err := promptAdditionalLocations([]string{loc})
	if err != nil {
		return err
	}

	return addWithFlags(root, cfg, files, locs, dest, platformFlags, ignoreConfig)
}

func locationsMatch(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
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
			if err := saveConfig(root, cfg); err != nil {
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
			if err := saveConfig(root, cfg); err != nil {
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
	if err := saveConfig(root, cfg); err != nil {
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

func addFromArchive(root string, cfg *config.Config, archiveURL string) error {
	var extractInput string
	if err := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("File(s) to extract from archive (space-separated)").
				Placeholder("gifski").
				Value(&extractInput),
		),
	).Run(); err != nil {
		return err
	}
	if extractInput == "" {
		fmt.Println("No files specified, nothing to do.")
		return nil
	}
	extractPaths := strings.Fields(extractInput)

	var loc, dest string
	if err := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Location to sync to").
				Placeholder("~/bin").
				Value(&loc),
			huh.NewInput().
				Title("Destination folder in each location (leave blank for root)").
				Value(&dest),
		),
	).Run(); err != nil {
		return err
	}
	if loc == "" {
		fmt.Println("No location entered, nothing to do.")
		return nil
	}

	locs, err := promptAdditionalLocations([]string{loc})
	if err != nil {
		return err
	}

	fmt.Printf("Downloading and extracting %s...\n", archiveURL)
	if _, err := installer.EnsureArchive(archiveURL, extractPaths, false); err != nil {
		return fmt.Errorf("failed to extract archive: %w", err)
	}

	rule := config.Rule{
		Archive:   true,
		Extract:   extractPaths,
		Files:     []string{archiveURL},
		Dest:      dest,
		Locations: locs,
	}
	cfg.Rules = append(cfg.Rules, rule)
	if err := saveConfig(root, cfg); err != nil {
		return err
	}

	results, err := installer.SyncRule(root, rule, locs, false, true)
	if err != nil {
		return err
	}
	for _, extractPath := range extractPaths {
		name := filepath.Base(extractPath)
		fmt.Printf("\nAdded %s:\n\n", name)
		for _, r := range results {
			if filepath.Base(r.File) == name {
				fmt.Printf("  %s → %s\n", r.Package, r.Dest)
			}
		}
	}
	return nil
}

func addOnce(root, file string) error {
	isRepo := installer.IsGitRepo(file)
	repoFile := ""

	var cacheDir string
	if isRepo {
		fmt.Printf("Cloning %s...\n", file)
		var err error
		cacheDir, err = installer.EnsureRepo(file)
		if err != nil {
			return fmt.Errorf("failed to clone repo: %w", err)
		}
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
	} else if !installer.IsURL(file) {
		src := installer.ResolvePath(root, file)
		if _, err := os.Stat(src); err != nil {
			return fmt.Errorf("file not found: %s", file)
		}
	}

	var loc, dest string
	if err := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Location to sync to").
				Placeholder("packages/auth").
				Value(&loc),
			huh.NewInput().
				Title("Destination folder (leave blank for root)").
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

	locs, err := promptAdditionalLocations([]string{loc})
	if err != nil {
		return err
	}

	srcFile := file
	if isRepo {
		srcFile = repoFile
	}

	rule := config.Rule{
		Dest:      dest,
		Files:     []string{srcFile},
		Locations: locs,
	}
	if isRepo {
		rule.Repo = file
	}

	fileRoot := root
	if isRepo {
		fileRoot = cacheDir
	}

	results, err := installer.SyncRule(fileRoot, rule, locs, false, true)
	if err != nil {
		return err
	}
	fmt.Printf("\nSynced %s:\n\n", installer.FileName(srcFile))
	for _, r := range results {
		fmt.Printf("  %s → %s\n", r.Package, r.Dest)
	}
	return nil
}

func promptAdditionalLocations(locs []string) ([]string, error) {
	for {
		var next string
		if err := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Add another location? (leave blank to finish)").
					Value(&next),
			),
		).Run(); err != nil {
			return nil, err
		}
		if next == "" {
			break
		}
		locs = append(locs, next)
	}
	return locs, nil
}

func init() {
	addCmd.Flags().BoolVarP(&once, "ignore", "i", false, "Sync without adding to valet.yaml")
	addCmd.Flags().StringArrayVarP(&locationFlags, "location", "l", nil, "Location to sync to (repeatable)")
	addCmd.Flags().StringVar(&destFlag, "dest", "", "Destination folder within each location")
	addCmd.Flags().StringArrayVar(&platformFlags, "platform", nil, "Restrict rule to platform(s): darwin, linux, windows (repeatable)")
	rootCmd.AddCommand(addCmd)
}
