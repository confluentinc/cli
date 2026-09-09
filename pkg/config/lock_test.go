package config

import (
	"os"
	"path/filepath"
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
