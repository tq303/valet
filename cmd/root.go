package cmd

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/tq303/valet/internal/config"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "val",
	Short: "Valet — sync config files across your monorepo",
	Long:  "Valet manages and syncs files across locations. Add any file, URL or repo location once or cache to local file.",
}

func Execute() {
	rootCmd.CompletionOptions.HiddenDefaultCmd = true
	rootCmd.SetHelpCommand(&cobra.Command{Hidden: true})
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func findRoot() (string, error) {
	if cfgFile != "" {
		abs, err := filepath.Abs(cfgFile)
		if err != nil {
			return "", err
		}
		return filepath.Dir(abs), nil
	}
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return config.FindRoot(cwd)
}

func loadConfig(root string) (*config.Config, error) {
	if cfgFile != "" {
		abs, err := filepath.Abs(cfgFile)
		if err != nil {
			return nil, err
		}
		return config.LoadFile(abs)
	}
	return config.Load(root)
}

func saveConfig(root string, cfg *config.Config) error {
	if cfgFile != "" {
		abs, err := filepath.Abs(cfgFile)
		if err != nil {
			return err
		}
		return config.SaveFile(abs, cfg)
	}
	return config.Save(root, cfg)
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "Path to a valet.yaml config file")
}
