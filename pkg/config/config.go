package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/log"
	"github.com/confluentinc/cli/v4/pkg/secret"
	"github.com/confluentinc/cli/v4/pkg/utils"
	pversion "github.com/confluentinc/cli/v4/pkg/version"
)

const emptyFieldIndicator = "EMPTY"

const signupSuggestion = "If you need a Confluent Cloud account, sign up with `confluent cloud-signup`."

var (
	RequireCloudLoginErr = &errors.RunRequirementError{
		ErrorMsg:       "you must log in to Confluent Cloud to use this command",
		SuggestionsMsg: "Log in with `confluent login`.\n" + signupSuggestion,
	}
	RequireCloudLoginOrgUnsuspendedErr = &errors.RunRequirementError{
		ErrorMsg:       "you must unsuspend your organization to use this command",
		SuggestionsMsg: errors.SuspendedOrganizationSuggestions,
	}
	RequireCloudLoginFreeTrialEndedOrgUnsuspendedErr = &errors.RunRequirementError{
		ErrorMsg:       "you must unsuspend your organization to use this command",
		SuggestionsMsg: errors.EndOfFreeTrialSuggestions,
	}
	RequireCloudLoginPauseTrialOrgUnsuspendedErr = &errors.RunRequirementError{
		ErrorMsg:       "you must resume your organization to use this command",
		SuggestionsMsg: errors.PauseTrialSuggestions,
	}
	RequireCloudLoginOrOnPremErr = &errors.RunRequirementError{
		ErrorMsg:       "you must log in to use this command",
		SuggestionsMsg: "Log in with `confluent login`.\n" + signupSuggestion,
	}
	RequireNonAPIKeyCloudLoginErr = &errors.RunRequirementError{
		ErrorMsg:       "you must log in to Confluent Cloud with a username and password to use this command",
		SuggestionsMsg: "Log in with `confluent login`.\n" + signupSuggestion,
	}
	RequireNonAPIKeyCloudLoginOrOnPremLoginErr = &errors.RunRequirementError{
		ErrorMsg:       "you must log in to Confluent Cloud with a username and password or log in to Confluent Platform to use this command",
		SuggestionsMsg: "Log in with `confluent login` or `confluent login --url <mds-url>`.\n" + signupSuggestion,
	}
	RequireCloudLogout = &errors.RunRequirementError{
		ErrorMsg:       "you must log out of Confluent Cloud to use this command",
		SuggestionsMsg: "Log out with `confluent logout`.\n",
	}
	RequireOnPremLoginErr = &errors.RunRequirementError{
		ErrorMsg:       "you must log in to Confluent Platform to use this command",
		SuggestionsMsg: "Log in to Confluent Platform with `confluent login --url <mds-url>`.",
	}
	RunningOnPremCommandInCloudErr = &errors.RunRequirementError{
		ErrorMsg:       "this is not a Confluent Cloud command. You must log in to Confluent Platform to use this command",
		SuggestionsMsg: "Log in to Confluent Platform with `confluent login --url <mds-url>`.\nSee available commands with `--help`.",
	}
	RunningSimilarOnPremCommandInCloudErr = func(cmdPath, suggestedCmdPath string) errors.CLITypedError {
		return &errors.RunRequirementError{
			ErrorMsg:       fmt.Sprintf("`%s` is not a Confluent Cloud command. Did you mean `%s`?", cmdPath, suggestedCmdPath),
			SuggestionsMsg: fmt.Sprintf("If you are a Confluent Cloud user, run `%s` instead.\nIf you are attempting to connect to Confluent Platform, login with `confluent login --url <mds-url>` to use `%s`.", suggestedCmdPath, cmdPath),
		}
	}
)

// Config represents the CLI configuration.
type Config struct {
	DisableFeatureFlags       bool       `json:"disable_feature_flags"`
	DisablePlugins            bool       `json:"disable_plugins"`
	DisablePluginsOnceWindows bool       `json:"disable_plugins_once_windows,omitempty"`
	DisableUpdateCheck        bool       `json:"disable_update_check"`
	EnableColor               bool       `json:"enable_color"`
	LastUpdateCheckAt         *time.Time `json:"-"`

	Platforms        map[string]*Platform        `json:"platforms,omitempty"`
	Credentials      map[string]*Credential      `json:"credentials,omitempty"`
	CurrentContext   string                      `json:"current_context"`
	Contexts         map[string]*Context         `json:"contexts,omitempty"`
	ContextStates    map[string]*ContextState    `json:"context_states,omitempty"`
	SavedCredentials map[string]*LoginCredential `json:"saved_credentials,omitempty"`
	LocalPorts       *LocalPorts                 `json:"local_ports,omitempty"`

	// The following configurations are not persisted between runs or are read-only
	DisableUpdates bool              `json:"-"`
	Filename       string            `json:"-"`
	IsTest         bool              `json:"-"`
	Version        *pversion.Version `json:"-"`

	overwrittenCurrentContext      string
	overwrittenCurrentEnvironment  string
	overwrittenCurrentKafkaCluster string

	// baseline is this process's view of the persisted config as of its last
	// load or successful save. Save() diffs baseline against the live config to
	// learn what THIS process changed, so a concurrent writer's fields can be
	// preserved. Never serialized.
	baseline *Config

	// secretBaseline is this process's view of the persisted secret store as of its last
	// load or successful save, mirroring baseline above but for secrets.json. It needs its
	// own field (rather than living inside baseline) because deepCopyPersisted's JSON round
	// trip drops every json:"-" secret field, so baseline can never hold one. Never
	// serialized; see loadSecretStore and saveSecretStore in secret_store.go.
	secretBaseline *secretFile

	// writing is set on the exact config whose Validate() runs under the write
	// lock. Validate()'s in-memory normalization (nil-map init, invalid-active-
	// cluster reset) reaches back through Context.Save() to persist itself; that
	// nested Save() would deadlock on the already-held sidecar lock. The flag
	// short-circuits it, so the enclosing locked write persists the normalized
	// struct as soon as Validate() returns. Never serialized.
	//
	// Known limitation: KafkaClusterContext.Validate() couples normalization to
	// persistence via Save(), so a lock timeout hit during that normalization
	// surfaces as a panic rather than a returned error. Splitting normalization
	// from persistence is a deferred follow-up.
	writing bool

	// Deprecated
	DisablePluginsOnce bool `json:"disable_plugins_once,omitempty"`
}

func (c *Config) SetOverwrittenCurrentContext(context string) {
	if context == "" {
		context = emptyFieldIndicator
	}
	if c.overwrittenCurrentContext == "" {
		c.overwrittenCurrentContext = context
	}
}

func (c *Config) SetOverwrittenCurrentEnvironment(environmentId string) {
	if c.overwrittenCurrentEnvironment == "" {
		c.overwrittenCurrentEnvironment = environmentId
	}
}

func (c *Config) SetOverwrittenCurrentKafkaCluster(clusterId string) {
	if clusterId == "" {
		clusterId = emptyFieldIndicator
	}
	if c.overwrittenCurrentKafkaCluster == "" {
		c.overwrittenCurrentKafkaCluster = clusterId
	}
}

func New() *Config {
	return &Config{
		Platforms:        make(map[string]*Platform),
		Credentials:      make(map[string]*Credential),
		Contexts:         make(map[string]*Context),
		ContextStates:    make(map[string]*ContextState),
		SavedCredentials: make(map[string]*LoginCredential),
		Version:          new(pversion.Version),
		EnableColor:      true,
	}
}

func (c *Config) DecryptContextStates() error {
	if context := c.Context(); context != nil {
		state := c.ContextStates[context.Name]
		if state != nil {
			if err := state.DecryptAuthToken(context.Name); err != nil {
				return err
			}
			if err := state.DecryptAuthRefreshToken(context.Name); err != nil {
				return err
			}
		}
		context.State = state
	}
	return c.Validate()
}

func (c *Config) DecryptCredentials() error {
	if credentials := c.Credentials; c.Credentials != nil {
		for _, credential := range credentials {
			if credential.APIKeyPair != nil {
				if err := credential.APIKeyPair.DecryptSecret(); err != nil {
					return err
				}
			}
		}
	}
	return c.Validate()
}

// afterMissingConfigRead is a test seam invoked in Load's missing-file branch after the
// read and before the initial (locked) save. It is a no-op in production; a test uses it
// to deterministically interleave a competing writer at exactly that point.
var afterMissingConfigRead = func() {}

// Load reads the CLI config from disk.
// Save a default version if none exists yet.
func (c *Config) Load() error {
	filename := c.GetFilename()

	missing, err := c.loadLocked(filename)
	if err != nil {
		return err
	}
	if missing {
		// Save a default version if none exists yet. Snapshot a baseline first so this
		// save merges under the lock instead of overwriting: another session can create
		// the config between this missing-file read and the locked save, and that file
		// must survive. A config constructed without Load keeps a nil baseline and still
		// writes whole. This runs outside loadLocked's lock: Save() re-acquires the same
		// sidecar lock, and flock is not reentrant.
		c.snapshotBaseline()
		afterMissingConfigRead()
		if err := c.Save(); err != nil {
			return fmt.Errorf("unable to save configuration file: %w", err)
		}
		return nil
	}

	if err := c.wireContexts(); err != nil {
		return err
	}
	c.snapshotBaseline() // baseline = pristine on-disk state, before migrations

	var save bool
	for _, context := range c.Contexts {
		// Migrate deprecated NetrcMachineName to MachineName
		if context.NetrcMachineName != "" && context.MachineName == "" {
			context.MachineName = context.NetrcMachineName
			context.NetrcMachineName = ""
			save = true
		}
	}

	// Migrate deprecated DisablePluginsOnce to DisablePluginsOnceWindows
	if c.DisablePluginsOnce && !c.DisablePluginsOnceWindows {
		c.DisablePluginsOnceWindows = true
		c.DisablePluginsOnce = false
		save = true
	}

	if runtime.GOOS == "windows" && !c.DisablePluginsOnceWindows {
		c.DisablePlugins = true
		c.DisablePluginsOnceWindows = true
		save = true
	}

	if save {
		if err := c.Save(); err != nil {
			return err
		}
	}

	return c.Validate()
}

// loadLocked performs Load's disk reads - config.json, the secret store, and the cache -
// under the same sidecar lock Save() uses, so a reader never sees config.json and
// secrets.json from two different Save() generations. It returns missing=true when
// config.json does not exist yet, leaving that branch's handling (which calls Save() and
// so must not run while this lock is held) to the caller. The lock is released via defer
// before this function returns, well before wireContexts/Validate run: Validate's
// normalization can re-enter Save(), which acquires this same lock, and flock is not
// reentrant.
func (c *Config) loadLocked(filename string) (bool, error) {
	// Create the config directory before opening the sidecar lock file inside it, same as
	// Save(): on a fresh machine (parent directory absent) opening the lock would ENOENT
	// before we ever get to discover the config file itself is missing.
	if err := os.MkdirAll(filepath.Dir(filename), 0700); err != nil {
		return false, fmt.Errorf("unable to create config directory %s: %w", filename, err)
	}

	lock := newFileLock(filename)
	if err := lock.lock(lockTimeout); err != nil {
		return false, err
	}
	defer func() { _ = lock.unlock() }()

	input, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return true, nil
		}
		return false, fmt.Errorf(errors.UnableToReadConfigurationFileErrorMsg, filename, err)
	}

	if err := json.Unmarshal(input, c); err != nil {
		return false, fmt.Errorf(errors.UnableToReadConfigurationFileErrorMsg, filename, err)
	}

	if err := c.loadSecretStore(); err != nil {
		return false, err
	}

	// Load the cache here, under the same lock, rather than after a migration-triggered
	// Save() below: that Save() also persists the cache, and saving before the cache is
	// loaded would overwrite the on-disk cache (update-check timestamp and feature flags)
	// with zero-value fields.
	c.loadCache()

	return false, nil
}

// updateCheckCache is the disposable cache-store representation of the fields split
// out of Config below. loadCache/saveCache round-trip them through the cache store
// instead of the config file.
type updateCheckCache struct {
	LastUpdateCheckAt *time.Time `json:"last_update_check_at,omitempty"`
}

func (c *Config) loadCache() {
	s := newCacheStore()
	var uc updateCheckCache
	if s.readJSON("update_check.json", &uc) {
		c.LastUpdateCheckAt = uc.LastUpdateCheckAt
	}
	c.loadFeatureFlagCache(s)
}

// saveCache persists the cache-store fields best-effort: CacheDir is disposable and
// non-authoritative, so a write failure here (e.g. a blocked/full/permission-denied
// cache dir) must never fail an otherwise-successful Save().
func (c *Config) saveCache() {
	s := newCacheStore()
	if err := s.writeJSON("update_check.json", updateCheckCache{LastUpdateCheckAt: c.LastUpdateCheckAt}); err != nil {
		log.CliLogger.Warnf("unable to persist cache: %v", err)
	}
	if err := c.saveFeatureFlagCache(s); err != nil {
		log.CliLogger.Warnf("unable to persist cache: %v", err)
	}
}

// loadFeatureFlagCache restores each context's FeatureFlags from the cache store, keyed
// by context name.
func (c *Config) loadFeatureFlagCache(s *cacheStore) {
	flags := map[string]*FeatureFlags{}
	if !s.readJSON("feature_flags.json", &flags) {
		return
	}
	for name, ctx := range c.Contexts {
		if ff, ok := flags[name]; ok {
			ctx.FeatureFlags = ff
		}
	}
}

func (c *Config) saveFeatureFlagCache(s *cacheStore) error {
	flags := map[string]*FeatureFlags{}
	for name, ctx := range c.Contexts {
		if ctx.FeatureFlags != nil {
			flags[name] = ctx.FeatureFlags
		}
	}
	// Always rewrite the whole map, even when empty: a context whose flags were
	// cleared or deleted must not leave a stale entry that Load would resurrect.
	return s.writeJSON("feature_flags.json", flags)
}

// wireContexts rebuilds the cross-references Load() relies on: each context's
// Credential/Platform/Config back-pointer, its KafkaClusterContext parent, and
// its State pointer aliased to c.ContextStates[name]. Validate()'s DeepEqual on
// context state only holds because State and ContextStates[name] are the same
// object, so a bare json.Unmarshal is never enough.
func (c *Config) wireContexts() error {
	for _, context := range c.Contexts {
		if context.Name == "" {
			return errors.NewCorruptedConfigError(errors.NoNameContextErrorMsg, "", c.Filename)
		}
		if context.CredentialName == "" {
			return errors.NewCorruptedConfigError(errors.UnspecifiedCredentialErrorMsg, context.Name, c.Filename)
		}
		if context.PlatformName == "" {
			return errors.NewCorruptedConfigError(errors.UnspecifiedPlatformErrorMsg, context.Name, c.Filename)
		}
		context.Credential = c.Credentials[context.CredentialName]
		context.Platform = c.Platforms[context.PlatformName]
		context.Config = c
		if context.KafkaClusterContext == nil {
			return errors.NewCorruptedConfigError(`context "%s" missing KafkaClusterContext`, context.Name, c.Filename)
		}
		context.KafkaClusterContext.Context = context
		context.State = c.ContextStates[context.Name]
	}
	return nil
}

// readConfigFromDisk re-reads the persisted config and rebuilds its pointer
// graph. It does NOT run migrations and never writes; it is the "theirs" side
// of Save()'s merge, so it must not recurse into Save(). json:"-" fields
// (Filename, IsTest, Version, DisableUpdates) are copied from template because a
// fresh unmarshal cannot recover them.
func readConfigFromDisk(path string, template *Config) (*Config, error) {
	input, err := os.ReadFile(path)
	if err != nil {
		// Keep the raw error for a missing file so saveLocked's os.IsNotExist
		// check still fires; wrap any other read error as Load() does.
		if os.IsNotExist(err) {
			return nil, err
		}
		return nil, fmt.Errorf(errors.UnableToReadConfigurationFileErrorMsg, path, err)
	}

	disk := New()
	if err := json.Unmarshal(input, disk); err != nil {
		return nil, fmt.Errorf(errors.UnableToReadConfigurationFileErrorMsg, path, err)
	}

	disk.Filename = template.Filename
	disk.IsTest = template.IsTest
	disk.Version = template.Version
	disk.DisableUpdates = template.DisableUpdates

	if err := disk.wireContexts(); err != nil {
		return nil, err
	}
	return disk, nil
}

// snapshotBaseline deep-copies the persisted fields into c.baseline via a JSON
// round-trip (json:"-" and unexported fields are intentionally excluded, since the
// merge only diffs persisted state).
func (c *Config) snapshotBaseline() {
	c.baseline = c.deepCopyPersisted()
}

// deepCopyPersisted returns an independent copy of c's persisted fields via a
// JSON round-trip. json:"-" and unexported fields (Filename, baseline, ...) are
// intentionally dropped, so only persisted state participates in the merge. The
// copy shares no pointers with c, so wiring or encrypting it never mutates c.
func (c *Config) deepCopyPersisted() *Config {
	data, err := json.Marshal(c)
	if err != nil {
		return New()
	}
	b := New()
	_ = json.Unmarshal(data, b)
	return b
}

// Save atomically and safely persists the config. It serializes writers on a
// sidecar lock, re-reads the current on-disk state under the lock, three-way-
// merges this process's own changes onto it (so a concurrent session's fields
// are not lost), and writes the result atomically.
func (c *Config) Save() error {
	// A Save() re-entered from Validate()'s normalization while this process
	// already holds the lock is a no-op: the enclosing locked write persists the
	// fully-normalized struct as soon as Validate() returns.
	if c.writing {
		return nil
	}

	// Create the config directory before opening the sidecar lock file inside it:
	// on a fresh machine (~/.confluent absent) opening the lock would ENOENT.
	filename := c.GetFilename()
	if err := os.MkdirAll(filepath.Dir(filename), 0700); err != nil {
		return fmt.Errorf("unable to create config directory %s: %w", filename, err)
	}

	lock := newFileLock(filename)
	if err := lock.lock(lockTimeout); err != nil {
		return err
	}
	defer func() { _ = lock.unlock() }()
	if err := c.saveLocked(); err != nil {
		return err
	}
	// Persist the disposable cache once, after the config write succeeds and while the
	// lock is still held.
	c.saveCache()
	return nil
}

// saveLocked runs the read-merge-write under an already-held lock.
func (c *Config) saveLocked() error {
	// Resolve flag overrides on the live config so the merge sees the user's
	// real selection, not an ephemeral --context/--environment/--cluster value.
	tempKafkaCluster := c.resolveOverwrittenKafkaCluster()
	defer c.restoreOverwrittenKafkaCluster(tempKafkaCluster)
	tempEnvironment := c.resolveOverwrittenCurrentEnvironment()
	defer c.restoreOverwrittenEnvironment(tempEnvironment)
	tempContext := c.resolveOverwrittenContext()
	defer c.restoreOverwrittenContext(tempContext)

	// No baseline means this config was constructed, not loaded, so there is no
	// common ancestor to merge against and the caller is declaring its state whole
	// (e.g. test config reset). Overwrite directly, matching pre-merge semantics.
	if c.baseline == nil {
		return c.writeWholeConfig()
	}

	disk, err := readConfigFromDisk(c.GetFilename(), c)
	if err != nil {
		// A missing or empty (e.g. a freshly-created temp) file has nothing to
		// preserve: write our state directly, no merge.
		if os.IsNotExist(err) || isEmptyFile(c.GetFilename()) {
			return c.writeWholeConfig()
		}
		return err
	}

	// ours is the live config's persisted state. PreRun left it partially decrypted:
	// every credential secret and the current context's tokens are plaintext, while
	// other contexts' tokens stay encrypted.
	ours := c.deepCopyPersisted()

	// Decrypt the encrypted baseline to ours' representation before diffing, so a
	// secret we did not touch is not mistaken for a local change. We decrypt (a
	// deterministic operation on every platform) rather than re-encrypt ours: Windows
	// DPAPI ciphertext is not reproducible, so an encrypt-based match would flag every
	// secret as changed and reintroduce the very lost-write bug this guards against.
	base := c.baseline.deepCopyPersisted()
	if err := base.decryptToMatch(ours); err != nil {
		return err
	}

	// Captured before threeWayMerge mutates disk in place (it returns disk itself, with its
	// fields overwritten by the merge result) - saveSecretStore needs disk's PRE-merge
	// context set to tell a token's owning context being deleted from a token deliberately
	// cleared elsewhere (see saveSecretStore's own comment on this).
	diskContextNames := make(map[string]bool, len(disk.Contexts))
	for name := range disk.Contexts {
		diskContextNames[name] = true
	}

	merged, err := threeWayMerge(base, ours, disk)
	if err != nil {
		return err
	}
	if err := merged.wireContexts(); err != nil {
		return err
	}

	// merged's nested API-key secrets (GlobalAPIKeys, KafkaClusterConfig.APIKeys) always read
	// as empty at this point - see rehydrateNestedAPIKeySecretPresence - which would otherwise
	// make the Validate() call below delete the key's public id along with its "missing"
	// secret. Rehydrate presence (not correctness: the value is discarded before the write
	// below either way) from c before that happens.
	if err := rehydrateNestedAPIKeySecretPresence(c, merged); err != nil {
		return err
	}

	// Re-encrypt the secrets that ended up plaintext (the ones we changed, taken from
	// ours). Untouched secrets came from disk still encrypted; the guards skip them.
	if err := merged.encryptSecrets(); err != nil {
		return err
	}

	// Validate() normalizes merged in memory and, via Context.Save(), tries to
	// re-persist under the lock we already hold. writing neutralizes that nested
	// Save(); the normalized merged is written just below.
	merged.writing = true
	defer func() { merged.writing = false }()
	if err := merged.Validate(); err != nil {
		return err
	}

	data, err := json.MarshalIndent(merged, "", "  ")
	if err != nil {
		return fmt.Errorf("unable to marshal config: %w", err)
	}

	if err := writeFileAtomic(c.GetFilename(), data); err != nil {
		return err
	}

	// Read from c, not merged: merged's fields came through threeWayMerge's JSON-based
	// deep copies, which drop every json:"-" field (that's how a secret leaves the merge's
	// own diffing, on top of leaving the final config.json write) - so merged never carries
	// a usable secret value. saveSecretStore reads c directly and encrypts what it finds
	// still plaintext, so this is correct regardless of what merged did or didn't preserve.
	if err := c.saveSecretStore(diskContextNames); err != nil {
		return err
	}

	// The next Save's three-way merge diffs baseline against the live config, so the
	// baseline must match the live config for every field this process did not edit,
	// not the merged disk state. merged holds concurrent changes this process pulled in
	// from disk but never wrote back into c; using merged as the ancestor would make a
	// later Save read those untouched fields as local edits and revert the concurrent
	// change. Snapshotting c keeps the ancestor aligned with what this process knows.
	// c itself is left as-is, so its in-memory view of a field it did not touch stays
	// at the loaded value until the process exits; that matches how a load-once CLI
	// already behaves, and only the ancestor advances here.
	c.baseline = c.deepCopyPersisted()
	return nil
}

// writeWholeConfig persists c directly (no merge) and refreshes the baseline from
// the resulting on-disk state, so the next Save has an encrypted ancestor to diff
// against. Used when there is nothing to merge: a missing or empty file, or a
// config that was constructed rather than loaded.
func (c *Config) writeWholeConfig() error {
	if err := c.save(); err != nil {
		return err
	}
	// Refresh the baseline from disk (encrypted) rather than from the live config,
	// which save() has restored to its decrypted form. A read failure here does not
	// undo the successful write, so fall back to the live snapshot.
	if disk, err := readConfigFromDisk(c.GetFilename(), c); err == nil {
		c.baseline = disk
	} else {
		c.snapshotBaseline()
	}
	return nil
}

// encryptSecrets encrypts c's plaintext secrets into their on-disk form: every
// context's auth tokens and all credential API secrets. Save() calls it on the merged
// result to re-encrypt the secrets it took from the (plaintext) live config; secrets
// carried over from disk are already encrypted and the guards skip them. Every context
// is covered (not just the current one) because a merge can leave a plaintext token in
// a context that is not current at write time, and none may reach disk.
//
// Nested secrets (Context.GlobalAPIKeys, KafkaClusterConfig.APIKeys) are intentionally
// not handled here, mirroring decryptToMatch's scope: no command decrypts them in place
// and then saves. They are decrypted only in read paths (ResolveKafkaAPIKey) or encrypted
// before the save (StoreGlobalAPIKey), so they always cross the merge as an intact
// ciphertext/salt/nonce unit and never need re-encryption from a plaintext form.
func (c *Config) encryptSecrets() error {
	for _, ctx := range c.Contexts {
		if ctx.GetState() == nil {
			continue
		}
		st := ctx.GetState()
		if err := c.encryptStateTokensForContext(ctx, st.AuthToken, st.AuthRefreshToken); err != nil {
			return err
		}
	}
	return c.encryptCredentialsAPISecret()
}

// decryptToMatch brings c's secrets into the same representation as ref so the two
// can be diffed field for field. That means both the ciphertext AND its salt/nonce:
// a secret is a (ciphertext, salt, nonce) unit, and aligning only the ciphertext
// leaves the field-wise merge free to pair one source's ciphertext with another's
// salt. PreRun leaves the live config (ref) with every
// credential secret and the current context's tokens in plaintext while other
// contexts' tokens stay encrypted; this decrypts exactly the fields ref holds in
// plaintext. Save() calls it on the baseline, which is encrypted straight from load
// or writeWholeConfig but already in ref's representation after a merged save (whose
// baseline is a live-config snapshot); the guard below only decrypts a field where
// the baseline holds ciphertext ref does not, so an already-aligned field is a no-op.
// It decrypts rather than re-encrypting ref because that is deterministic on every
// platform (Windows DPAPI ciphertext is not), and it never runs Validate (which would
// re-enter Save under the held lock).
func (c *Config) decryptToMatch(ref *Config) error {
	// Decrypt a field only where c holds ciphertext and ref holds plaintext: that is
	// the one case where the two representations differ and c must be brought down to
	// match. Gating on c's own cipher marker matters because a plaintext token that is
	// not cipher-prefixed (e.g. a cloud refresh token, which is never encrypted) would
	// otherwise be fed to Decrypt and fail authentication.
	for name, credential := range c.Credentials {
		refCredential := ref.Credentials[name]
		if credential.APIKeyPair == nil || refCredential == nil || refCredential.APIKeyPair == nil {
			continue
		}
		if isEncryptedSecret(credential.APIKeyPair.Secret) && !isEncryptedSecret(refCredential.APIKeyPair.Secret) {
			if err := credential.APIKeyPair.DecryptSecret(); err != nil {
				return err
			}
			// A secret's salt and nonce are part of its representation, not just its
			// ciphertext. Align them to ref too, so the three-way merge diffs the whole
			// crypto triple as one unit. Leaving c's stale salt here lets the field-wise
			// merge pair disk's ciphertext with a regenerated salt, which then fails GCM
			// authentication on the next load.
			credential.APIKeyPair.Salt = refCredential.APIKeyPair.Salt
			credential.APIKeyPair.Nonce = refCredential.APIKeyPair.Nonce
		}
	}

	for name, state := range c.ContextStates {
		refState := ref.ContextStates[name]
		if state == nil || refState == nil {
			continue
		}
		if isEncryptedSecret(state.AuthToken) && !isEncryptedSecret(refState.AuthToken) {
			if err := state.DecryptAuthToken(name); err != nil {
				return err
			}
		}
		if isEncryptedSecret(state.AuthRefreshToken) && !isEncryptedSecret(refState.AuthRefreshToken) {
			if err := state.DecryptAuthRefreshToken(name); err != nil {
				return err
			}
		}
		// Align salt/nonce to ref for the same reason as credentials above, but only once
		// no encrypted token in c still depends on them: a state's single salt/nonce is
		// shared by both tokens, so clearing it while one stays encrypted would strand
		// that ciphertext.
		if !isEncryptedSecret(state.AuthToken) && !isEncryptedSecret(state.AuthRefreshToken) {
			state.Salt = refState.Salt
			state.Nonce = refState.Nonce
		}
	}
	return nil
}

// isEncryptedSecret reports whether s carries a cipher marker (so it is ciphertext,
// not a plaintext value a command left in place). The ":" is part of the marker.
func isEncryptedSecret(s string) bool {
	return strings.HasPrefix(s, secret.AesGcm+":") || strings.HasPrefix(s, secret.Dpapi+":")
}

// save marshals and atomically writes the live config WITHOUT locking or
// merging. Callers must hold the lock (or knowingly not need it, e.g. writing a
// brand-new default file). Save() is the normal, locked, merging entry point.
//
// The only caller is writeWholeConfig, reached from saveLocked, which has already
// resolved the temporary --context/--environment/--cluster overrides to their
// persisted values. save() must not resolve them again: a second pass runs against
// the now-switched current context and would write one context's flag value onto
// another. Override resolution therefore lives only in the Save()/saveLocked entry.
func (c *Config) save() error {
	var tempAuthToken, tempAuthRefreshToken string
	tempCredentials := map[string]string{}
	if c.Context() != nil {
		tempAuthToken = c.Context().GetState().AuthToken
		tempAuthRefreshToken = c.Context().GetState().AuthRefreshToken
		if err := c.encryptContextStateTokens(tempAuthToken, tempAuthRefreshToken); err != nil {
			return err
		}
		defer c.restoreOverwrittenAuthToken(tempAuthToken)
		defer c.restoreOverwrittenAuthRefreshToken(tempAuthRefreshToken)
	}
	if c.Credentials != nil {
		for name, credential := range c.Credentials {
			if credential.APIKeyPair != nil {
				tempCredentials[name] = credential.APIKeyPair.Secret
			}
		}
		if err := c.encryptCredentialsAPISecret(); err != nil {
			return err
		}
		defer c.restoreOverwrittenCredentials(tempCredentials)
	}

	// See saveLocked: writing neutralizes the nested Save() that Validate()'s
	// normalization triggers, so it does not re-acquire the held lock.
	c.writing = true
	defer func() { c.writing = false }()
	if err := c.Validate(); err != nil {
		return err
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("unable to marshal config: %w", err)
	}

	if err := writeFileAtomic(c.GetFilename(), data); err != nil {
		return err
	}

	// c's own secrets are already ciphertext here (encrypted above, restored to plaintext
	// only by the deferred calls below, which haven't run yet), so saveSecretStore's own
	// encrypt-if-needed step is a no-op for them; it still needs to run to pick up c's
	// saved password and nested API keys, which never round-trip through plaintext at all.
	// nil: this whole-config write (see writeWholeConfig's callers) has nothing to merge
	// secrets.json against either.
	if err := c.saveSecretStore(nil); err != nil {
		return err
	}

	return nil
}

// isEmptyFile reports whether path exists but holds no bytes. This tolerates a
// legacy zero-byte config file (nothing to preserve, so the caller skips the
// merge); a non-empty but corrupt file is not "empty" and is correctly surfaced
// as a hard error by the unmarshal instead. A missing file or any stat error is
// not "empty"; the caller handles absence via os.IsNotExist.
func isEmptyFile(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.Size() == 0
}

func (c *Config) encryptCredentialsAPISecret() error {
	for _, credential := range c.Credentials {
		if credential.APIKeyPair != nil {
			err := credential.APIKeyPair.EncryptSecret()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *Config) encryptContextStateTokens(tempAuthToken, tempAuthRefreshToken string) error {
	return c.encryptStateTokensForContext(c.Context(), tempAuthToken, tempAuthRefreshToken)
}

// encryptStateTokensForContext encrypts ctx's auth tokens in place. It is idempotent:
// an already-encrypted token carries a cipher prefix and matches none of the plaintext
// token shapes below, so it is left untouched. The ctx parameter lets Save() encrypt
// every context's state, not only the current one, so a plaintext token can never
// reach disk under a context that is not current at write time.
func (c *Config) encryptStateTokensForContext(ctx *Context, tempAuthToken, tempAuthRefreshToken string) error {
	state := ctx.GetState()
	if state.Salt == nil || state.Nonce == nil {
		salt, nonce, err := secret.GenerateSaltAndNonce()
		if err != nil {
			return err
		}
		state.Salt = salt
		state.Nonce = nonce
	}

	// Guard every secret.Encrypt call against empty input: DPAPI rejects it on Windows. The auth
	// token regex never matches "", but the belt-and-suspenders check keeps the invariant local.
	if tempAuthToken != "" && regexp.MustCompile(authTokenRegex).MatchString(tempAuthToken) {
		encryptedAuthToken, err := secret.Encrypt(ctx.Name, tempAuthToken, state.Salt, state.Nonce)
		if err != nil {
			return err
		}
		state.AuthToken = encryptedAuthToken
	}

	// The Confluent Gov environment and the Confluent Platform MDS return a refresh token that does not match `authRefreshTokenRegex` and cannot be distinguished from an already encrypted refresh token.
	// We prefix encrypted tokens with "AES/GCM/NoPadding:" on Unix systems and "DPAPI:" on Windows to ensure that they are only encrypted once. The ":" is part of the marker, so a plaintext token merely beginning with the marker word is still encrypted.
	prefix := secret.AesGcm + ":"
	if runtime.GOOS == "windows" {
		prefix = secret.Dpapi + ":"
	}
	isUnencryptedConfluentGov := !strings.HasPrefix(tempAuthRefreshToken, prefix) && (strings.Contains(ctx.PlatformName, "confluentgov.com") || strings.Contains(ctx.PlatformName, "confluentgov-internal.com"))

	isUnencryptedConfluentPlatform := tempAuthRefreshToken != "" && !strings.HasPrefix(tempAuthRefreshToken, prefix) && !ctx.IsCloud(c.IsTest)

	// The gov branch does not itself require a non-empty token, so guard empty explicitly here too:
	// DPAPI rejects empty input on Windows.
	if tempAuthRefreshToken != "" && (regexp.MustCompile(authRefreshTokenRegex).MatchString(tempAuthRefreshToken) || isUnencryptedConfluentGov || isUnencryptedConfluentPlatform) {
		encryptedAuthRefreshToken, err := secret.Encrypt(ctx.Name, tempAuthRefreshToken, state.Salt, state.Nonce)
		if err != nil {
			return err
		}
		state.AuthRefreshToken = encryptedAuthRefreshToken
	}

	return nil
}

// If active Kafka cluster has been overwritten by flag value; if so, replace with previous active kafka
// Return the flag value so that it can be restored after writing to file so that continued execution uses flag value
// This prevents flags from updating state
func (c *Config) resolveOverwrittenKafkaCluster() string {
	ctx := c.Context()
	var tempKafka string
	if c.overwrittenCurrentKafkaCluster != "" && ctx != nil && ctx.KafkaClusterContext != nil {
		if c.overwrittenCurrentKafkaCluster == emptyFieldIndicator {
			c.overwrittenCurrentKafkaCluster = ""
		}
		tempKafka = ctx.KafkaClusterContext.GetActiveKafkaClusterId()
		ctx.KafkaClusterContext.SetActiveKafkaCluster(c.overwrittenCurrentKafkaCluster)
	}
	return tempKafka
}

// Restore the flag cluster back into the struct so that it is used for any execution after Save()
func (c *Config) restoreOverwrittenKafkaCluster(tempKafkaCluster string) {
	if tempKafkaCluster != "" {
		c.Context().KafkaClusterContext.SetActiveKafkaCluster(tempKafkaCluster)
	}
}

func (c *Config) restoreOverwrittenAuthToken(tempAuthToken string) {
	if tempAuthToken != "" {
		c.Context().GetState().AuthToken = tempAuthToken
	}
}

func (c *Config) restoreOverwrittenAuthRefreshToken(tempAuthRefreshToken string) {
	if tempAuthRefreshToken != "" {
		c.Context().GetState().AuthRefreshToken = tempAuthRefreshToken
	}
}

func (c *Config) restoreOverwrittenCredentials(tempApiSecrets map[string]string) {
	for name, secret := range tempApiSecrets {
		if secret != "" {
			c.Credentials[name].APIKeyPair.Secret = secret
		}
	}
}

// Switch the initial config context back into the struct so that it is saved and not the flag value
// Return the overwriting flag context value so that it can be restored after writing the file
func (c *Config) resolveOverwrittenContext() string {
	var tempContext string
	if c.overwrittenCurrentContext != "" && c != nil {
		if c.overwrittenCurrentContext == emptyFieldIndicator {
			c.overwrittenCurrentContext = ""
		}
		tempContext = c.CurrentContext
		c.CurrentContext = c.overwrittenCurrentContext
	}
	return tempContext
}

// Restore the flag context back into the struct so that it is used for any execution after Save()
func (c *Config) restoreOverwrittenContext(tempContext string) {
	if tempContext != "" {
		c.CurrentContext = tempContext
	}
}

// Switch the initial config account back into the struct so that it is saved and not the flag value
// Return the overwriting flag account value so that it can be restored after writing the file
func (c *Config) resolveOverwrittenCurrentEnvironment() string {
	var tempEnvironment string
	if c.overwrittenCurrentEnvironment != "" {
		tempEnvironment = c.Context().GetCurrentEnvironment()
		c.Context().SetCurrentEnvironment(c.overwrittenCurrentEnvironment)
	}
	return tempEnvironment
}

// Restore the flag account back into the struct so that it is used for any execution after Save()
func (c *Config) restoreOverwrittenEnvironment(id string) {
	if id != "" {
		c.Context().SetCurrentEnvironment(id)
	}
}

func (c *Config) Validate() error {
	// Validate that current context exists.
	if c.CurrentContext != "" {
		if _, ok := c.Contexts[c.CurrentContext]; !ok {
			log.CliLogger.Trace("current context does not exist")
			return errors.NewCorruptedConfigError(`the current context "%s" does not exist`, c.CurrentContext, c.Filename)
		}
	}

	// Validate that every context:
	// 1. Has no hanging references between the context and the config.
	// 2. Is mapped by name correctly in the config.
	for _, context := range c.Contexts {
		if err := context.validate(); err != nil {
			log.CliLogger.Trace("context validation error")
			return err
		}
		if _, ok := c.Credentials[context.CredentialName]; !ok {
			log.CliLogger.Trace("unspecified credential error")
			return errors.NewCorruptedConfigError(errors.UnspecifiedCredentialErrorMsg, context.Name, c.Filename)
		}
		if _, ok := c.Platforms[context.PlatformName]; !ok {
			log.CliLogger.Trace("unspecified platform error")
			return errors.NewCorruptedConfigError(errors.UnspecifiedPlatformErrorMsg, context.Name, c.Filename)
		}
		if _, ok := c.ContextStates[context.Name]; !ok {
			c.ContextStates[context.Name] = new(ContextState)
		}
		if !c.IsTest && !reflect.DeepEqual(*c.ContextStates[context.Name], *context.State) {
			log.CliLogger.Tracef("state of context %s in config does not match actual state of context", context.Name)
			return errors.NewCorruptedConfigError(`context state mismatch for context "%s"`, context.Name, c.Filename)
		}
	}

	// Validate that all context states are mapped to an existing context.
	for contextName := range c.ContextStates {
		if _, ok := c.Contexts[contextName]; !ok {
			log.CliLogger.Trace("context state mapped to nonexistent context")
			return errors.NewCorruptedConfigError(`context state mapping error for context "%s"`, contextName, c.Filename)
		}
	}

	return nil
}

// DeleteContext deletes the specified context, and returns an error if it's not found.
func (c *Config) DeleteContext(name string) error {
	if _, err := c.FindContext(name); err != nil {
		return err
	}
	delete(c.Contexts, name)
	delete(c.ContextStates, name)

	if name == c.CurrentContext {
		c.CurrentContext = ""
	}

	return c.Save()
}

// FindContext finds a context by name, and returns nil if not found.
func (c *Config) FindContext(name string) (*Context, error) {
	context, ok := c.Contexts[name]
	if !ok {
		return nil, fmt.Errorf(errors.ContextDoesNotExistErrorMsg, name)
	}
	return context, nil
}

func (c *Config) AddContext(name, platformName, credentialName string, kafkaClusters map[string]*KafkaClusterConfig, kafka string, kafkaEndpoint string, state *ContextState, organizationId, environmentId string, isMFA bool) error {
	if _, ok := c.Contexts[name]; ok {
		return fmt.Errorf(errors.ContextAlreadyExistsErrorMsg, name)
	}

	credential, ok := c.Credentials[credentialName]
	if !ok {
		return fmt.Errorf(`credential "%s" not found`, credentialName)
	}

	platform, ok := c.Platforms[platformName]
	if !ok {
		return fmt.Errorf(`platform "%s" not found`, platformName)
	}

	ctx, err := newContext(name, platform, credential, kafkaClusters, kafka, kafkaEndpoint, state, c, organizationId, environmentId, isMFA)
	if err != nil {
		return err
	}

	c.Contexts[name] = ctx
	c.ContextStates[name] = ctx.State

	if err := c.Validate(); err != nil {
		return err
	}

	return c.Save()
}

// CreateContext creates a new context.
func (c *Config) CreateContext(name, bootstrapURL, apiKey, apiSecret string) error {
	apiKeyPair := &APIKeyPair{
		Key:    apiKey,
		Secret: apiSecret,
	}

	// Hardcoded for now, since username/password isn't implemented yet.
	credential := &Credential{
		APIKeyPair:     apiKeyPair,
		CredentialType: APIKey,
		Name:           fmt.Sprintf("%s-%s", APIKey, apiKey),
	}

	if err := c.SaveCredential(credential); err != nil {
		return err
	}

	// Inject credential and platforms name for now, until users can provide custom names.
	platform := &Platform{
		Server: bootstrapURL,
		Name:   strings.TrimPrefix(bootstrapURL, "https://"),
	}

	if err := c.SavePlatform(platform); err != nil {
		return err
	}

	kafkaClusterCfg := &KafkaClusterConfig{
		ID:        "anonymous-id",
		Name:      "anonymous-cluster",
		Bootstrap: bootstrapURL,
		APIKeys:   map[string]*APIKeyPair{apiKey: apiKeyPair},
		APIKey:    apiKey,
	}
	kafkaClusters := map[string]*KafkaClusterConfig{kafkaClusterCfg.ID: kafkaClusterCfg}

	return c.AddContext(name, platform.Name, credential.Name, kafkaClusters, kafkaClusterCfg.ID, "", nil, "", "", false)
}

// UseContext sets the current context, if it exists.
func (c *Config) UseContext(name string) error {
	if _, err := c.FindContext(name); err != nil {
		return err
	}
	c.CurrentContext = name
	return c.Save()
}

func (c *Config) SaveCredential(credential *Credential) error {
	if credential.Name == "" {
		return fmt.Errorf("credential must have a name")
	}
	c.Credentials[credential.Name] = credential
	return c.Save()
}

func (c *Config) SaveLoginCredential(ctxName string, loginCredential *LoginCredential) error {
	if ctxName == "" {
		return fmt.Errorf("saved credential must match a context")
	}
	c.SavedCredentials[ctxName] = loginCredential
	return c.Save()
}

func (c *Config) SavePlatform(platform *Platform) error {
	if platform.Name == "" {
		return fmt.Errorf("platform must have a name")
	}
	c.Platforms[platform.Name] = platform
	return c.Save()
}

// Context returns the current context.
func (c *Config) Context() *Context {
	if c == nil {
		return nil
	}
	return c.Contexts[c.CurrentContext]
}

// CredentialType returns the credential type used in the current context: API key, username & password, or neither.
func (c *Config) CredentialType() CredentialType {
	if c.hasAPIKeyLogin() {
		return APIKey
	}

	if c.HasBasicLogin() {
		return Username
	}

	return None
}

// hasAPIKeyLogin returns true if the user has valid API Key credentials.
func (c *Config) hasAPIKeyLogin() bool {
	return c.Context().GetCredentialType() == APIKey
}

// HasBasicLogin returns true if the user has valid username & password credentials.
func (c *Config) HasBasicLogin() bool {
	ctx := c.Context()
	if ctx == nil {
		return false
	}

	if c.IsCloudLogin() {
		return ctx.HasLogin() && ctx.GetCurrentEnvironment() != ""
	} else {
		return ctx.HasLogin()
	}
}

func (c *Config) GetFilename() string {
	if c.Filename == "" {
		c.Filename = GetDefaultFilename()
	}
	return c.Filename
}

// StateDirName is the name of the CLI's state directory within the user's home directory. It is
// scoped to the build's release channel so a production install, a prerelease, and a local build
// cannot read or overwrite each other's state. A stable build returns ".confluent", unchanged.
func StateDirName() string {
	return ".confluent" + pversion.ProcessChannel().StateDirSuffix()
}

// StateDir is the absolute path of the CLI's state directory.
func StateDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", errors.NewErrorWithSuggestions(
			fmt.Sprintf("unable to determine the home directory holding the CLI's state: %v", err),
			"Set the `HOME` environment variable (`USERPROFILE` on Windows) to a writable directory.",
		)
	}
	return filepath.Join(home, StateDirName()), nil
}

// GetDefaultFilename swallows a missing home directory because it backs a flag default built at
// command-construction time, where there is no error to return. Prefer StateDir where you can.
func GetDefaultFilename() string {
	return stateDirPath("config.json")
}

func (c *Config) CheckIsOnPremLogin() error {
	ctx := c.Context()
	if ctx != nil && ctx.PlatformName != "" {
		if !c.isCloud() {
			return nil
		} else {
			return RunningOnPremCommandInCloudErr
		}
	}
	return RequireOnPremLoginErr
}

func (c *Config) CheckIsCloudLogin() error {
	if !c.isCloud() {
		return RequireCloudLoginErr
	}

	if c.isContextStatePresent() && c.isOrgSuspended() {
		if c.isOrgPauseTrialSuspension() {
			return RequireCloudLoginPauseTrialOrgUnsuspendedErr
		} else if c.isLoginBlockedByOrgSuspension() {
			return RequireCloudLoginOrgUnsuspendedErr
		} else {
			return RequireCloudLoginFreeTrialEndedOrgUnsuspendedErr
		}
	}

	return nil
}

func (c *Config) CheckIsCloudLoginAllowFreeTrialEnded() error {
	if !c.isCloud() {
		return RequireCloudLoginErr
	}

	if c.isContextStatePresent() && c.isLoginBlockedByOrgSuspension() {
		return RequireCloudLoginOrgUnsuspendedErr
	}

	return nil
}

func (c *Config) CheckIsCloudLoginOrOnPremLogin() error {
	isCloudLoginErr := c.CheckIsCloudLogin()
	isOnPremLoginErr := c.CheckIsOnPremLogin()

	if !(isCloudLoginErr == nil || isOnPremLoginErr == nil) {
		// return org suspension errors
		if isCloudLoginErr != nil && isCloudLoginErr != RequireCloudLoginErr {
			return isCloudLoginErr
		}
		return RequireCloudLoginOrOnPremErr
	}

	return nil
}

func (c *Config) CheckIsNonAPIKeyCloudLogin() error {
	isCloudLoginErr := c.CheckIsCloudLogin()

	if !(c.CredentialType() != APIKey && isCloudLoginErr == nil) {
		// return org suspension errors
		if isCloudLoginErr != nil && isCloudLoginErr != RequireCloudLoginErr {
			return isCloudLoginErr
		}
		return RequireNonAPIKeyCloudLoginErr
	}

	return nil
}

func (c *Config) CheckIsNonAPIKeyCloudLoginOrOnPremLogin() error {
	isNonAPIKeyCloudLoginErr := c.CheckIsNonAPIKeyCloudLogin()
	isOnPremLoginErr := c.CheckIsOnPremLogin()

	if !(isNonAPIKeyCloudLoginErr == nil || isOnPremLoginErr == nil) {
		// return org suspension errors
		if isNonAPIKeyCloudLoginErr != nil && isNonAPIKeyCloudLoginErr != RequireCloudLoginErr && isNonAPIKeyCloudLoginErr != RequireNonAPIKeyCloudLoginErr {
			return isNonAPIKeyCloudLoginErr
		}
		return RequireNonAPIKeyCloudLoginOrOnPremLoginErr
	}

	return nil
}

func (c *Config) CheckIsCloudLogout() error {
	if c.isCloud() {
		return RequireCloudLogout
	}
	return nil
}

func (c *Config) IsCloudLogin() bool {
	return c.CheckIsCloudLogin() == nil
}

func (c *Config) HasGovHostname() bool {
	ctx := c.Context()
	if ctx == nil {
		return false
	}

	for _, hostname := range []string{"confluentgov-internal.com", "confluentgov.com"} {
		if strings.Contains(ctx.PlatformName, hostname) {
			return true
		}
	}

	return false
}

func (c *Config) IsOnPremLogin() bool {
	return c.CheckIsOnPremLogin() == nil
}

func (c *Config) isCloud() bool {
	ctx := c.Context()
	if ctx == nil {
		return false
	}

	return ctx.IsCloud(c.IsTest)
}

func (c *Config) isContextStatePresent() bool {
	ctx := c.Context()
	if ctx == nil {
		return false
	}

	if ctx.GetOrganization() == nil {
		log.CliLogger.Trace("current context state is not set up properly for checking org suspension status")
		return false
	}

	return true
}

func (c *Config) isOrgSuspended() bool {
	return utils.IsOrgSuspended(c.Context().GetSuspensionStatus())
}

func (c *Config) isLoginBlockedByOrgSuspension() bool {
	return utils.IsLoginBlockedByOrgSuspension(c.Context().GetSuspensionStatus())
}

func (c *Config) isOrgPauseTrialSuspension() bool {
	return utils.IsOrgPauseTrialSuspended(c.Context().GetSuspensionStatus())
}

// Parse `--context` flag value into config struct
// Call ParseFlagsIntoContext which handles environment and cluster flags
func (c *Config) ParseFlagsIntoConfig(cmd *cobra.Command) error {
	if context, _ := cmd.Flags().GetString("context"); context != "" {
		if _, err := c.FindContext(context); err != nil {
			return err
		}
		c.SetOverwrittenCurrentContext(c.CurrentContext)
		c.CurrentContext = context
	}

	return nil
}
