package discovery

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestDiscoverRootFallback(t *testing.T) {
	dir := t.TempDir()
	packages, err := Discover(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(packages) != 1 || packages[0].Path != "." {
		t.Errorf("expected root package, got %v", packages)
	}
}

func TestDiscoverNPMWorkspaces(t *testing.T) {
	dir := t.TempDir()

	pkgJSON, _ := json.Marshal(map[string]interface{}{
		"workspaces": []string{"packages/*"},
	})
	os.WriteFile(filepath.Join(dir, "package.json"), pkgJSON, 0644)
	os.MkdirAll(filepath.Join(dir, "packages/auth"), 0755)
	os.MkdirAll(filepath.Join(dir, "packages/api"), 0755)

	packages, err := Discover(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// root + 2 workspace packages
	if len(packages) != 3 {
		t.Errorf("expected 3 packages, got %d: %v", len(packages), packages)
	}
}

func TestDiscoverNPMWorkspacesObject(t *testing.T) {
	dir := t.TempDir()

	pkgJSON, _ := json.Marshal(map[string]interface{}{
		"workspaces": map[string]interface{}{
			"packages": []string{"apps/*"},
		},
	})
	os.WriteFile(filepath.Join(dir, "package.json"), pkgJSON, 0644)
	os.MkdirAll(filepath.Join(dir, "apps/web"), 0755)

	packages, err := Discover(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(packages) != 2 {
		t.Errorf("expected 2 packages, got %d", len(packages))
	}
}

func TestDiscoverPNPM(t *testing.T) {
	dir := t.TempDir()

	workspace := map[string]interface{}{
		"packages": []string{"packages/*"},
	}
	data, _ := yaml.Marshal(workspace)
	os.WriteFile(filepath.Join(dir, "pnpm-workspace.yaml"), data, 0644)
	os.MkdirAll(filepath.Join(dir, "packages/core"), 0755)

	packages, err := Discover(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(packages) != 2 {
		t.Errorf("expected 2 packages, got %d", len(packages))
	}
}

func TestDetectTools(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, ".claude"), 0755)
	os.MkdirAll(filepath.Join(dir, ".cursor"), 0755)

	tools := detectTools(dir, ".")
	if len(tools) != 2 {
		t.Errorf("expected 2 tools, got %v", tools)
	}
}
