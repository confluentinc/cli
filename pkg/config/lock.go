package config

import (
	"fmt"
	"os"
	"time"

	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/output"
)

// lockTimeout bounds how long a writer waits for the sidecar lock. It still
// dwarfs the sub-millisecond read-modify-write it guards, so only genuine
// contention can time out, but is short enough to fail fast rather than hang a
// user behind a stuck or long-running peer.
const lockTimeout = 3 * time.Second

// contentionWarnDelay is how long lock() polls unsuccessfully before printing a
// one-time note that another process holds the lock. It sits well under
// lockTimeout, so a genuinely contended wait is explained while the
// sub-millisecond common case stays silent.
const contentionWarnDelay = 1 * time.Second

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

	if err := acquireWithTimeout(f, timeout); err != nil {
		f.Close()
		l.f = nil
		return err
	}
	return nil
}

// acquireWithTimeout polls tryLockHandle (the per-OS non-blocking single
// attempt) until it acquires the lock or timeout elapses. A blocking OS lock
// would ignore the timeout, so we spin at a fixed interval. After
// contentionWarnDelay of failed attempts it prints a one-time stderr note, so a
// user genuinely waiting on another confluent process is not left staring at a
// silent hang; the fast common case acquires on the first try and stays quiet.
func acquireWithTimeout(f *os.File, timeout time.Duration) error {
	start := time.Now()
	deadline := start.Add(timeout)
	warned := false
	for {
		acquired, err := tryLockHandle(f)
		if err != nil {
			return fmt.Errorf("unable to lock config file: %w", err)
		}
		if acquired {
			return nil
		}
		if !warned && time.Since(start) >= contentionWarnDelay {
			output.ErrPrintln(false, "Waiting for another confluent process to finish updating the configuration file...")
			warned = true
		}
		if time.Now().After(deadline) {
			return errConfigLockContended
		}
		time.Sleep(10 * time.Millisecond)
	}
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
