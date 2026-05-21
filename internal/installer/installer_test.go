package installer

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/tq303/val/internal/config"
)

func setup(t *testing.T) (root string, rule config.Rule) {
	t.Helper()
	root = t.TempDir()

	os.WriteFile(filepath.Join(root, "shared.md"), []byte("# shared rules"), 0644)
	os.MkdirAll(filepath.Join(root, "packages/auth"), 0755)
	os.MkdirAll(filepath.Join(root, "packages/api"), 0755)

	rule = config.Rule{
		File: "shared.md",
		Name: "shared.md",
		Packages: []config.RulePackage{
			{Path: "packages/auth"},
			{Path: "packages/api"},
		},
	}
	return
}

func TestInstallFileCopies(t *testing.T) {
	root, rule := setup(t)
	pkgs := []string{"packages/auth", "packages/api"}
	results, err := InstallFile(root, rule, pkgs, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	for _, r := range results {
		if _, err := os.Stat(r.Dest); err != nil {
			t.Errorf("expected file at %s: %v", r.Dest, err)
		}
	}
}

func TestInstallFileDryRun(t *testing.T) {
	root, rule := setup(t)
	results, err := InstallFile(root, rule, []string{"packages/auth"}, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := os.Stat(results[0].Dest); !os.IsNotExist(err) {
		t.Error("dry-run should not write files")
	}
}

func TestInstallFileMissingSource(t *testing.T) {
	root := t.TempDir()
	rule := config.Rule{File: "nonexistent.md", Name: "nonexistent.md"}
	_, err := InstallFile(root, rule, []string{"."}, false)
	if err == nil {
		t.Error("expected error for missing source file")
	}
}

func TestInstallAll(t *testing.T) {
	root, rule := setup(t)
	cfg := &config.Config{
		Version: 1,
		Rules:   []config.Rule{rule},
	}
	results, err := InstallAll(root, cfg, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
}
