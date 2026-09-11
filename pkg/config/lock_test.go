package config

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestFileLock_SerializesTwoHolders(t *testing.T) {
	cfgPath := filepath.Join(t.TempDir(), "config.json")

	l1 := newFileLock(cfgPath)
	require.NoError(t, l1.lock(lockTimeout))

	// The sidecar sits next to the data file and is created, not the data file.
	_, err := os.Stat(cfgPath + ".lock")
	require.NoError(t, err)

	var held atomic.Bool
	held.Store(true)
	acquired := make(chan struct{})

	go func() {
		l2 := newFileLock(cfgPath)
		require.NoError(t, l2.lock(lockTimeout))
		require.False(t, held.Load(), "second holder acquired while first still held the lock")
		require.NoError(t, l2.unlock())
		close(acquired)
	}()

	time.Sleep(100 * time.Millisecond) // give the goroutine time to block on lock()
	held.Store(false)
	require.NoError(t, l1.unlock())

	select {
	case <-acquired:
	case <-time.After(lockTimeout):
		t.Fatal("second holder never acquired the lock after release")
	}
}

func TestFileLock_TimesOut(t *testing.T) {
	cfgPath := filepath.Join(t.TempDir(), "config.json")

	l1 := newFileLock(cfgPath)
	require.NoError(t, l1.lock(lockTimeout))
	defer func() { _ = l1.unlock() }()

	l2 := newFileLock(cfgPath)
	err := l2.lock(50 * time.Millisecond)

	require.Error(t, err)
}

// A writer that waits past contentionWarnDelay for a held lock prints a one-time
// note, so a user is not left staring at a silent hang. The uncontended fast path
// (which every other test exercises) must stay quiet.
func TestFileLock_WarnsOnlyAfterSustainedContention(t *testing.T) {
	cfgPath := filepath.Join(t.TempDir(), "config.json")
	const waitingNote = "Waiting for another confluent process"

	// Contended: the first holder keeps the lock while the second waits past the
	// warn delay before timing out, so the note must appear exactly once.
	l1 := newFileLock(cfgPath)
	require.NoError(t, l1.lock(lockTimeout))

	contendedStderr := captureStderr(t, func() {
		l2 := newFileLock(cfgPath)
		require.Error(t, l2.lock(contentionWarnDelay+500*time.Millisecond))
	})
	require.NoError(t, l1.unlock())

	require.Equal(t, 1, strings.Count(contendedStderr, waitingNote), "the waiting note must print exactly once under sustained contention")

	// Uncontended: the lock is free, so acquisition is immediate and silent.
	uncontendedStderr := captureStderr(t, func() {
		l3 := newFileLock(cfgPath)
		require.NoError(t, l3.lock(lockTimeout))
		require.NoError(t, l3.unlock())
	})

	require.NotContains(t, uncontendedStderr, waitingNote, "the fast uncontended path must not print the waiting note")
}

// captureStderr redirects os.Stderr for the duration of fn and returns what was
// written, so a test can assert on output.ErrPrintln notices.
func captureStderr(t *testing.T, fn func()) string {
	t.Helper()

	r, w, err := os.Pipe()
	require.NoError(t, err)
	orig := os.Stderr
	os.Stderr = w
	defer func() { os.Stderr = orig }()

	fn()

	require.NoError(t, w.Close())
	out, err := io.ReadAll(r)
	require.NoError(t, err)
	return string(out)
}
