//go:build windows

package config

import (
	"os"

	"golang.org/x/sys/windows"
)

// tryLockHandle makes one non-blocking attempt at an exclusive lock. It reports
// whether the lock was acquired; ERROR_LOCK_VIOLATION means another holder has
// it (not acquired, no error), while any other error is a real failure. The
// shared acquireWithTimeout loop owns the polling and timeout (a blocking
// LockFileEx would block uninterruptibly and ignore the timeout).
func tryLockHandle(f *os.File) (bool, error) {
	var overlapped windows.Overlapped
	// LOCKFILE_FAIL_IMMEDIATELY makes the call non-blocking so the shared loop can poll.
	err := windows.LockFileEx(windows.Handle(f.Fd()), windows.LOCKFILE_EXCLUSIVE_LOCK|windows.LOCKFILE_FAIL_IMMEDIATELY, 0, 1, 0, &overlapped)
	if err == nil {
		return true, nil
	}
	if err == windows.ERROR_LOCK_VIOLATION {
		return false, nil
	}
	return false, err
}

// unlockHandle releases the lock taken by tryLockHandle.
func unlockHandle(f *os.File) error {
	var overlapped windows.Overlapped
	return windows.UnlockFileEx(windows.Handle(f.Fd()), 0, 1, 0, &overlapped)
}
