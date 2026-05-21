package installer

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"github.com/tq303/val/internal/config"
)

type Result struct {
	Package string
	File    string
	Dest    string
	Skipped bool
}

// InstallFile copies a single rule's source file into each of the given packages,
// skipping any listed in rule.Exclude.
func InstallFile(root string, rule config.Rule, packages []string, dryRun bool) ([]Result, error) {
	src := filepath.Join(root, rule.File)
	if _, err := os.Stat(src); err != nil {
		return nil, fmt.Errorf("source file not found: %s", rule.File)
	}

	excluded := map[string]bool{}
	for _, e := range rule.Exclude {
		excluded[e] = true
	}

	var results []Result
	for _, pkg := range packages {
		if excluded[pkg] {
			results = append(results, Result{Package: pkg, File: rule.File, Skipped: true})
			continue
		}
		dest := filepath.Join(root, pkg, rule.Dest)
		results = append(results, Result{Package: pkg, File: rule.File, Dest: dest})
		if dryRun {
			continue
		}
		if err := copyFile(src, dest); err != nil {
			return nil, fmt.Errorf("failed to install %s into %s: %w", rule.File, pkg, err)
		}
	}
	return results, nil
}

// InstallAll applies every rule in cfg to its target packages.
func InstallAll(root string, cfg *config.Config, dryRun bool) ([]Result, error) {
	pkgPaths := make([]string, len(cfg.Packages))
	for i, p := range cfg.Packages {
		pkgPaths[i] = p.Path
	}

	var all []Result
	for _, rule := range cfg.Rules {
		results, err := InstallFile(root, rule, pkgPaths, dryRun)
		if err != nil {
			return nil, err
		}
		all = append(all, results...)
	}
	return all, nil
}

func copyFile(src, dest string) error {
	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return err
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}
