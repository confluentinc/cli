//go:build windows

package config

import (
	"os"
	"time"

	"golang.org/x/sys/windows"
)

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
		if time.Now().After(deadline) {
			return errConfigLockContended
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func unlockHandle(f *os.File) error {
	var overlapped windows.Overlapped
	return windows.UnlockFileEx(windows.Handle(f.Fd()), 0, 1, 0, &overlapped)
}
