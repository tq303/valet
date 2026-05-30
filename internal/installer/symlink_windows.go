//go:build windows

package installer

import "os"

func symlink(src, dest string) error {
	if err := mkdirForDest(dest); err != nil {
		return err
	}
	os.Remove(dest)
	if err := os.Symlink(src, dest); err != nil {
		return CopyFile(src, dest)
	}
	return nil
}
