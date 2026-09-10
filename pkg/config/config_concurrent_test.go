package config

import (
	"fmt"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

// Two sessions each add a DIFFERENT platform concurrently. Before the lock+merge
// work, the last writer's whole-file overwrite dropped the other's platform.
func TestSave_ConcurrentDifferentFields_NoLostWrite(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	seed := New()
	seed.Filename = path
	require.NoError(t, seed.Save())

	const n = 8
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			c, err := readConfigFromDisk(path, seed)
			require.NoError(t, err)
			c.snapshotBaseline()
			name := fmt.Sprintf("platform-%d", i)
			c.Platforms[name] = &Platform{Name: name}
			require.NoError(t, c.Save())
		}(i)
	}
	wg.Wait()

	final, err := readConfigFromDisk(path, seed)
	require.NoError(t, err)
	for i := 0; i < n; i++ {
		require.Contains(t, final.Platforms, fmt.Sprintf("platform-%d", i),
			"every concurrent session's platform must survive")
	}
}
