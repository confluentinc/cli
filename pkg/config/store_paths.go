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

// legacyBackupName and legacyMigratingName sit in a channel's state directory; the backup there
// means the channel has migrated, and the marker means a migration started and hasn't finished.
const (
	legacyBackupName    = "config.json.v4-backup"
	legacyMigratingName = "config.json.v4-migrating"
)

// legacyConfigFilename is v4's config.json, or ("", false) when the home directory is unknown.
func legacyConfigFilename() (string, bool) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", false
	}
	return filepath.Join(home, ".confluent", "config.json"), true
}

// legacyBackupFilename is the running channel's byte-identical copy of the migrated v4 file.
func legacyBackupFilename() string { return stateDirPath(legacyBackupName) }

// legacyMigratingFilename marks a migration that has started writing stores but not the backup.
func legacyMigratingFilename() string { return stateDirPath(legacyMigratingName) }
