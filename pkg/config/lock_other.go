//go:build !windows

package config

import (
	"fmt"
	"os"
	"time"

	"golang.org/x/sys/unix"
)

// lockHandle takes an exclusive flock, polling in non-blocking mode so we can honor the
// timeout (a bare LOCK_EX would block uninterruptibly). Only EWOULDBLOCK means contention
// from another holder; any other error is a real failure and is returned immediately
// instead of being retried until timeout.
func lockHandle(f *os.File, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for {
		err := unix.Flock(int(f.Fd()), unix.LOCK_EX|unix.LOCK_NB)
		if err == nil {
			return nil
		}
		if err != unix.EWOULDBLOCK {
			return fmt.Errorf("unable to lock config file: %w", err)
		}
		if time.Now().After(deadline) {
			return errConfigLockContended
		}
		time.Sleep(10 * time.Millisecond)
	}
}

// unlockHandle releases the flock taken by lockHandle.
func unlockHandle(f *os.File) error {
	return unix.Flock(int(f.Fd()), unix.LOCK_UN)
}
