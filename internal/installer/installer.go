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
}

// SyncRule copies all files in a rule into the given packages.
func SyncRule(root string, rule config.Rule, packages []string, dryRun bool) ([]Result, error) {
	var results []Result
	for _, file := range rule.Files {
		src := file
		if !filepath.IsAbs(src) {
			src = filepath.Join(root, src)
		}
		if _, err := os.Stat(src); err != nil {
			return nil, fmt.Errorf("source file not found: %s", file)
		}

		for _, pkg := range packages {
			destDir := pkg
			if rule.Dest != "" {
				destDir = filepath.Join(pkg, rule.Dest)
			}
			dest := filepath.Join(root, destDir, filepath.Base(file))
			results = append(results, Result{Package: pkg, File: file, Dest: dest})
			if dryRun {
				continue
			}
			if err := copyFile(src, dest); err != nil {
				return nil, fmt.Errorf("failed to sync %s into %s: %w", file, pkg, err)
			}
		}
	}
	return results, nil
}

// SyncAll applies every rule in cfg to its own package list.
func SyncAll(root string, cfg *config.Config, dryRun bool) ([]Result, error) {
	var all []Result
	for _, rule := range cfg.Rules {
		pkgs := make([]string, len(rule.Packages))
		for i, p := range rule.Packages {
			pkgs[i] = p.Path
		}
		results, err := SyncRule(root, rule, pkgs, dryRun)
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
