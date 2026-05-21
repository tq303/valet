package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
	"github.com/tq303/val/internal/config"
	"github.com/tq303/val/internal/discovery"
	"github.com/tq303/val/internal/preset"
	valrules "github.com/tq303/val/internal/rules"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialise Valet in a monorepo",
	Long:  "Detects monorepo structure, scans rules, and creates a valet.yaml config file.",
	RunE: func(cmd *cobra.Command, args []string) error {
		root, err := os.Getwd()
		if err != nil {
			return err
		}

		// Discover packages
		fmt.Println("Scanning monorepo...")
		packages, err := discovery.Discover(root)
		if err != nil {
			return err
		}
		fmt.Printf("Found %d package(s):\n\n", len(packages))
		for _, p := range packages {
			fmt.Printf("  %s\n", p.Path)
		}

		// Ensure .rules exists
		rulesDir := filepath.Join(root, ".rules")
		if _, err := os.Stat(rulesDir); os.IsNotExist(err) {
			if err := os.Mkdir(rulesDir, 0755); err != nil {
				return fmt.Errorf("could not create .rules folder: %w", err)
			}
			fmt.Println("\nCreated .rules/ — add your rule files (.md) then re-run `val init`.")
			return nil
		}

		// Scan rule files
		scanned, err := valrules.ScanFiles(rulesDir)
		if err != nil {
			return err
		}
		if len(scanned) == 0 {
			fmt.Println("\n.rules/ is empty — add your rule files (.md) then re-run `val init`.")
			return nil
		}

		// Separate pre-assigned (tool subdir) from unassigned (root level)
		var unassigned []valrules.ScannedRule
		var configRules []config.Rule

		for _, r := range scanned {
			if r.Tools != nil {
				// Use preset to resolve dest for each assigned tool
				for _, tool := range r.Tools {
					configRules = append(configRules, config.Rule{
						File:   r.Name,
						Preset: tool,
						Dest:   preset.Dest(tool, filepath.Base(r.Name)),
					})
				}
			} else {
				unassigned = append(unassigned, r)
			}
		}

		if len(unassigned) > 0 {
			var fileOptions []huh.Option[string]
			for _, r := range unassigned {
				fileOptions = append(fileOptions, huh.NewOption(r.Name, r.Path))
			}

			var selectedPaths []string
			if err := huh.NewForm(
				huh.NewGroup(
					huh.NewMultiSelect[string]().
						Title("Select rules to include").
						Options(fileOptions...).
						Value(&selectedPaths),
				),
			).Run(); err != nil {
				return err
			}

			toolOptions := []huh.Option[string]{
				huh.NewOption("Claude Code", preset.Claude),
				huh.NewOption("Cursor", preset.Cursor),
			}

			for _, path := range selectedPaths {
				name := filepath.Base(path)
				var selectedTools []string
				if err := huh.NewForm(
					huh.NewGroup(
						huh.NewMultiSelect[string]().
							Title(fmt.Sprintf("Which tools should %s apply to?", name)).
							Options(toolOptions...).
							Value(&selectedTools),
					),
				).Run(); err != nil {
					return err
				}
				for _, tool := range selectedTools {
					configRules = append(configRules, config.Rule{
						File:   name,
						Preset: tool,
						Dest:   preset.Dest(tool, name),
					})
				}
			}
		}

		// Build packages list
		var configPackages []config.Package
		for _, p := range packages {
			configPackages = append(configPackages, config.Package{Path: p.Path})
		}

		cfg := config.Config{
			Version:  1,
			Rules:    configRules,
			Packages: configPackages,
		}
		if err := config.Save(root, &cfg); err != nil {
			return fmt.Errorf("could not write valet.yaml: %w", err)
		}

		fmt.Println("\nCreated valet.yaml — run `val install` to apply rules to packages.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
