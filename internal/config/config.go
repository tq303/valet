package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const Filename = "valet.yaml"

type Rule struct {
	Repo      string   `yaml:"repo,omitempty"`
	Dest      string   `yaml:"dest,omitempty"`
	Files     []string `yaml:"files"`
	Link      bool     `yaml:"link,omitempty"`
	Locations []string `yaml:"locations"`
}

type Config struct {
	Version int    `yaml:"version"`
	Rules   []Rule `yaml:"rules"`
}

func Load(root string) (*Config, error) {
	data, err := os.ReadFile(filepath.Join(root, Filename))
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
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(root, Filename), data, 0644)
}
