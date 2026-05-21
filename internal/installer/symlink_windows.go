//go:build windows

package installer

func symlink(src, dest string) error {
	return CopyFile(src, dest)
}
