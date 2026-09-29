package config

import (
	"encoding/json"
	"fmt"
	"os"
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
// the Stable install's legacy file. Args: legacyConfigPath(), the running channel's StateDir().
const migrationSeedMsg = `Confluent CLI copied your contexts and logins from "%s" into this build's own configuration in "%s". Changes you make here won't affect that installation, or vice versa.`

// legacyFileChangedWarningMsg warns that a pre-v5 install wrote to the frozen legacy file after
// migration. Arg: legacyConfigPath().
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

// statLegacyFileStamp stats path and reports its stamp, or (zero, false) when the file is gone.
func statLegacyFileStamp(path string) (legacyFileStamp, bool) {
	info, err := os.Stat(path)
	if err != nil {
		return legacyFileStamp{}, false
	}
	return legacyFileStamp{ModTimeUnixNano: info.ModTime().UnixNano(), Size: info.Size()}, true
}

// recordLegacyFileStamp stamps the legacy file's current mtime/size, best-effort: a failed write
// here must never fail an otherwise-successful migration or load.
func recordLegacyFileStamp(path string) {
	stamp, ok := statLegacyFileStamp(path)
	if !ok {
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

// readLegacyConfigFile reads a v4 config.json at path. It returns (nil, false, nil) when the
// file is missing, not a regular file (following symlinks), or zero bytes; any other read
// failure is a hard error.
func readLegacyConfigFile(path string) ([]byte, bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf(errors.UnableToReadConfigurationFileErrorMsg, path, err)
	}
	if !info.Mode().IsRegular() || info.Size() == 0 {
		return nil, false, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, false, fmt.Errorf(errors.UnableToReadConfigurationFileErrorMsg, path, err)
	}
	if len(data) == 0 {
		return nil, false, nil
	}
	return data, true, nil
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

// afterLegacyMigration is a test seam called once per completed migration.
var afterLegacyMigration = func() {}

// migrateFromLegacy seeds the split stores from a v4 config.json when one is pending, and reports
// whether it did. The caller must hold the store lock. The backup is written last, as the commit
// marker: until it exists, a failed or interrupted run is retried whole by the next load.
func (c *Config) migrateFromLegacy() (bool, error) {
	pending, err := legacyMigrationPending()
	if err != nil || !pending {
		return false, err
	}

	path := legacyConfigPath()
	data, found, err := readLegacyConfigFile(path)
	if err != nil || !found {
		return false, err
	}

	if err := c.applyLegacyConfig(path, data); err != nil {
		return false, err
	}
	// save() and saveSecretStore reach tokens through ctx.GetState(), which only wiring sets.
	if err := c.wireContexts(); err != nil {
		return false, fmt.Errorf(migrationErrorMsg, path, err)
	}
	// nothing on disk is an ancestor of the legacy state, so there is no baseline to merge against.
	c.baseline = nil
	c.secretBaseline = nil
	if err := c.saveLocked(); err != nil {
		return false, fmt.Errorf(migrationErrorMsg, path, err)
	}
	c.saveCache()

	if err := writeFileAtomic(legacyBackupPath(), data); err != nil {
		return false, fmt.Errorf("unable to write migration backup %s: %w", legacyBackupPath(), err)
	}
	// only the Stable channel ever re-detects a downgrade writing to this file, so only it needs
	// the stamp; Prerelease/Dev never write or compare against it.
	if pversion.ProcessChannel() == pversion.Stable {
		recordLegacyFileStamp(path)
	}
	afterLegacyMigration()
	return true, nil
}

// announceMigration prints the one-time user-facing message for this Load, or (when this Load
// did not migrate) re-checks the frozen legacy file for a downgrade write. It runs after
// loadLocked's lock is released and only on a successful Load.
func (c *Config) announceMigration(migrated bool) {
	if migrated {
		c.announceCompletedMigration()
		return
	}
	if pversion.ProcessChannel() == pversion.Stable {
		c.warnIfLegacyFileChanged()
	}
}

// announceCompletedMigration prints the Stable move announcement or the non-Stable seed
// announcement, skipping either when the migrated config has no contexts worth telling the user
// about.
func (c *Config) announceCompletedMigration() {
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
	output.ErrPrintln(c.EnableColor, fmt.Sprintf(migrationSeedMsg, legacyConfigPath(), stateDir))
	output.ErrPrintln(c.EnableColor, "")
}

// warnIfLegacyFileChanged re-detects a v4 downgrade writing to the frozen legacy file. It only
// runs once a migration has completed (the backup exists) and costs one stat of the legacy file.
func (c *Config) warnIfLegacyFileChanged() {
	if _, err := os.Stat(legacyBackupPath()); err != nil {
		return
	}

	current, exists := statLegacyFileStamp(legacyConfigPath())
	if !exists {
		return
	}

	s := newCacheStore()
	var previous legacyFileStamp
	hadStamp := s.readJSON(legacyFileStampCache, &previous)

	if hadStamp && (previous.ModTimeUnixNano != current.ModTimeUnixNano || previous.Size != current.Size) {
		output.ErrPrintln(c.EnableColor, fmt.Sprintf(legacyFileChangedWarningMsg, legacyConfigPath()))
		output.ErrPrintln(c.EnableColor, "")
	}

	if err := s.writeJSON(legacyFileStampCache, current); err != nil {
		log.CliLogger.Warnf("unable to persist cache: %v", err)
	}
}

// legacyMigrationPending is the store side of the migration guard: no backup yet (a completed
// migration never re-runs, even if the stores are later deleted) and no secrets.json. Keying on
// secrets.json alone is sound because every save writes settings.json, then contexts.json, then
// secrets.json, so an interrupted run never leaves secrets.json behind. It also means deleting
// contexts.json (the documented context reset) never re-imports the legacy file.
func legacyMigrationPending() (bool, error) {
	if _, err := os.Stat(legacyBackupPath()); err == nil {
		return false, nil
	} else if !os.IsNotExist(err) {
		return false, fmt.Errorf("unable to check migration backup %s: %w", legacyBackupPath(), err)
	}
	return !storeFileHasData(SecretsFilename()), nil
}

// storeFileHasData reports whether path is a regular file holding data. Missing or zero-byte holds
// nothing, matching readStoreFile.
func storeFileHasData(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.Mode().IsRegular() && info.Size() > 0
}
