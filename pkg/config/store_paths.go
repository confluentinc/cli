package config

import (
	"os"
	"path/filepath"
)

// stateDirPath joins elems under the channel-aware state directory (e.g.
// ~/.confluent or ~/.confluent-prerelease). A missing home yields a relative
// path, mirroring GetDefaultFilename's tolerance.
func stateDirPath(elem ...string) string {
	home, _ := os.UserHomeDir()
	return filepath.Join(append([]string{home, StateDirName()}, elem...)...)
}

// CacheDir is the disposable cache directory. Nothing here is authoritative; a
// wiped CacheDir degrades to a refetch, never an error.
func CacheDir() string { return stateDirPath(".cache") }
