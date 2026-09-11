package config

import (
	"fmt"
	"os"
	"time"

	"github.com/confluentinc/cli/v4/pkg/errors"
)

// lockTimeout bounds how long a writer waits for the sidecar lock. It must
// dwarf the sub-millisecond read-modify-write it guards, so only genuine
// contention, never the operation itself, can time out.
const lockTimeout = 10 * time.Second

// fileLock guards config writes with a sidecar `<config>.lock` file. We lock a
// dedicated sidecar, never the data file: the atomic rename in writeFileAtomic
// swaps a new inode into the data path, so a flock held on the data file would
// stop excluding anyone. The sidecar is created once and never renamed or
// unlinked. flock (POSIX) and LockFileEx (Windows) both release on handle close
// or process death, so there is deliberately no stale-lock recovery.
type fileLock struct {
	path string
	f    *os.File
}

// newFileLock builds a lock for the sidecar file next to configPath. It does not open or
// acquire anything yet; call lock to do that.
func newFileLock(configPath string) *fileLock {
	return &fileLock{path: configPath + ".lock"}
}

// lock opens (creating if needed) the sidecar file and blocks until it acquires an exclusive
// lock on it or timeout elapses, whichever comes first.
func (l *fileLock) lock(timeout time.Duration) error {
	f, err := os.OpenFile(l.path, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return fmt.Errorf("unable to open config lock file %s: %w", l.path, err)
	}
	l.f = f

	if err := lockHandle(f, timeout); err != nil {
		f.Close()
		l.f = nil
		return err
	}
	return nil
}

// unlock releases the lock and closes the underlying file handle. It is a no-op if the lock
// was never acquired.
func (l *fileLock) unlock() error {
	if l.f == nil {
		return nil
	}
	err := unlockHandle(l.f)
	closeErr := l.f.Close()
	l.f = nil
	if err != nil {
		return err
	}
	return closeErr
}

var errConfigLockContended = errors.NewErrorWithSuggestions(
	"another `confluent` process is updating the configuration file",
	"Wait for the other command to finish, then retry.",
)
