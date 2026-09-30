package test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/confluentinc/cli/v4/pkg/utils"
)

// legacyConfigDirName and legacyConfigFileName mirror pkg/config's legacyConfigPath: v4 only
// ever wrote the stable path, so every channel reads its seed file from the same place.
const (
	legacyConfigDirName  = ".confluent"
	legacyConfigFileName = "config.json"
	legacyBackupFileName = "config.json.v4-backup"

	devStateDirName = ".confluent-dev"

	settingsFileName = "settings.json"
	contextsFileName = "contexts.json"
	secretsFileName  = "secrets.json"
	cacheDirName     = ".cache"

	configMigrationFixtureDir  = "config-migration"
	configMigrationFixtureFile = "legacy_config.json"

	configMigrationStableGolden  = "config-migration/stable-announcement.golden"
	configMigrationDevSeedGolden = "config-migration/dev-seed-announcement.golden"
	configMigrationWarningGolden = "config-migration/downgrade-warning.golden"
	configMigrationTableGolden   = "config-migration/context-list.golden"

	configMigrationListArgs = "context list"
)

// TestConfigMigrationDevSeed covers the non-Stable path: the shared (unstamped) test binary runs
// on the Dev channel, so a v4 config.json at HOME seeds the Dev channel's own state directory
// read-only, leaving the legacy file untouched. Uses a standalone flow (not CLITest/runIntegrationTest)
// because resetConfiguration writes stores into the suite's shared temp HOME before every non-workflow
// test, which would satisfy the migration guard before this test's own HOME override ever runs.
func (s *CLITestSuite) TestConfigMigrationDevSeed() {
	t := s.T()
	home := t.TempDir()
	env := []string{"HOME=" + home, "USERPROFILE=" + home}

	legacyPath := filepath.Join(home, legacyConfigDirName, legacyConfigFileName)
	legacyBytes := seedLegacyConfig(t, home)

	firstRun := normalizeHome(runCommand(t, testBin, env, configMigrationListArgs, 0, ""), home)
	compareOrUpdateGolden(t, configMigrationDevSeedGolden, firstRun)

	requireBytesUnchanged(t, legacyPath, legacyBytes)
	requireNoPath(t, filepath.Join(home, legacyConfigDirName, legacyBackupFileName))
	requireMigratedStores(t, filepath.Join(home, devStateDirName), legacyBytes)

	secondRun := normalizeHome(runCommand(t, testBin, env, configMigrationListArgs, 0, ""), home)
	compareOrUpdateGolden(t, configMigrationTableGolden, secondRun)
}

// TestConfigMigrationStable covers the Stable path end to end: the migration announcement, the
// legacy file staying frozen in place, and the once-per-change downgrade warning. It builds
// exactly one stamped "9.9.9" binary (the Windows CI agent has shown go build exe-lock flakes) and
// runs every step against it.
func TestConfigMigrationStable(t *testing.T) {
	home := t.TempDir()
	binary := buildStampedCli(t, "9.9.9")

	legacyPath := filepath.Join(home, legacyConfigDirName, legacyConfigFileName)
	legacyBytes := seedLegacyConfig(t, home)

	firstRun := normalizeHome(runStampedCliCombinedOutput(t, binary, home, "context", "list"), home)
	compareOrUpdateGolden(t, configMigrationStableGolden, firstRun)

	requireBytesUnchanged(t, legacyPath, legacyBytes)
	requireMigratedStores(t, filepath.Join(home, legacyConfigDirName), legacyBytes)

	secondRun := normalizeHome(runStampedCliCombinedOutput(t, binary, home, "context", "list"), home)
	compareOrUpdateGolden(t, configMigrationTableGolden, secondRun)

	// simulate a pre-v5 install writing to the frozen legacy file: content and mtime both
	// change, since warnIfLegacyFileChanged compares the stat stamp, not the file's contents.
	rewritten := []byte("this file was rewritten by a v4 install\n")
	require.NoError(t, os.WriteFile(legacyPath, rewritten, 0600))
	future := time.Now().Add(time.Hour)
	require.NoError(t, os.Chtimes(legacyPath, future, future))

	warningRun := normalizeHome(runStampedCliCombinedOutput(t, binary, home, "context", "list"), home)
	compareOrUpdateGolden(t, configMigrationWarningGolden, warningRun)

	silentRun := normalizeHome(runStampedCliCombinedOutput(t, binary, home, "context", "list"), home)
	compareOrUpdateGolden(t, configMigrationTableGolden, silentRun)
}

// seedLegacyConfig copies the config-migration fixture into <home>/.confluent/config.json,
// mirroring what a v4 install would have left behind, and returns the seeded bytes.
func seedLegacyConfig(t *testing.T, home string) []byte {
	t.Helper()

	data, err := os.ReadFile(getInputFixturePath(configMigrationFixtureDir, configMigrationFixtureFile))
	require.NoError(t, err)

	dir := filepath.Join(home, legacyConfigDirName)
	require.NoError(t, os.MkdirAll(dir, 0700))
	require.NoError(t, os.WriteFile(filepath.Join(dir, legacyConfigFileName), data, 0600))

	return data
}

// normalizeHome replaces the temp HOME with a stable token and converts path separators to
// forward slashes, so a migration golden embedding a store path is portable to Windows.
func normalizeHome(output, home string) string {
	return strings.ReplaceAll(filepath.ToSlash(output), filepath.ToSlash(home), "<HOME>")
}

// compareOrUpdateGolden checks actual against fixture, or (with -update) writes actual as the new
// golden content; mirrors CLITestSuite.validateTestOutput's -update handling for tests that
// bypass the CLITest struct.
func compareOrUpdateGolden(t *testing.T, fixture, actual string) {
	t.Helper()

	if *update {
		writeFixture(t, fixture, actual)
		return
	}
	require.Equal(t, utils.NormalizeNewLines(LoadFixture(t, fixture)), utils.NormalizeNewLines(actual))
}

// requireBytesUnchanged asserts the file at path still holds exactly want, i.e. migration never
// wrote, renamed, or otherwise touched the legacy file it read from.
func requireBytesUnchanged(t *testing.T, path string, want []byte) {
	t.Helper()

	got, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Equal(t, want, got)
}

// requireNoPath asserts nothing exists at path.
func requireNoPath(t *testing.T, path string) {
	t.Helper()

	_, err := os.Stat(path)
	require.True(t, os.IsNotExist(err), "expected %s to not exist", path)
}

// requireMigratedStores asserts the running channel's own state directory holds every migrated
// store plus a byte-identical backup of the legacy file, the migration's commit marker.
func requireMigratedStores(t *testing.T, stateDir string, legacyBytes []byte) {
	t.Helper()

	for _, name := range []string{settingsFileName, contextsFileName, secretsFileName} {
		require.FileExists(t, filepath.Join(stateDir, name))
	}
	require.DirExists(t, filepath.Join(stateDir, cacheDirName))
	requireBytesUnchanged(t, filepath.Join(stateDir, legacyBackupFileName), legacyBytes)
}

// runStampedCliCombinedOutput is runStampedCli's sibling that returns output instead of
// discarding it, for assertions that need to see the migration announcement or warning.
func runStampedCliCombinedOutput(t *testing.T, binary, home string, args ...string) string {
	t.Helper()

	cmd := exec.Command(binary, args...)
	cmd.Env = append(os.Environ(), "HOME="+home, "USERPROFILE="+home)

	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "%s %v failed: %s", binary, args, output)

	return string(output)
}
