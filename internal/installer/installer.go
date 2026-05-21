package installer

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/tq303/val/internal/config"
)

type Result struct {
	Package string
	File    string
	Dest    string
}

// SyncRule syncs all files in a rule into the given locations.
func SyncRule(root string, rule config.Rule, locations []string, dryRun bool) ([]Result, error) {
	var results []Result
	for _, file := range rule.Files {
		src := ResolvePath(root, file)
		if _, err := os.Stat(src); err != nil {
			return nil, fmt.Errorf("source not found: %s", file)
		}

		for _, loc := range locations {
			destDir := ResolvePath(root, loc)
			if rule.Dest != "" {
				destDir = filepath.Join(destDir, rule.Dest)
			}
			dest := filepath.Join(destDir, filepath.Base(file))
			results = append(results, Result{Package: loc, File: file, Dest: dest})
			if dryRun {
				continue
			}
			if rule.Link {
				if err := symlink(src, dest); err != nil {
					return nil, fmt.Errorf("failed to symlink %s to %s: %w", file, dest, err)
				}
			} else {
				if err := copyAny(src, dest); err != nil {
					return nil, fmt.Errorf("failed to sync %s to %s: %w", file, dest, err)
				}
			}
		}
	}
	return results, nil
}

// SyncAll applies every rule in cfg to its own location list.
func SyncAll(root string, cfg *config.Config, dryRun bool) ([]Result, error) {
	var all []Result
	for _, rule := range cfg.Rules {
		results, err := SyncRule(root, rule, rule.Locations, dryRun)
		if err != nil {
			return nil, err
		}
		all = append(all, results...)
	}
	return all, nil
}

// ResolvePath expands ~/ and resolves relative paths against root.
func ResolvePath(root, path string) string {
	if strings.HasPrefix(path, "~/") {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, path[2:])
	}
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(root, path)
}

func symlink(src, dest string) error {
	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return err
	}
	// Remove existing file/symlink/dir at dest
	os.Remove(dest)
	return os.Symlink(src, dest)
}

func copyAny(src, dest string) error {
	info, err := os.Stat(src)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return copyDir(src, dest)
	}
	return CopyFile(src, dest)
}

func copyDir(src, dest string) error {
	if err := os.MkdirAll(dest, 0755); err != nil {
		return err
	}
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	for _, e := range entries {
		if err := copyAny(filepath.Join(src, e.Name()), filepath.Join(dest, e.Name())); err != nil {
			return err
		}
	}
	return nil
}

func CopyFile(src, dest string) error {
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
