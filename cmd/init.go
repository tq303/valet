package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"github.com/charmbracelet/huh"
	"github.com/tq303/val/internal/config"
	"github.com/tq303/val/internal/discovery"
	valrules "github.com/tq303/val/internal/rules"
	"gopkg.in/yaml.v3"
	"github.com/spf13/cobra"
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

		// Separate pre-assigned rules from ones that need tool selection
		var unassigned []valrules.ScannedRule
		var configRules []config.Rule

		for _, r := range scanned {
			if r.Tools != nil {
				var tools []config.Tool
				for _, t := range r.Tools {
					tools = append(tools, config.Tool(t))
				}
				configRules = append(configRules, config.Rule{File: r.Name, Tools: tools})
			} else {
				unassigned = append(unassigned, r)
			}
		}

		if len(unassigned) > 0 {
			// Step 1: select which unassigned rules to include
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

			// Step 2: for each selected rule, pick tools
			toolOptions := []huh.Option[string]{
				huh.NewOption("Claude Code", string(config.ToolClaude)),
				huh.NewOption("Cursor", string(config.ToolCursor)),
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
				var tools []config.Tool
				for _, t := range selectedTools {
					tools = append(tools, config.Tool(t))
				}
				configRules = append(configRules, config.Rule{File: name, Tools: tools})
			}
		}

		// Build packages list
		var configPackages []config.Package
		for _, p := range packages {
			configPackages = append(configPackages, config.Package{Path: p.Path})
		}

		// Write valet.yaml
		cfg := config.Config{
			Version:  1,
			Rules:    configRules,
			Packages: configPackages,
		}
		data, err := yaml.Marshal(cfg)
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(root, "valet.yaml"), data, 0644); err != nil {
			return fmt.Errorf("could not write valet.yaml: %w", err)
		}

		fmt.Println("\nCreated valet.yaml — run `val install` to apply rules to packages.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
