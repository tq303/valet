package installer

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/tq303/valet/internal/config"
)

type Result struct {
	Package string
	File    string
	Dest    string
}

func IsURL(s string) bool {
	return strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://")
}


// FileName returns the base filename for a local path or URL.
func FileName(file string) string {
	if IsURL(file) {
		u, _ := url.Parse(file)
		return path.Base(u.Path)
	}
	return filepath.Base(file)
}

// SyncRule syncs all files in a rule into the given locations.
func SyncRule(root string, rule config.Rule, locations []string, dryRun, force bool) ([]Result, error) {
	fileRoot := root
	if rule.Repo != "" && !dryRun {
		cacheDir, err := EnsureRepo(rule.Repo)
		if err != nil {
			return nil, fmt.Errorf("failed to sync repo %s: %w", rule.Repo, err)
		}
		fileRoot = cacheDir
	}

	var results []Result
	for _, file := range rule.Files {
		name := FileName(file)

		if !IsURL(file) && !dryRun && rule.Repo == "" {
			src := ResolvePath(fileRoot, file)
			if _, err := os.Stat(src); err != nil {
				if err := touchFile(src); err != nil {
					return nil, fmt.Errorf("could not create %s: %w", file, err)
				}
			}
		}

		for _, loc := range locations {
			destDir := ResolvePath(root, loc)
			if rule.Dest != "" {
				destDir = filepath.Join(destDir, rule.Dest)
			}
			dest := filepath.Join(destDir, name)

			if !dryRun && !force && !IsURL(file) && !rule.Link {
				src := ResolvePath(fileRoot, file)
				if upToDate(src, dest) {
					continue
				}
			}

			results = append(results, Result{Package: loc, File: file, Dest: dest})
			if dryRun {
				continue
			}
			if IsURL(file) {
				if err := writeFromURL(file, dest); err != nil {
					return nil, fmt.Errorf("failed to fetch %s: %w", file, err)
				}
			} else if rule.Link {
				src := ResolvePath(fileRoot, file)
				if err := symlink(src, dest); err != nil {
					return nil, fmt.Errorf("failed to symlink %s to %s: %w", file, dest, err)
				}
			} else {
				src := ResolvePath(fileRoot, file)
				if err := copyAny(src, dest); err != nil {
					return nil, fmt.Errorf("failed to sync %s to %s: %w", file, dest, err)
				}
			}
		}
	}
	return results, nil
}

// SyncAll applies every rule in cfg to its own location list.
func SyncAll(root string, cfg *config.Config, dryRun, force bool) ([]Result, error) {
	var all []Result
	for _, rule := range cfg.Rules {
		results, err := SyncRule(root, rule, rule.Locations, dryRun, force)
		if err != nil {
			return nil, err
		}
		all = append(all, results...)
	}
	return all, nil
}

func upToDate(src, dest string) bool {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return false
	}
	if srcInfo.IsDir() {
		return false
	}
	a, err := os.ReadFile(src)
	if err != nil {
		return false
	}
	b, err := os.ReadFile(dest)
	if err != nil {
		return false
	}
	return bytes.Equal(a, b)
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

func writeFromURL(rawURL, dest string) error {
	resp, err := http.Get(rawURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d fetching %s", resp.StatusCode, rawURL)
	}
	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return err
	}
	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, resp.Body)
	return err
}

func touchFile(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	return f.Close()
}

func symlink(src, dest string) error {
	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return err
	}
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
	if info, err := os.Stat(dest); err == nil && !info.IsDir() {
		os.Remove(dest)
	}
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
