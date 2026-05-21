//go:build !windows

package installer

import "os"

func symlink(src, dest string) error {
	if err := mkdirForDest(dest); err != nil {
		return err
	}
	os.Remove(dest)
	return os.Symlink(src, dest)
}
