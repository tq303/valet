package discovery

import (
	"encoding/json"
	"os"
	"path/filepath"
	"gopkg.in/yaml.v3"
)

type Tool string

const (
	ToolClaude Tool = "claude"
	ToolCursor Tool = "cursor"
)

type Package struct {
	Name  string
	Path  string
	Tools []Tool
}

// Discover finds all packages in the monorepo rooted at root.
// It tries npm/yarn workspaces first, then pnpm.
func Discover(root string) ([]Package, error) {
	paths, err := discoverPaths(root)
	if err != nil {
		return nil, err
	}
	var packages []Package

	// Root is always included
	packages = append(packages, Package{
		Name:  filepath.Base(root),
		Path:  ".",
		Tools: detectTools(root, "."),
	})

	for _, p := range paths {
		packages = append(packages, Package{
			Name:  filepath.Base(p),
			Path:  p,
			Tools: detectTools(root, p),
		})
	}
	return packages, nil
}

func discoverPaths(root string) ([]string, error) {
	// Try npm/yarn (package.json workspaces)
	paths, err := discoverNPM(root)
	if err == nil && len(paths) > 0 {
		return paths, nil
	}

	// Try pnpm (pnpm-workspace.yaml)
	paths, err = discoverPNPM(root)
	if err == nil && len(paths) > 0 {
		return paths, nil
	}

	return nil, nil
}

func discoverNPM(root string) ([]string, error) {
	data, err := os.ReadFile(filepath.Join(root, "package.json"))
	if err != nil {
		return nil, err
	}

	var pkg struct {
		Workspaces interface{} `json:"workspaces"`
	}
	if err := json.Unmarshal(data, &pkg); err != nil {
		return nil, err
	}

	// workspaces can be []string or {"packages": []string}
	var globs []string
	switch v := pkg.Workspaces.(type) {
	case []interface{}:
		for _, g := range v {
			if s, ok := g.(string); ok {
				globs = append(globs, s)
			}
		}
	case map[string]interface{}:
		if pkgs, ok := v["packages"].([]interface{}); ok {
			for _, g := range pkgs {
				if s, ok := g.(string); ok {
					globs = append(globs, s)
				}
			}
		}
	}

	return expandGlobs(root, globs)
}

func discoverPNPM(root string) ([]string, error) {
	data, err := os.ReadFile(filepath.Join(root, "pnpm-workspace.yaml"))
	if err != nil {
		return nil, err
	}

	var workspace struct {
		Packages []string `yaml:"packages"`
	}
	if err := yaml.Unmarshal(data, &workspace); err != nil {
		return nil, err
	}

	return expandGlobs(root, workspace.Packages)
}

func expandGlobs(root string, globs []string) ([]string, error) {
	var paths []string
	seen := map[string]bool{}

	for _, pattern := range globs {
		matches, err := filepath.Glob(filepath.Join(root, pattern))
		if err != nil {
			return nil, err
		}
		for _, m := range matches {
			info, err := os.Stat(m)
			if err != nil || !info.IsDir() {
				continue
			}
			rel, err := filepath.Rel(root, m)
			if err != nil {
				continue
			}
			if !seen[rel] {
				seen[rel] = true
				paths = append(paths, rel)
			}
		}
	}
	return paths, nil
}

func detectTools(root, pkgPath string) []Tool {
	var tools []Tool
	base := filepath.Join(root, pkgPath)

	if dirExists(filepath.Join(base, ".claude")) {
		tools = append(tools, ToolClaude)
	}
	if dirExists(filepath.Join(base, ".cursor")) {
		tools = append(tools, ToolCursor)
	}
	return tools
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
