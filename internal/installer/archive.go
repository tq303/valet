package installer

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/bzip2"
	"compress/gzip"
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/ulikunitz/xz"
)

const archiveCacheBase = "/tmp/valet/archives"

func IsArchiveURL(u string) bool {
	lower := strings.ToLower(u)
	return strings.HasSuffix(lower, ".tar.gz") ||
		strings.HasSuffix(lower, ".tar.xz") ||
		strings.HasSuffix(lower, ".tar.bz2") ||
		strings.HasSuffix(lower, ".tgz") ||
		strings.HasSuffix(lower, ".zip")
}

// EnsureArchive downloads and extracts the archive at rawURL into a cache dir,
// returning the cache dir path. Re-extracts if force is true.
func EnsureArchive(rawURL string, extractPaths []string, force bool) (string, error) {
	hash := fmt.Sprintf("%x", sha256.Sum256([]byte(rawURL)))
	cacheDir := filepath.Join(archiveCacheBase, hash)

	if !force {
		if _, err := os.Stat(cacheDir); err == nil {
			return cacheDir, nil
		}
	}

	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return "", err
	}

	resp, err := http.Get(rawURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP %d fetching %s", resp.StatusCode, rawURL)
	}

	lower := strings.ToLower(rawURL)
	if strings.HasSuffix(lower, ".zip") {
		if err := extractZip(resp.Body, extractPaths, cacheDir); err != nil {
			return "", err
		}
	} else {
		if err := extractTar(lower, resp.Body, extractPaths, cacheDir); err != nil {
			return "", err
		}
	}

	return cacheDir, nil
}

func extractTar(lowerURL string, r io.Reader, extractPaths []string, destDir string) error {
	var tarReader io.Reader
	switch {
	case strings.HasSuffix(lowerURL, ".tar.gz") || strings.HasSuffix(lowerURL, ".tgz"):
		gr, err := gzip.NewReader(r)
		if err != nil {
			return err
		}
		defer gr.Close()
		tarReader = gr
	case strings.HasSuffix(lowerURL, ".tar.xz"):
		xzr, err := xz.NewReader(r)
		if err != nil {
			return err
		}
		tarReader = xzr
	case strings.HasSuffix(lowerURL, ".tar.bz2"):
		tarReader = bzip2.NewReader(r)
	default:
		tarReader = r
	}

	want := buildWantSet(extractPaths)
	tr := tar.NewReader(tarReader)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if hdr.Typeflag == tar.TypeDir {
			continue
		}
		if !matchesWant(hdr.Name, want) {
			continue
		}
		dest := filepath.Join(destDir, filepath.Base(hdr.Name))
		out, err := os.OpenFile(dest, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(hdr.Mode)|0100)
		if err != nil {
			return err
		}
		_, copyErr := io.Copy(out, tr)
		out.Close()
		if copyErr != nil {
			return copyErr
		}
	}
	return nil
}

func extractZip(r io.Reader, extractPaths []string, destDir string) error {
	data, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return err
	}
	want := buildWantSet(extractPaths)
	for _, f := range zr.File {
		if f.FileInfo().IsDir() || !matchesWant(f.Name, want) {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return err
		}
		dest := filepath.Join(destDir, filepath.Base(f.Name))
		out, err := os.OpenFile(dest, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, f.Mode()|0100)
		if err != nil {
			rc.Close()
			return err
		}
		_, copyErr := io.Copy(out, rc)
		rc.Close()
		out.Close()
		if copyErr != nil {
			return copyErr
		}
	}
	return nil
}

// buildWantSet builds a lookup of all target basenames (and exact paths) to match against archive entries.
func buildWantSet(extractPaths []string) map[string]bool {
	want := make(map[string]bool, len(extractPaths)*2)
	for _, p := range extractPaths {
		want[p] = true
		want[filepath.Base(p)] = true
	}
	return want
}

// matchesWant returns true if the archive entry name matches any wanted path or basename,
// including after stripping a leading directory component (common in release tarballs).
func matchesWant(name string, want map[string]bool) bool {
	base := filepath.Base(name)
	if want[name] || want[base] {
		return true
	}
	// strip one leading path component (e.g. "gifski-1.34.0/gifski" -> "gifski")
	if i := strings.Index(name, "/"); i >= 0 {
		stripped := name[i+1:]
		if want[stripped] || want[filepath.Base(stripped)] {
			return true
		}
	}
	return false
}
