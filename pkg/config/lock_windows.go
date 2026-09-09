//go:build windows

package config

import (
	"fmt"
	"os"
	"time"

	"golang.org/x/sys/windows"
)

// lockHandle takes an exclusive lock, polling in non-blocking mode so we can honor the
// timeout (a blocking LockFileEx would block uninterruptibly). Only ERROR_LOCK_VIOLATION
// means contention from another holder; any other error is a real failure and is returned
// immediately instead of being retried until timeout.
func lockHandle(f *os.File, timeout time.Duration) error {
	handle := windows.Handle(f.Fd())
	deadline := time.Now().Add(timeout)
	var overlapped windows.Overlapped
	for {
		// LOCKFILE_FAIL_IMMEDIATELY makes the call non-blocking so we can poll to a deadline.
		err := windows.LockFileEx(handle, windows.LOCKFILE_EXCLUSIVE_LOCK|windows.LOCKFILE_FAIL_IMMEDIATELY, 0, 1, 0, &overlapped)
		if err == nil {
			return nil
		}
		if err != windows.ERROR_LOCK_VIOLATION {
			return fmt.Errorf("unable to lock config file: %w", err)
		}
		if time.Now().After(deadline) {
			return errConfigLockContended
		}
		time.Sleep(10 * time.Millisecond)
	}
}

// unlockHandle releases the lock taken by lockHandle.
func unlockHandle(f *os.File) error {
	var overlapped windows.Overlapped
	return windows.UnlockFileEx(windows.Handle(f.Fd()), 0, 1, 0, &overlapped)
}
