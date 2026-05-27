package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const Filename = "valet.yaml"

type Rule struct {
	Repo      string   `yaml:"repo,omitempty"`
	Archive   bool     `yaml:"archive,omitempty"`
	Extract   []string `yaml:"extract,omitempty"`
	Dest      string   `yaml:"dest,omitempty"`
	Files     []string `yaml:"files"`
	Copy      bool     `yaml:"copy,omitempty"`
	Locations []string `yaml:"locations"`
	Platforms []string `yaml:"platforms,omitempty"`
}

type Config struct {
	Version int    `yaml:"version"`
	Rules   []Rule `yaml:"rules"`
}

func FindRoot(cwd string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return cwd, nil
	}
	dir := cwd
	for {
		if _, err := os.Stat(filepath.Join(dir, Filename)); err == nil {
			return dir, nil
		}
		if dir == home {
			break
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return cwd, nil
}

func Load(root string) (*Config, error) {
	return LoadFile(filepath.Join(root, Filename))
}

func LoadFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return &Config{Version: 1}, nil
	}
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func Save(root string, cfg *Config) error {
	return SaveFile(filepath.Join(root, Filename), cfg)
}

func SaveFile(path string, cfg *Config) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
