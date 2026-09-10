//go:build !windows

package config

import (
	"fmt"
	"os"
)

// renameReplace atomically replaces newpath with oldpath via POSIX rename(2).
func renameReplace(oldpath, newpath string) error {
	return os.Rename(oldpath, newpath)
}

// fsyncDir flushes the directory entry so the rename survives a crash.
func fsyncDir(dir string) error {
	d, err := os.Open(dir)
	if err != nil {
		return fmt.Errorf("unable to open config directory for fsync: %w", err)
	}
	defer d.Close()
	if err := d.Sync(); err != nil {
		return fmt.Errorf("unable to fsync config directory: %w", err)
	}
	return nil
}
