package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const Filename = "valet.yaml"

type RulePackage struct {
	Path string `yaml:"path"`
}

type Rule struct {
	Dest     string        `yaml:"dest,omitempty"`
	Files    []string      `yaml:"files"`
	Packages []RulePackage `yaml:"packages"`
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
