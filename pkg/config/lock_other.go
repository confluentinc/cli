//go:build darwin || linux

package config

import (
	"os"

	"golang.org/x/sys/unix"
)

// tryLockHandle makes one non-blocking attempt at an exclusive flock. It reports
// whether the lock was acquired; EWOULDBLOCK means another holder has it (not
// acquired, no error), while any other error is a real failure. The shared
// acquireWithTimeout loop owns the polling and timeout (a bare LOCK_EX would
// block uninterruptibly and ignore the timeout).
func tryLockHandle(f *os.File) (bool, error) {
	err := unix.Flock(int(f.Fd()), unix.LOCK_EX|unix.LOCK_NB)
	if err == nil {
		return true, nil
	}
	if err == unix.EWOULDBLOCK {
		return false, nil
	}
	return false, err
}

// unlockHandle releases the flock taken by tryLockHandle.
func unlockHandle(f *os.File) error {
	return unix.Flock(int(f.Fd()), unix.LOCK_UN)
}
