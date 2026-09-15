//go:build windows

package config

import (
	"os"
	"time"
)

// renameReplace retries the rename briefly: on Windows a third-party handle
// holder (antivirus, indexer, backup agent) transiently holding the destination
// open makes MoveFileEx(REPLACE_EXISTING) fail. This is unrelated to our lock,
// which is on the sidecar, not the data file.
func renameReplace(oldpath, newpath string) error {
	var err error
	for _, backoff := range []time.Duration{0, 20 * time.Millisecond, 50 * time.Millisecond, 100 * time.Millisecond, 200 * time.Millisecond} {
		if backoff > 0 {
			time.Sleep(backoff)
		}
		if err = os.Rename(oldpath, newpath); err == nil {
			return nil
		}
	}
	return err
}

// fsyncDir is a no-op on Windows: there is no directory fsync, and the NTFS
// rename is itself the durability boundary.
func fsyncDir(_ string) error { return nil }
