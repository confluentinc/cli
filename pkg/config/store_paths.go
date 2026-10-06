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

// SecretsFilename is the encrypted secret store, kept out of the plaintext config file.
func SecretsFilename() string { return stateDirPath("secrets.json") }

// SettingsFilename holds preferences and the platforms every context references.
func SettingsFilename() string { return stateDirPath("settings.json") }

// ContextsFilename is the secret-free context selection state.
func ContextsFilename() string { return stateDirPath("contexts.json") }
