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

	src := filepath.Join(root, "shared.md")
	os.WriteFile(src, []byte("# shared rules"), 0644)
	os.MkdirAll(filepath.Join(root, "packages/auth"), 0755)
	os.MkdirAll(filepath.Join(root, "packages/api"), 0755)

	rule = config.Rule{File: "shared.md", Dest: "shared.md"}
	return
}

func TestInstallFileCopies(t *testing.T) {
	root, rule := setup(t)
	results, err := InstallFile(root, rule, []string{"packages/auth", "packages/api"}, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	for _, r := range results {
		if _, err := os.Stat(r.Dest); err != nil {
			t.Errorf("expected file at %s, got: %v", r.Dest, err)
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

func TestInstallFileExclude(t *testing.T) {
	root, rule := setup(t)
	rule.Exclude = []string{"packages/api"}

	results, err := InstallFile(root, rule, []string{"packages/auth", "packages/api"}, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	skipped := 0
	for _, r := range results {
		if r.Skipped {
			skipped++
		}
	}
	if skipped != 1 {
		t.Errorf("expected 1 skipped, got %d", skipped)
	}
	if _, err := os.Stat(filepath.Join(root, "packages/api/shared.md")); !os.IsNotExist(err) {
		t.Error("excluded package should not have file installed")
	}
}

func TestInstallFileMissingSource(t *testing.T) {
	root := t.TempDir()
	rule := config.Rule{File: "nonexistent.md", Dest: "nonexistent.md"}
	_, err := InstallFile(root, rule, []string{"."}, false)
	if err == nil {
		t.Error("expected error for missing source file")
	}
}
