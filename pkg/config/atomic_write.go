package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// configFilePerm is the mode config files are written with: owner read/write
// only, since they hold credentials.
const configFilePerm os.FileMode = 0600

// maxSymlinkHops bounds symlink chain resolution so a cyclic link can't loop forever,
// mirroring the OS's own ELOOP guard.
const maxSymlinkHops = 40

// resolveConfigTarget follows a symlink at path all the way to the real file it points
// to, so an atomic rename replaces that target rather than a link in the chain.
// os.WriteFile (the pre-atomic write) followed symlinks this way, including chains that
// end at a not-yet-existing target, so a user-managed config.json symlink keeps working.
// A non-symlink or absent path is returned unchanged (the write creates/replaces it).
func resolveConfigTarget(path string) (string, error) {
	for hop := 0; hop < maxSymlinkHops; hop++ {
		info, err := os.Lstat(path)
		if err != nil || info.Mode()&os.ModeSymlink == 0 {
			return path, nil // absent (fresh file), unstattable, or a regular file: write here
		}
		target, err := os.Readlink(path)
		if err != nil {
			return "", fmt.Errorf("unable to resolve config symlink %s: %w", path, err)
		}
		if !filepath.IsAbs(target) {
			target = filepath.Join(filepath.Dir(path), target)
		}
		path = target
	}
	return "", fmt.Errorf("config symlink chain too deep (over %d hops)", maxSymlinkHops)
}

// writeFileAtomic writes data to path via a temp file in the SAME directory,
// then renames it into place. Same-directory is required: a temp on another
// filesystem breaks rename with EXDEV. On POSIX the parent directory is fsynced
// after the rename so a crash can't lose the rename even when the bytes are
// durable; on Windows the rename itself is the durability boundary. A symlinked
// path is followed to its target so the rename replaces the target, not the link.
func writeFileAtomic(path string, data []byte) error {
	target, err := resolveConfigTarget(path)
	if err != nil {
		return err
	}
	dir := filepath.Dir(target)

	tmp, err := os.CreateTemp(dir, filepath.Base(target)+".tmp-*")
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

	if err := renameReplace(tmpName, target); err != nil {
		return fmt.Errorf("unable to rename config file into place: %w", err)
	}

	return fsyncDir(dir)
}
