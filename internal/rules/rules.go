package rules

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type ScannedRule struct {
	Name  string
	Path  string
	Tools []string // pre-assigned from folder; nil means prompt the user
}

// ScanFiles scans the rules directory and returns all non-empty .md files.
// Files in .claude/ or .cursor/ subdirectories are pre-assigned to their tool.
// Files at the root of dir have no pre-assigned tools and should be prompted.
func ScanFiles(dir string) ([]ScannedRule, error) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil, fmt.Errorf(".rules folder not found — run `val init` first")
	}

	var scanned []ScannedRule

	// Scan tool-specific subdirectories
	toolDirs := map[string]string{
		".claude": "claude",
		".cursor": "cursor",
	}
	for folder, tool := range toolDirs {
		subdir := filepath.Join(dir, folder)
		files, err := scanDir(subdir)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, err
		}
		for _, f := range files {
			scanned = append(scanned, ScannedRule{
				Name:  filepath.Join(folder, filepath.Base(f)),
				Path:  f,
				Tools: []string{tool},
			})
		}
	}

	// Scan root-level files (no pre-assigned tool)
	files, err := scanDir(dir)
	if err != nil {
		return nil, err
	}
	for _, f := range files {
		scanned = append(scanned, ScannedRule{
			Name:  filepath.Base(f),
			Path:  f,
			Tools: nil,
		})
	}

	return scanned, nil
}

func scanDir(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		info, err := e.Info()
		if err != nil {
			return nil, err
		}
		if info.Size() == 0 {
			return nil, fmt.Errorf("rule file %s is empty", e.Name())
		}
		files = append(files, filepath.Join(dir, e.Name()))
	}
	return files, nil
}
