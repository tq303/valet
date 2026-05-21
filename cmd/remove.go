package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
	"github.com/tq303/valet/internal/config"
	"github.com/tq303/valet/internal/installer"
)

var removeFiles bool

var removeCmd = &cobra.Command{
	Use:   "remove [file]",
	Short: "Remove a file from valet.yaml",
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
			fmt.Println("No files configured.")
			return nil
		}

		var file string
		if len(args) == 1 {
			file = args[0]
		} else {
			var options []huh.Option[string]
			seen := map[string]bool{}
			for _, rule := range cfg.Rules {
				for _, f := range rule.Files {
					name := installer.FileName(f)
					if !seen[name] {
						options = append(options, huh.NewOption(name, f))
						seen[name] = true
					}
				}
			}
			if err := huh.NewForm(
				huh.NewGroup(
					huh.NewSelect[string]().
						Title("Which file to remove?").
						Options(options...).
						Value(&file),
				),
			).Run(); err != nil {
				return err
			}
		}

		filename := installer.FileName(file)
		found := false

		var newRules []config.Rule
		var deletePaths []string
		for _, rule := range cfg.Rules {
			var newFiles []string
			for _, f := range rule.Files {
				if installer.FileName(f) == filename {
					found = true
					if removeFiles {
						for _, loc := range rule.Locations {
							destDir := installer.ResolvePath(root, loc)
							if rule.Dest != "" {
								destDir = filepath.Join(destDir, rule.Dest)
							}
							deletePaths = append(deletePaths, filepath.Join(destDir, filename))
						}
					}
					continue
				}
				newFiles = append(newFiles, f)
			}
			if len(newFiles) > 0 {
				rule.Files = newFiles
				newRules = append(newRules, rule)
			}
		}

		if !found {
			return fmt.Errorf("%s is not tracked", file)
		}

		cfg.Rules = newRules
		if err := config.Save(root, cfg); err != nil {
			return err
		}

		for _, p := range deletePaths {
			os.RemoveAll(p)
		}

		if removeFiles && len(deletePaths) > 0 {
			fmt.Printf("Removed %s from valet.yaml and deleted %d synced copies.\n", filename, len(deletePaths))
		} else {
			fmt.Printf("Removed %s from valet.yaml.\n", filename)
		}
		return nil
	},
}

func init() {
	removeCmd.Flags().BoolVarP(&removeFiles, "files", "f", false, "Also delete synced copies from all locations")
	rootCmd.AddCommand(removeCmd)
}
