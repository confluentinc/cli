package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// configFilePerm is the mode config files are written with: owner read/write
// only, since they hold credentials.
const configFilePerm os.FileMode = 0600

// writeFileAtomic writes data to path via a temp file in the SAME directory,
// then renames it into place. Same-directory is required: a temp on another
// filesystem breaks rename with EXDEV. On POSIX the parent directory is fsynced
// after the rename so a crash can't lose the rename even when the bytes are
// durable; on Windows the rename itself is the durability boundary.
func writeFileAtomic(path string, data []byte) error {
	dir := filepath.Dir(path)

	tmp, err := os.CreateTemp(dir, filepath.Base(path)+".tmp-*")
	if err != nil {
		return fmt.Errorf("unable to create temp config file: %w", err)
	}
	tmpName := tmp.Name()
	// Best-effort cleanup if we bail before the rename; a successful rename
	// consumes the temp so the remove becomes a no-op.
	defer os.Remove(tmpName)

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return fmt.Errorf("unable to write temp config file: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return fmt.Errorf("unable to fsync temp config file: %w", err)
	}
	if err := tmp.Chmod(configFilePerm); err != nil {
		tmp.Close()
		return fmt.Errorf("unable to set temp config file permissions: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("unable to close temp config file: %w", err)
	}

	if err := renameReplace(tmpName, path); err != nil {
		return fmt.Errorf("unable to rename config file into place: %w", err)
	}

	return fsyncDir(dir)
}
