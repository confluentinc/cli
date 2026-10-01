package config

import (
	"encoding/json"
	stderrors "errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/log"
	"github.com/confluentinc/cli/v4/pkg/output"
	pversion "github.com/confluentinc/cli/v4/pkg/version"
)

// migrationErrorMsg wraps a migration-time failure with the legacy file it came from, quoted the
// way errors.UnableToReadConfigurationFileErrorMsg quotes its path.
const migrationErrorMsg = `unable to migrate configuration file "%s": %w`

// migrationAnnouncementMsg is printed once, after a Stable migration seeds a config with at
// least one context.
const migrationAnnouncementMsg = `Confluent CLI moved your configuration to settings.json, contexts.json, secrets.json, and .cache/. Versions before v5 keep using the old config.json and won't see logins or context changes you make here, or vice versa.`

// migrationSeedMsg is printed once, after a non-Stable channel copies contexts and logins from
// the Stable install's legacy file. Args: legacyConfigFilename(), the running channel's StateDir().
const migrationSeedMsg = `Confluent CLI copied your contexts and logins from "%s" into this build's own configuration in "%s". Changes you make here won't affect that installation, or vice versa.`

// migrationSeedSkippedMsg is printed when a non-Stable channel can't decode the legacy file and
// starts fresh instead. Args: legacyConfigFilename(), the decode error.
const migrationSeedSkippedMsg = `Skipped copying contexts and logins from "%s": %v`

// migrationCheckErrorMsg reports a guard stat failure, which must never pass for "absent".
const migrationCheckErrorMsg = `unable to check "%s" for a pending configuration migration: %w`

// migrationMarkerErrorMsg and migrationBackupErrorMsg name the migration file that failed to write.
const (
	migrationMarkerErrorMsg = `unable to write migration marker "%s": %w`
	migrationBackupErrorMsg = `unable to write migration backup "%s": %w`
)

// legacyFileChangedDuringReadErrorMsg reports a v4 write (which truncates first) racing the read.
const (
	legacyFileChangedDuringReadErrorMsg    = `"%s" changed while it was being read`
	legacyFileChangedDuringReadSuggestions = `Run the command again. If this keeps happening, restore or remove "%s".`
)

// legacyFileChangedDuringReadError is the retryable error for a legacy file caught mid-rewrite.
func legacyFileChangedDuringReadError(path string) error {
	return errors.NewErrorWithSuggestions(
		fmt.Sprintf(legacyFileChangedDuringReadErrorMsg, path),
		fmt.Sprintf(legacyFileChangedDuringReadSuggestions, path),
	)
}

// stableFileError marks a failure reading the Stable install's own files (the legacy file or its
// backup), which a non-Stable seed skips instead of failing every command on.
type stableFileError struct{ err error }

func (e *stableFileError) Error() string { return e.err.Error() }

func (e *stableFileError) Unwrap() error { return e.err }

func isStableFileError(err error) bool {
	var stableErr *stableFileError
	return stderrors.As(err, &stableErr)
}

// legacyFileChangedWarningMsg warns that a pre-v5 install wrote to the frozen legacy file after
// migration. Arg: legacyConfigFilename().
const legacyFileChangedWarningMsg = `"%s" changed after your configuration moved to settings.json, contexts.json, and secrets.json. This version doesn't read config.json, so logins and context changes made with a version before v5 won't appear here.`

// legacyFileStampCache is the cache file recording the legacy file's mtime and size at the last
// point it was known good, so a later Stable load can detect a downgraded v4 writing to it.
const legacyFileStampCache = "legacy_config.json"

// legacyFileStamp is the stat sample compared across loads: mtime and size are enough to detect a
// write without reading the file.
type legacyFileStamp struct {
	ModTimeUnixNano int64 `json:"mtime_unix_nano"`
	Size            int64 `json:"size"`
}

// legacyFileStampOf samples info's mtime and size.
func legacyFileStampOf(info os.FileInfo) legacyFileStamp {
	return legacyFileStamp{ModTimeUnixNano: info.ModTime().UnixNano(), Size: info.Size()}
}

// statLegacyFileStamp stats path and reports its stamp, or (zero, false) when the file is gone.
func statLegacyFileStamp(path string) (legacyFileStamp, bool) {
	info, err := os.Stat(path)
	if err != nil {
		return legacyFileStamp{}, false
	}
	return legacyFileStampOf(info), true
}

// recordLegacyFileStamp persists stamp, best-effort: a failed write here must never fail an
// otherwise-successful migration or load. Only Stable re-detects a downgrade writing to this
// file, so other channels skip it.
func recordLegacyFileStamp(stamp legacyFileStamp) {
	if pversion.ProcessChannel() != pversion.Stable {
		return
	}
	if err := newCacheStore().writeJSON(legacyFileStampCache, stamp); err != nil {
		log.CliLogger.Warnf("unable to persist cache: %v", err)
	}
}

// legacyAPIKeyPair holds the secret-bearing fields of a v4 APIKeyPair, retagged json:"-" at HEAD.
type legacyAPIKeyPair struct {
	Key    string `json:"api_key,omitempty"`
	Secret string `json:"api_secret,omitempty"`
	Salt   []byte `json:"salt,omitempty"`
	Nonce  []byte `json:"nonce,omitempty"`
}

// applyTo copies o's secret fields onto live as-is (ciphertext stays ciphertext).
func (o *legacyAPIKeyPair) applyTo(live *APIKeyPair) {
	if o == nil || live == nil {
		return
	}
	live.Secret = o.Secret
	live.Salt = o.Salt
	live.Nonce = o.Nonce
}

// legacyContextState holds the secret-bearing fields of a v4 ContextState, retagged json:"-" at
// HEAD.
type legacyContextState struct {
	AuthToken        string `json:"auth_token"`
	AuthRefreshToken string `json:"auth_refresh_token"`
	Salt             []byte `json:"salt,omitempty"`
	Nonce            []byte `json:"nonce,omitempty"`
}

func (o *legacyContextState) applyTo(live *ContextState) {
	if o == nil || live == nil {
		return
	}
	live.AuthToken = o.AuthToken
	live.AuthRefreshToken = o.AuthRefreshToken
	live.Salt = o.Salt
	live.Nonce = o.Nonce
}

// legacyLoginCredential holds the secret-bearing fields of a v4 LoginCredential, retagged
// json:"-" at HEAD.
type legacyLoginCredential struct {
	EncryptedPassword string `json:"encrypted_password"`
	Salt              []byte `json:"salt,omitempty"`
	Nonce             []byte `json:"nonce,omitempty"`
}

func (o *legacyLoginCredential) applyTo(live *LoginCredential) {
	if o == nil || live == nil {
		return
	}
	live.EncryptedPassword = o.EncryptedPassword
	live.Salt = o.Salt
	live.Nonce = o.Nonce
}

// legacyCredential reaches Credential.APIKeyPair, the only secret-bearing field under Credential.
// The deprecated Credential.Password field (v4 "password,omitempty", credential.go:6 on main) is
// not decoded: it was deleted at HEAD and is read by nothing.
type legacyCredential struct {
	APIKeyPair *legacyAPIKeyPair `json:"api_key_pair"`
}

// legacyKafkaClusterConfig reaches the API keys nested under a Kafka cluster config, keyed by
// API key, whether the config lives under kafka_cluster_configs or kafka_environment_contexts.
type legacyKafkaClusterConfig struct {
	APIKeys map[string]*legacyAPIKeyPair `json:"api_keys"`
}

// applyKafkaClusterConfigs copies secrets from legacy Kafka cluster configs onto their live
// counterparts, matched by cluster id and then by API key.
func applyKafkaClusterConfigs(legacyConfigs map[string]*legacyKafkaClusterConfig, live map[string]*KafkaClusterConfig) {
	for id, lcfg := range legacyConfigs {
		if lcfg == nil {
			continue
		}
		cfg, ok := live[id]
		if !ok {
			continue
		}
		for key, pair := range lcfg.APIKeys {
			if livePair, ok := cfg.APIKeys[key]; ok {
				pair.applyTo(livePair)
			}
		}
	}
}

// legacyKafkaEnvContext reaches Kafka cluster configs nested under an environment, whose v4 tag
// (kafka_cluster_infos) differs from the top-level kafka_cluster_configs tag.
type legacyKafkaEnvContext struct {
	KafkaClusterConfigs map[string]*legacyKafkaClusterConfig `json:"kafka_cluster_infos"`
}

// legacyKafkaClusterContext reaches both places Kafka cluster configs (and their nested API
// keys) can live: directly, and per environment.
type legacyKafkaClusterContext struct {
	KafkaClusterConfigs map[string]*legacyKafkaClusterConfig `json:"kafka_cluster_configs,omitempty"`
	KafkaEnvContexts    map[string]*legacyKafkaEnvContext    `json:"kafka_environment_contexts,omitempty"`
}

// legacySchemaRegistryCluster reaches the deprecated per-context Schema Registry credential.
type legacySchemaRegistryCluster struct {
	SrCredentials *legacyAPIKeyPair `json:"schema_registry_credentials"`
}

// legacyContext reaches every secret- or cache-bearing field nested under a v4 Context.
// FeatureFlags decodes directly into the live *FeatureFlags type: none of its own fields are
// retagged, only its holder (Context.FeatureFlags) is.
type legacyContext struct {
	FeatureFlags           *FeatureFlags                           `json:"feature_flags,omitempty"`
	GlobalAPIKeys          map[string]*legacyAPIKeyPair            `json:"global_api_keys,omitempty"`
	KafkaClusterContext    *legacyKafkaClusterContext              `json:"kafka_cluster_context"`
	SchemaRegistryClusters map[string]*legacySchemaRegistryCluster `json:"schema_registry_clusters,omitempty"`
}

// legacyConfig is the overlay for a v4 config.json: every secret- and cache-bearing field that
// HEAD retagged json:"-" or moved off Config. Decoding a v4 file into this type, alongside a
// normal decode into a live Config, recovers exactly the values HEAD's live decode drops.
type legacyConfig struct {
	LastUpdateCheckAt *time.Time                        `json:"last_update_check_at,omitempty"`
	Credentials       map[string]*legacyCredential      `json:"credentials,omitempty"`
	Contexts          map[string]*legacyContext         `json:"contexts,omitempty"`
	ContextStates     map[string]*legacyContextState    `json:"context_states,omitempty"`
	SavedCredentials  map[string]*legacyLoginCredential `json:"saved_credentials,omitempty"`
}

// applyTo copies every overlay value onto c's live, already-decoded structures, matched by the
// same map keys the JSON used. A key present in the overlay but absent from c (shouldn't happen,
// since both are decoded from the same bytes) is silently skipped rather than fabricating one.
func (o *legacyConfig) applyTo(c *Config) {
	c.LastUpdateCheckAt = o.LastUpdateCheckAt

	for name, lcred := range o.Credentials {
		if lcred == nil {
			continue
		}
		if cred, ok := c.Credentials[name]; ok {
			lcred.APIKeyPair.applyTo(cred.APIKeyPair)
		}
	}

	for name, lctx := range o.Contexts {
		ctx, ok := c.Contexts[name]
		if !ok || lctx == nil {
			continue
		}

		ctx.FeatureFlags = lctx.FeatureFlags

		for key, pair := range lctx.GlobalAPIKeys {
			if live, ok := ctx.GlobalAPIKeys[key]; ok {
				pair.applyTo(live)
			}
		}

		if lctx.KafkaClusterContext != nil && ctx.KafkaClusterContext != nil {
			applyKafkaClusterConfigs(lctx.KafkaClusterContext.KafkaClusterConfigs, ctx.KafkaClusterContext.KafkaClusterConfigs)

			for envName, lenv := range lctx.KafkaClusterContext.KafkaEnvContexts {
				if lenv == nil {
					continue
				}
				if env, ok := ctx.KafkaClusterContext.KafkaEnvContexts[envName]; ok {
					applyKafkaClusterConfigs(lenv.KafkaClusterConfigs, env.KafkaClusterConfigs)
				}
			}
		}

		for id, lsr := range lctx.SchemaRegistryClusters {
			if lsr == nil {
				continue
			}
			if sr, ok := ctx.SchemaRegistryClusters[id]; ok {
				lsr.SrCredentials.applyTo(sr.SrCredentials)
			}
		}
	}

	for name, lstate := range o.ContextStates {
		if lstate == nil {
			continue
		}
		if state, ok := c.ContextStates[name]; ok {
			lstate.applyTo(state)
		}
	}

	for name, lcred := range o.SavedCredentials {
		if lcred == nil {
			continue
		}
		if cred, ok := c.SavedCredentials[name]; ok {
			lcred.applyTo(cred)
		}
	}
}

// afterLegacyConfigStat is a test seam called between the legacy file's stat and its read, the
// window a v4 rewrite can race.
var afterLegacyConfigStat = func(path string) {}

// readLegacyConfigFile reads a v4 config.json at path, returning its bytes and the stat taken
// before the read. It returns (nil, nil, nil) when the file is missing, not a regular file
// (following symlinks), or zero bytes, and a retryable error when the file changed during the
// read; any other failure is a *stableFileError.
func readLegacyConfigFile(path string) ([]byte, os.FileInfo, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil, nil
		}
		return nil, nil, unreadableLegacyFileError(path, err)
	}
	if !info.Mode().IsRegular() || info.Size() == 0 {
		return nil, nil, nil
	}

	afterLegacyConfigStat(path)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, unreadableLegacyFileError(path, err)
	}

	// v4 truncates then writes in place, so a read overlapping that write can be torn (empty, or a
	// prefix that fails to parse) without the file ever looking absent: re-stat to catch it.
	after, err := os.Stat(path)
	if err != nil && !os.IsNotExist(err) {
		return nil, nil, unreadableLegacyFileError(path, err)
	}
	if err != nil || int64(len(data)) != info.Size() || after.Size() != info.Size() || !after.ModTime().Equal(info.ModTime()) {
		return nil, nil, legacyFileChangedDuringReadError(path)
	}
	return data, info, nil
}

func unreadableLegacyFileError(path string, err error) error {
	return &stableFileError{fmt.Errorf(errors.UnableToReadConfigurationFileErrorMsg, path, err)}
}

// applyLegacyConfig decodes a v4 config.json's bytes onto c, secrets included. It parses the
// live shape and the secret overlay into scratch values first, so a parse error leaves c
// untouched; only once both succeed does it decode onto c directly (preserving New()'s defaults
// for keys absent from the file, and c.Filename, exactly like v4's own decode-onto-c) and copy
// the overlay's values into c's json:"-" fields. It never wires contexts or saves.
func (c *Config) applyLegacyConfig(path string, data []byte) error {
	live := New()
	if err := json.Unmarshal(data, live); err != nil {
		return fmt.Errorf(errors.UnableToReadConfigurationFileErrorMsg, path, err)
	}

	overlay := new(legacyConfig)
	if err := json.Unmarshal(data, overlay); err != nil {
		return fmt.Errorf(errors.UnableToReadConfigurationFileErrorMsg, path, err)
	}

	if err := json.Unmarshal(data, c); err != nil {
		return fmt.Errorf(errors.UnableToReadConfigurationFileErrorMsg, path, err)
	}

	overlay.applyTo(c)
	return nil
}

// decodeLegacyConfig decodes a v4 config.json onto c and wires its contexts: save() and
// saveSecretStore reach tokens through ctx.GetState(), which only wiring sets.
func (c *Config) decodeLegacyConfig(path string, data []byte) error {
	if err := c.applyLegacyConfig(path, data); err != nil {
		return err
	}
	if err := c.wireContexts(); err != nil {
		return fmt.Errorf(migrationErrorMsg, path, err)
	}
	return nil
}

// afterLegacyMigration is a test seam called once per completed migration.
var afterLegacyMigration = func() {}

// legacyMigration is what one migrateFromLegacy call did, carried out of the store lock so
// announceMigration can print about it afterward.
type legacyMigration struct {
	legacyPath string
	migrated   bool
	// seedErr is why a non-Stable channel skipped its seed; Stable returns the error instead.
	seedErr error
}

// migrateFromLegacy seeds the split stores from a v4 config.json when one is pending. The caller
// must hold the store lock. v4 only ever wrote the stable path, so every channel reads it from
// there, and v5 never writes or renames it. The marker goes down before the first store write
// and the backup after the last, so a failed or interrupted run is retried whole by the next load.
func (c *Config) migrateFromLegacy() (legacyMigration, error) {
	path, ok := legacyConfigFilename()
	if !ok {
		// don't guess a cwd-relative legacy path, which could be any directory's config.json.
		return legacyMigration{}, nil
	}
	result := legacyMigration{legacyPath: path}

	pending, err := legacyMigrationPending(path)
	if err != nil {
		if isStableFileError(err) {
			return c.failOrSkipSeed(result, err)
		}
		return result, err
	}
	if !pending {
		return result, nil
	}

	data, info, err := readLegacyConfigFile(path)
	if err != nil {
		if isStableFileError(err) {
			return c.failOrSkipSeed(result, err)
		}
		return result, err
	}
	if info == nil {
		// the guard's stat saw a non-empty regular file, so it changed since: retry next load.
		return result, legacyFileChangedDuringReadError(path)
	}

	if err := c.decodeLegacyConfig(path, data); err != nil {
		return c.failOrSkipSeed(result, err)
	}

	marker := legacyMigratingFilename()
	if err := writeFileAtomic(marker, nil); err != nil {
		return result, fmt.Errorf(migrationMarkerErrorMsg, marker, err)
	}

	// nothing on disk is an ancestor of the legacy state, and an interrupted run's stores are only
	// its own partial output, so every store is written whole: a nil config baseline and a
	// non-nil secret one (a nil secret baseline would merge with an existing secrets.json).
	c.baseline = nil
	c.secretBaseline = &secretFile{}
	if err := c.saveLocked(); err != nil {
		return result, fmt.Errorf(migrationErrorMsg, path, err)
	}
	c.saveCache()

	backup := legacyBackupFilename()
	if err := writeFileAtomic(backup, data); err != nil {
		return result, fmt.Errorf(migrationBackupErrorMsg, backup, err)
	}
	// the pre-read stat, so a v4 write landing after the read still trips the downgrade warning.
	recordLegacyFileStamp(legacyFileStampOf(info))

	// a leftover marker is harmless once the backup exists: the next guard removes it.
	removeMigrationMarker()

	afterLegacyMigration()
	result.migrated = true
	return result, nil
}

// failOrSkipSeed returns err on Stable. A non-Stable seed is a convenience copy, so it must not
// fail every command: nothing is written yet, so this channel warns and starts fresh. Dropping a
// marker from an earlier run keeps the warning from repeating on every load.
func (c *Config) failOrSkipSeed(result legacyMigration, err error) (legacyMigration, error) {
	if pversion.ProcessChannel() == pversion.Stable {
		return result, err
	}
	c.resetToNew()
	clearMigrationMarker()
	var stableErr *stableFileError
	if stderrors.As(err, &stableErr) {
		err = stableErr.err
	}
	result.seedErr = errorCause(err)
	return result, nil
}

// announceMigration prints the one-time user-facing message for this Load, or (when this Load
// did not migrate) re-checks the frozen legacy file for a downgrade write. It runs after
// loadLocked's lock is released and only on a successful Load.
func (c *Config) announceMigration(m legacyMigration) {
	switch {
	case m.migrated:
		c.announceCompletedMigration(m.legacyPath)
	case m.seedErr != nil:
		output.ErrPrintln(c.EnableColor, fmt.Sprintf(migrationSeedSkippedMsg, m.legacyPath, m.seedErr))
		output.ErrPrintln(c.EnableColor, "")
	default:
		c.warnIfLegacyFileChanged()
	}
}

// announceCompletedMigration prints the Stable move announcement or the non-Stable seed
// announcement, skipping either when the migrated config has no contexts worth telling the user
// about.
func (c *Config) announceCompletedMigration(legacyPath string) {
	if len(c.Contexts) == 0 {
		return
	}

	if pversion.ProcessChannel() == pversion.Stable {
		output.ErrPrintln(c.EnableColor, migrationAnnouncementMsg)
		output.ErrPrintln(c.EnableColor, "")
		return
	}

	stateDir, err := StateDir()
	if err != nil {
		// StateDir only fails when the home directory can't be resolved; fall back to the same
		// (relative) path stateDirPath itself tolerates rather than skip the announcement.
		stateDir = stateDirPath("")
	}
	output.ErrPrintln(c.EnableColor, fmt.Sprintf(migrationSeedMsg, legacyPath, stateDir))
	output.ErrPrintln(c.EnableColor, "")
}

// warnIfLegacyFileChanged re-detects a v4 downgrade writing to the frozen legacy file. It only
// runs on Stable once a migration has completed (the backup exists) and costs one stat of the
// legacy file.
func (c *Config) warnIfLegacyFileChanged() {
	if pversion.ProcessChannel() != pversion.Stable {
		return
	}
	if _, err := os.Stat(legacyBackupFilename()); err != nil {
		return
	}

	legacyPath, ok := legacyConfigFilename()
	if !ok {
		return
	}
	current, exists := statLegacyFileStamp(legacyPath)
	if !exists {
		return
	}

	s := newCacheStore()
	var previous legacyFileStamp
	hadStamp := s.readJSON(legacyFileStampCache, &previous)
	if hadStamp && previous == current {
		return
	}

	if hadStamp {
		output.ErrPrintln(c.EnableColor, fmt.Sprintf(legacyFileChangedWarningMsg, legacyPath))
		output.ErrPrintln(c.EnableColor, "")
	}

	if err := s.writeJSON(legacyFileStampCache, current); err != nil {
		log.CliLogger.Warnf("unable to persist cache: %v", err)
	}
}

// legacyMigrationPending is the migration guard. It stats the legacy file first, so a v5-only
// machine never touches the migration files. A backup means this channel already migrated, so it
// never re-runs, even with every store deleted. Otherwise a marker means an interrupted run to
// resume, and without one only a machine with none of the three stores migrates: any store left
// on disk is v5 state, so a deliberate partial reset never re-imports the v4 file. Every "don't
// migrate" answer but a mid-rewrite legacy file drops the marker, so it can't revive a later run.
func legacyMigrationPending(legacyPath string) (bool, error) {
	legacy, legacyErr := os.Stat(legacyPath)
	if os.IsNotExist(legacyErr) || (legacyErr == nil && !legacy.Mode().IsRegular()) {
		clearMigrationMarker()
		return false, nil
	}

	migrated, err := migrationFileExists(legacyBackupFilename())
	if err != nil {
		return false, err
	}
	if migrated {
		clearMigrationMarker()
		return false, nil
	}

	interrupted, err := migrationFileExists(legacyMigratingFilename())
	if err != nil {
		return false, err
	}
	if !interrupted {
		fresh, err := noStoresExist()
		if err != nil || !fresh {
			return false, err
		}
	}

	if pversion.ProcessChannel() != pversion.Stable {
		// once Stable has migrated, the legacy file is frozen and stale, so it's no seed.
		stableMigrated, err := migrationFileExists(filepath.Join(filepath.Dir(legacyPath), legacyBackupName))
		if err != nil {
			return false, &stableFileError{err}
		}
		if stableMigrated {
			if interrupted {
				removeMigrationMarker()
			}
			return false, nil
		}
	}

	// deferred until here so an unreadable legacy file only fails a load that would migrate.
	if legacyErr != nil {
		return false, unreadableLegacyFileError(legacyPath, legacyErr)
	}
	if legacy.Size() == 0 {
		if interrupted {
			// v4 truncates before it rewrites, so keep the marker and resume on the next load.
			return false, legacyFileChangedDuringReadError(legacyPath)
		}
		return false, nil
	}
	return true, nil
}

// clearMigrationMarker removes the marker if there is one, best-effort.
func clearMigrationMarker() {
	marker := legacyMigratingFilename()
	if _, err := os.Stat(marker); err != nil {
		if !os.IsNotExist(err) {
			log.CliLogger.Warnf(`unable to check migration marker "%s": %v`, marker, err)
		}
		return
	}
	removeMigrationMarker()
}

// removeMigrationMarker removes a marker known to exist, best-effort: a leftover one is only
// ever acted on alongside a legacy file and no backup.
func removeMigrationMarker() {
	marker := legacyMigratingFilename()
	if err := os.Remove(marker); err != nil {
		log.CliLogger.Warnf(`unable to remove migration marker "%s": %v`, marker, err)
	}
}

// errorCause unwraps err once, so a warning that already names the legacy file doesn't repeat
// the wrapper that names it too.
func errorCause(err error) error {
	if cause := stderrors.Unwrap(err); cause != nil {
		return cause
	}
	return err
}

// noStoresExist reports whether settings.json, contexts.json, and secrets.json are all absent.
// A zero-byte or otherwise unreadable store is present.
func noStoresExist() (bool, error) {
	for _, store := range []string{SettingsFilename(), ContextsFilename(), SecretsFilename()} {
		exists, err := migrationFileExists(store)
		if err != nil || exists {
			return false, err
		}
	}
	return true, nil
}

// migrationFileExists reports whether path exists. Only "not exist" counts as absent: any other
// stat failure is returned, so an unreadable file never passes for a missing one.
func migrationFileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, fmt.Errorf(migrationCheckErrorMsg, path, err)
}
