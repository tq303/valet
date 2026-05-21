package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadMissing(t *testing.T) {
	dir := t.TempDir()
	cfg, err := Load(dir)
	if err != nil {
		t.Fatalf("expected no error for missing file, got: %v", err)
	}
	if cfg.Version != 1 {
		t.Errorf("expected version 1, got %d", cfg.Version)
	}
}

func TestSaveAndLoad(t *testing.T) {
	dir := t.TempDir()

	original := &Config{
		Version: 1,
		Rules: []Rule{
			{
				Files: []string{".eslintrc.js"},
				Locations: []Location{
					{Path: "."},
					{Path: "packages/auth"},
				},
			},
			{
				Dest:  ".claude",
				Files: []string{"CLAUDE.md"},
				Locations: []Location{
					{Path: "packages/api"},
				},
			},
		},
	}

	if err := Save(dir, original); err != nil {
		t.Fatalf("Save failed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, Filename)); err != nil {
		t.Fatalf("valet.yaml not created: %v", err)
	}

	loaded, err := Load(dir)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if len(loaded.Rules) != 2 {
		t.Errorf("expected 2 rules, got %d", len(loaded.Rules))
	}
	if len(loaded.Rules[0].Locations) != 2 {
		t.Errorf("expected 2 locations on first rule, got %d", len(loaded.Rules[0].Locations))
	}
}
