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
			{File: ".eslintrc.js", Dest: ".eslintrc.js"},
			{File: ".rules/.claude/auth.md", Preset: "claude", Dest: ".claude/CLAUDE.md", Exclude: []string{"packages/ui"}},
		},
		Packages: []Package{
			{Path: "."},
			{Path: "packages/auth"},
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
	if loaded.Rules[1].Exclude[0] != "packages/ui" {
		t.Errorf("exclude not preserved, got %v", loaded.Rules[1].Exclude)
	}
	if len(loaded.Packages) != 2 {
		t.Errorf("expected 2 packages, got %d", len(loaded.Packages))
	}
}
