package installer

import (
	"crypto/sha256"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var gitHosts = map[string]bool{
	"github.com":    true,
	"gitlab.com":    true,
	"bitbucket.org": true,
}

func IsGitRepo(s string) bool {
	if !IsURL(s) {
		return false
	}
	if strings.HasSuffix(s, ".git") {
		return true
	}
	u, err := url.Parse(s)
	if err != nil {
		return false
	}
	if !gitHosts[u.Host] {
		return false
	}
	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	return len(parts) == 2 && parts[0] != "" && parts[1] != ""
}

func RepoCacheDir(repoURL string) string {
	h := sha256.Sum256([]byte(repoURL))
	return filepath.Join(os.TempDir(), "valet", "repos", fmt.Sprintf("%x", h[:8]))
}

func EnsureRepo(repoURL string) (string, error) {
	dir := RepoCacheDir(repoURL)
	if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
		cmd := exec.Command("git", "-C", dir, "pull", "--ff-only", "--quiet")
		cmd.Stderr = os.Stderr
		return dir, cmd.Run()
	}
	if err := os.MkdirAll(filepath.Dir(dir), 0755); err != nil {
		return "", err
	}
	cmd := exec.Command("git", "clone", "--depth=1", "--quiet", repoURL, dir)
	cmd.Stderr = os.Stderr
	return dir, cmd.Run()
}
