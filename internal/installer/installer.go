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
	Changed bool
}

// MatchesPlatform reports whether a rule applies for the given platform tag.
// An empty platform (flag not set) or a rule with no platforms always matches.
func MatchesPlatform(rule config.Rule, platform string) bool {
	if platform == "" || len(rule.Platforms) == 0 {
		return true
	}
	for _, p := range rule.Platforms {
		if p == platform {
			return true
		}
	}
	return false
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
	if rule.Archive {
		return syncArchiveRule(root, rule, locations, dryRun, force)
	}

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

			changed := true
			if !force && !IsURL(file) && rule.Repo == "" {
				src := ResolvePath(fileRoot, file)
				if rule.Link {
					changed = !symlinkUpToDate(src, dest)
				} else {
					changed = !upToDate(src, dest)
				}
			}

			if !dryRun && !changed {
				continue
			}

			results = append(results, Result{Package: loc, File: file, Dest: dest, Changed: changed})
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

func syncArchiveRule(root string, rule config.Rule, locations []string, dryRun, force bool) ([]Result, error) {
	var results []Result
	for _, file := range rule.Files {
		var cacheDir string
		if !dryRun {
			var err error
			cacheDir, err = EnsureArchive(file, rule.Extract, force)
			if err != nil {
				return nil, fmt.Errorf("failed to extract %s: %w", file, err)
			}
		}
		for _, extract := range rule.Extract {
			name := filepath.Base(extract)
			for _, loc := range locations {
				destDir := ResolvePath(root, loc)
				if rule.Dest != "" {
					destDir = filepath.Join(destDir, rule.Dest)
				}
				dest := filepath.Join(destDir, name)

				changed := true
				if !dryRun && !force {
					src := filepath.Join(cacheDir, name)
					changed = !upToDate(src, dest)
				}

				results = append(results, Result{Package: loc, File: extract, Dest: dest, Changed: changed})
				if dryRun || !changed {
					continue
				}

				src := filepath.Join(cacheDir, name)
				if err := CopyFile(src, dest); err != nil {
					return nil, fmt.Errorf("failed to copy %s to %s: %w", name, dest, err)
				}
			}
		}
	}
	return results, nil
}

// SyncAll applies every rule in cfg to its own location list.
// If platform is non-empty, rules with a platforms list that excludes it are skipped.
func SyncAll(root string, cfg *config.Config, platform string, dryRun, force bool) ([]Result, error) {
	var all []Result
	for _, rule := range cfg.Rules {
		if !MatchesPlatform(rule, platform) {
			continue
		}
		results, err := SyncRule(root, rule, rule.Locations, dryRun, force)
		if err != nil {
			return nil, err
		}
		all = append(all, results...)
	}
	return all, nil
}

func symlinkUpToDate(src, dest string) bool {
	target, err := os.Readlink(dest)
	if err != nil {
		return false
	}
	return target == src
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

func mkdirForDest(dest string) error {
	return os.MkdirAll(filepath.Dir(dest), 0755)
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
