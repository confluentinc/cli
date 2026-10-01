package config

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	perrors "github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/secret"
	pversion "github.com/confluentinc/cli/v4/pkg/version"
)

// legacyConfigFileCase is one readLegacyConfigFile scenario: a path, whether it should be found,
// and the data it should return.
type legacyConfigFileCase struct {
	name      string
	path      string
	wantFound bool
	wantData  []byte
}

func TestReadLegacyConfigFile(t *testing.T) {
	dir := t.TempDir()

	zero := filepath.Join(dir, "zero.json")
	require.NoError(t, os.WriteFile(zero, nil, 0600))

	nonEmpty := filepath.Join(dir, "config.json")
	want := []byte(`{"current_context":"ctx1"}`)
	require.NoError(t, os.WriteFile(nonEmpty, want, 0600))

	tests := []legacyConfigFileCase{
		{name: "missing", path: filepath.Join(dir, "missing.json"), wantFound: false},
		{name: "directory", path: dir, wantFound: false},
		{name: "zero-byte file", path: zero, wantFound: false},
		{name: "non-empty file", path: nonEmpty, wantFound: true, wantData: want},
	}

	// readLegacyConfigFile must follow symlinks, so exercise it against a symlink to a file, a
	// dangling symlink, and a symlink to a directory. Symlink creation can fail without elevated
	// privilege (notably on Windows), so skip these cases rather than fail the whole test.
	symlinkToFile := filepath.Join(dir, "symlink-to-file.json")
	if err := os.Symlink(nonEmpty, symlinkToFile); err != nil {
		t.Skip("creating a symlink is not supported on this platform:", err)
	}
	tests = append(tests, legacyConfigFileCase{name: "symlink to file", path: symlinkToFile, wantFound: true, wantData: want})

	danglingSymlink := filepath.Join(dir, "dangling-symlink.json")
	require.NoError(t, os.Symlink(filepath.Join(dir, "does-not-exist.json"), danglingSymlink))
	tests = append(tests, legacyConfigFileCase{name: "dangling symlink", path: danglingSymlink, wantFound: false})

	symlinkToDir := filepath.Join(dir, "symlink-to-dir.json")
	require.NoError(t, os.Symlink(dir, symlinkToDir))
	tests = append(tests, legacyConfigFileCase{name: "symlink to directory", path: symlinkToDir, wantFound: false})

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			data, info, err := readLegacyConfigFile(tc.path)

			require.NoError(t, err)
			require.Equal(t, tc.wantFound, info != nil)
			require.Equal(t, tc.wantData, data)
		})
	}
}

func TestApplyLegacyConfig_Malformed(t *testing.T) {
	tests := []struct {
		name string
		data string
	}{
		{
			name: "syntax error",
			data: `{"contexts":`,
		},
		{
			// "current_context" decodes before encoding/json hits the type mismatch on
			// "contexts" and skips it, so a naive decode straight onto c would leave
			// c.CurrentContext set even though the call returns an error.
			name: "live decode type mismatch after a field is already set",
			data: `{"current_context":"x","contexts":5}`,
		},
		{
			// Salt is json:"-" on the live ContextState, so the live decode never looks at
			// it and succeeds; only the overlay decode (which does have a "salt" field)
			// hits the invalid base64.
			name: "overlay-only failure",
			data: `{"context_states":{"ctx1":{"salt":"!!"}}}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			setTestHome(t, t.TempDir())
			c := New()
			path := filepath.Join(t.TempDir(), "config.json")

			err := c.applyLegacyConfig(path, []byte(tc.data))

			require.ErrorContains(t, err, readConfigurationFileErrorPrefix)
			require.ErrorContains(t, err, path)
			require.Equal(t, New(), c)
		})
	}
}

// legacyFixture builds a v4 config.json (inline, not a committed fixture) carrying a distinct
// marker string in every secret path, plus last_update_check_at, a context's feature_flags, and
// non-secret "pinned" selections (current_environment, the top-level and per-environment
// active_kafka) that must survive the decode untouched.
func legacyFixture() []byte {
	return legacyFixtureWithSecretPrefix("")
}

// legacyFixtureWithSecretPrefix is legacyFixture with p prepended to every secret marker (not to
// salts, nonces, or feature flags).
func legacyFixtureWithSecretPrefix(p string) []byte {
	b64 := func(marker string) string { return base64.StdEncoding.EncodeToString([]byte(marker)) }

	return []byte(fmt.Sprintf(`{
		"current_context": "ctx1",
		"last_update_check_at": "2020-01-02T03:04:05Z",
		"platforms": {
			"platform1": {"name": "platform1", "server": "https://example.com"}
		},
		"credentials": {
			"cred1": {
				"name": "cred1",
				"username": "",
				"api_key_pair": {
					"api_key": "key1",
					"api_secret": "%s",
					"salt": "%s",
					"nonce": "%s"
				},
				"credential_type": 1
			}
		},
		"contexts": {
			"ctx1": {
				"name": "ctx1",
				"machine_name": "ctx1",
				"platform": "platform1",
				"credential": "cred1",
				"current_environment": "env1",
				"active_global_api_key": "global1",
				"feature_flags": {
					"values": {"marker": "%s"},
					"ccloud_values": {},
					"last_update_time": 42
				},
				"global_api_keys": {
					"global1": {
						"api_key": "global1",
						"api_secret": "%s"
					}
				},
				"kafka_cluster_context": {
					"environment_context": true,
					"active_kafka": "cluster1",
					"kafka_cluster_configs": {
						"cluster1": {
							"id": "cluster1",
							"name": "cluster1",
							"bootstrap_servers": "b1",
							"api_key": "key1",
							"api_keys": {
								"key1": {"api_key": "key1", "api_secret": "%s"}
							}
						}
					},
					"kafka_environment_contexts": {
						"env1": {
							"active_kafka": "cluster2",
							"active_kafka_endpoint": "",
							"kafka_cluster_infos": {
								"cluster2": {
									"id": "cluster2",
									"name": "cluster2",
									"bootstrap_servers": "b2",
									"api_key": "key1",
									"api_keys": {
										"key1": {"api_key": "key1", "api_secret": "%s"}
									}
								}
							}
						}
					}
				},
				"schema_registry_clusters": {
					"sr1": {
						"id": "sr1",
						"schema_registry_endpoint": "https://sr",
						"schema_registry_credentials": {"api_key": "sr-key", "api_secret": "%s"}
					}
				}
			}
		},
		"context_states": {
			"ctx1": {
				"auth_token": "%s",
				"auth_refresh_token": "%s",
				"salt": "%s",
				"nonce": "%s"
			}
		},
		"saved_credentials": {
			"ctx1": {
				"is_cloud": true,
				"url": "",
				"username": "test-user",
				"encrypted_password": "%s",
				"salt": "%s",
				"nonce": "%s"
			}
		}
	}`,
		p+"MARKER-cred-api-secret", b64("MARKER-cred-salt"), b64("MARKER-cred-nonce"),
		"MARKER-feature-flags-value",
		p+"MARKER-global-api-secret",
		p+"MARKER-kafka-cluster-configs-secret",
		p+"MARKER-kafka-env-contexts-secret",
		p+"MARKER-sr-credential-secret",
		p+"MARKER-auth-token", p+"MARKER-auth-refresh-token", b64("MARKER-state-salt"), b64("MARKER-state-nonce"),
		p+"MARKER-saved-password", b64("MARKER-saved-salt"), b64("MARKER-saved-nonce"),
	))
}

func TestApplyLegacyConfig_CopiesSecretsVerbatim(t *testing.T) {
	setTestHome(t, t.TempDir())
	c := New()
	path := filepath.Join(t.TempDir(), "config.json")

	err := c.applyLegacyConfig(path, legacyFixture())
	require.NoError(t, err)

	require.NotNil(t, c.LastUpdateCheckAt)
	require.Equal(t, "2020-01-02T03:04:05Z", c.LastUpdateCheckAt.UTC().Format("2006-01-02T15:04:05Z"))

	cred := c.Credentials["cred1"]
	require.Equal(t, "MARKER-cred-api-secret", cred.APIKeyPair.Secret)
	require.Equal(t, []byte("MARKER-cred-salt"), cred.APIKeyPair.Salt)
	require.Equal(t, []byte("MARKER-cred-nonce"), cred.APIKeyPair.Nonce)

	ctx := c.Contexts["ctx1"]
	require.NotNil(t, ctx.FeatureFlags)
	require.Equal(t, "MARKER-feature-flags-value", ctx.FeatureFlags.CliValues["marker"])

	require.Equal(t, "MARKER-global-api-secret", ctx.GlobalAPIKeys["global1"].Secret)

	kcc := ctx.KafkaClusterContext.KafkaClusterConfigs["cluster1"]
	require.Equal(t, "MARKER-kafka-cluster-configs-secret", kcc.APIKeys["key1"].Secret)

	envCtx := ctx.KafkaClusterContext.KafkaEnvContexts["env1"]
	envCfg := envCtx.KafkaClusterConfigs["cluster2"]
	require.Equal(t, "MARKER-kafka-env-contexts-secret", envCfg.APIKeys["key1"].Secret)

	require.Equal(t, "MARKER-sr-credential-secret", ctx.SchemaRegistryClusters["sr1"].SrCredentials.Secret)

	state := c.ContextStates["ctx1"]
	require.Equal(t, "MARKER-auth-token", state.AuthToken)
	require.Equal(t, "MARKER-auth-refresh-token", state.AuthRefreshToken)
	require.Equal(t, []byte("MARKER-state-salt"), state.Salt)
	require.Equal(t, []byte("MARKER-state-nonce"), state.Nonce)

	saved := c.SavedCredentials["ctx1"]
	require.Equal(t, "MARKER-saved-password", saved.EncryptedPassword)
	require.Equal(t, []byte("MARKER-saved-salt"), saved.Salt)
	require.Equal(t, []byte("MARKER-saved-nonce"), saved.Nonce)

	// non-secret "pinned" selections and other plain fields survive the decode untouched.
	require.Equal(t, "ctx1", c.CurrentContext)
	require.Equal(t, "platform1", ctx.PlatformName)
	require.Equal(t, "env1", ctx.CurrentEnvironment)
	require.Equal(t, "cluster1", ctx.KafkaClusterContext.ActiveKafkaCluster)
	require.Equal(t, "cluster2", envCtx.ActiveKafkaCluster)
	require.Equal(t, "global1", ctx.ActiveGlobalAPIKey)
}

func TestApplyLegacyConfig_StatefulFixtures(t *testing.T) {
	tests := []struct {
		fixture string
		// kafkaViaEnvContext is true when the fixture's Kafka cluster config (and its nested
		// API key) live under kafka_environment_contexts rather than kafka_cluster_configs.
		kafkaViaEnvContext bool
	}{
		{fixture: "stateful_cloud.json", kafkaViaEnvContext: true},
		{fixture: "stateful_onprem.json", kafkaViaEnvContext: false},
	}

	for _, tc := range tests {
		t.Run(tc.fixture, func(t *testing.T) {
			home := t.TempDir()
			setTestHome(t, home)
			c := New()

			src, err := os.ReadFile(filepath.Join("test_json", tc.fixture))
			require.NoError(t, err)

			path := filepath.Join(home, "config.json")
			require.NoError(t, os.WriteFile(path, src, 0600))

			data, info, err := readLegacyConfigFile(path)
			require.NoError(t, err)
			require.NotNil(t, info)

			require.NoError(t, c.applyLegacyConfig(path, data))

			ctx := c.Contexts["my-context"]
			require.Equal(t, "def-secret-456", c.Credentials["api-key-abc-key-123"].APIKeyPair.Secret)

			var kafkaKey *APIKeyPair
			if tc.kafkaViaEnvContext {
				kafkaKey = ctx.KafkaClusterContext.KafkaEnvContexts["env-123456"].KafkaClusterConfigs["anonymous-id"].APIKeys["abc-key-123"]
			} else {
				kafkaKey = ctx.KafkaClusterContext.KafkaClusterConfigs["anonymous-id"].APIKeys["abc-key-123"]
			}
			require.Equal(t, "def-secret-456", kafkaKey.Secret)

			state := c.ContextStates["my-context"]
			require.Equal(t, "eyJ.eyJ.abc", state.AuthToken)
			require.Equal(t, "v1.abc", state.AuthRefreshToken)

			require.Equal(t, "encrypted-password", c.SavedCredentials["my-context"].EncryptedPassword)
		})
	}
}

// legacyOverlayFor maps a live type name to the overlay type covering its retagged fields, for
// every live type the walk in TestLegacyOverlay_CoversEveryRetaggedField can reach.
var legacyOverlayFor = map[string]reflect.Type{
	"Config":          reflect.TypeOf(legacyConfig{}),
	"Context":         reflect.TypeOf(legacyContext{}),
	"ContextState":    reflect.TypeOf(legacyContextState{}),
	"LoginCredential": reflect.TypeOf(legacyLoginCredential{}),
	"APIKeyPair":      reflect.TypeOf(legacyAPIKeyPair{}),
}

// legacyOverlayAllowlist lists json:"-" fields under a persisted path that were never persisted
// in v4 either (runtime/self references, or fields HEAD already excluded), derived from
// `git show main:pkg/config/*.go`: these carry no overlay entry and none is expected.
var legacyOverlayAllowlist = map[string]bool{
	"Config.DisableUpdates":       true,
	"Config.Filename":             true,
	"Config.IsTest":               true,
	"Config.Version":              true,
	"Context.Platform":            true,
	"Context.Credential":          true,
	"Context.State":               true,
	"Context.Config":              true,
	"KafkaClusterContext.Context": true,
}

// TestLegacyOverlay_CoversEveryRetaggedField walks the live Config type graph, following only
// persisted (non json:"-") fields, and checks that every json:"-" field it finds under
// pkg/config's own types has either an allowlist entry or a same-named field on a registered
// overlay type; it also fails if an allowlist entry is never visited, catching a stale entry
// that could mask real drift elsewhere. It only checks that a retagged field's TYPE has a
// matching overlay field, not that the overlay's wrapper structs reach every holder along the
// way (e.g. both KafkaClusterConfigs and KafkaEnvContexts);
// TestApplyLegacyConfig_CopiesSecretsVerbatim covers that with real markers.
func TestLegacyOverlay_CoversEveryRetaggedField(t *testing.T) {
	pkgPath := reflect.TypeOf(Config{}).PkgPath()
	visited := map[reflect.Type]bool{}
	visitedAllowlist := map[string]bool{}

	var walk func(rt reflect.Type)
	walk = func(rt reflect.Type) {
		if rt.Kind() != reflect.Struct || rt.PkgPath() != pkgPath || visited[rt] {
			return
		}
		visited[rt] = true

		for i := 0; i < rt.NumField(); i++ {
			field := rt.Field(i)
			if !field.IsExported() {
				continue
			}

			tag := field.Tag.Get("json")
			if tag == "-" {
				key := rt.Name() + "." + field.Name
				if legacyOverlayAllowlist[key] {
					visitedAllowlist[key] = true
					continue
				}

				overlay, ok := legacyOverlayFor[rt.Name()]
				require.True(t, ok, "no legacy overlay registered for %s (retagged field %s)", rt.Name(), key)

				_, found := overlay.FieldByName(field.Name)
				require.True(t, found, "legacy overlay %s is missing field %s for %s", overlay.Name(), field.Name, key)
				continue
			}

			elem := field.Type
			for elem.Kind() == reflect.Ptr || elem.Kind() == reflect.Map || elem.Kind() == reflect.Slice {
				elem = elem.Elem()
			}
			walk(elem)
		}
	}

	walk(reflect.TypeOf(Config{}))

	for key := range legacyOverlayAllowlist {
		require.True(t, visitedAllowlist[key], "stale allowlist entry %s: never visited by the walk", key)
	}
}

// nativeCipherPrefix is the cipher marker this platform's save path treats as already encrypted
// for every secret (an on-prem refresh token is only recognized by the native one), so a marker
// carrying it is stored as-is instead of being encrypted.
func nativeCipherPrefix() string {
	if runtime.GOOS == "windows" {
		return secret.Dpapi + ":"
	}
	return secret.AesGcm + ":"
}

// cipherLegacyFixture is the all-markers legacy file with every secret shaped like ciphertext.
// Its Kafka cluster context is an environment context, so only the kafka_environment_contexts
// key is reachable by the save path (allKafkaClusterConfigs), not the kafka_cluster_configs one.
func cipherLegacyFixture() []byte {
	return legacyFixtureWithSecretPrefix(nativeCipherPrefix())
}

// directKafkaCipherLegacyFixture is cipherLegacyFixture with a non-environment Kafka cluster
// context, the shape whose kafka_cluster_configs key the save path reaches instead.
func directKafkaCipherLegacyFixture() []byte {
	return bytes.Replace(cipherLegacyFixture(), []byte(`"environment_context": true`), []byte(`"environment_context": false`), 1)
}

// setTestChannel runs the rest of the test on ch, restoring the previous channel afterward.
func setTestChannel(t *testing.T, ch pversion.Channel) {
	t.Helper()
	prev := pversion.ProcessChannel()
	pversion.SetProcessChannel(ch)
	t.Cleanup(func() { pversion.SetProcessChannel(prev) })
}

// countLegacyMigrations counts afterLegacyMigration calls for the rest of the test.
func countLegacyMigrations(t *testing.T) *atomic.Int32 {
	t.Helper()
	var n atomic.Int32
	afterLegacyMigration = func() { n.Add(1) }
	t.Cleanup(func() { afterLegacyMigration = func() {} })
	return &n
}

// seedLegacyConfig writes data as home's v4 config.json and returns its path.
func seedLegacyConfig(t *testing.T, home string, data []byte) string {
	t.Helper()
	path := filepath.Join(home, ".confluent", "config.json")
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0700))
	require.NoError(t, os.WriteFile(path, data, 0600))
	return path
}

// writeTestStore writes a store file directly, bypassing the save path.
func writeTestStore(t *testing.T, path, data string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0700))
	require.NoError(t, os.WriteFile(path, []byte(data), 0600))
}

// migrationBackupPath is where the running channel's backup of the legacy file lands.
func migrationBackupPath(home string) string {
	return filepath.Join(home, StateDirName(), "config.json.v4-backup")
}

// migrationMarkerPath is where the running channel's in-progress migration marker lands.
func migrationMarkerPath(home string) string {
	return filepath.Join(home, StateDirName(), "config.json.v4-migrating")
}

// stableBackupPath is the Stable channel's backup, which every channel can see.
func stableBackupPath(home string) string {
	return filepath.Join(home, ".confluent", "config.json.v4-backup")
}

// onLegacyConfigStat runs fn between the migration's stat and read of the legacy file, for the
// rest of the test.
func onLegacyConfigStat(t *testing.T, fn func(path string)) {
	t.Helper()
	afterLegacyConfigStat = fn
	t.Cleanup(func() { afterLegacyConfigStat = func(string) {} })
}

// loadQuietly loads c with stderr captured, keeping migration announcements out of test output.
func loadQuietly(t *testing.T, c *Config) error {
	t.Helper()
	var err error
	captureStderr(t, func() { err = c.Load() })
	return err
}

// newStableMigrationTest isolates HOME on the Stable channel and counts migrations.
func newStableMigrationTest(t *testing.T) (string, *atomic.Int32) {
	t.Helper()
	home := t.TempDir()
	setTestHome(t, home)
	setTestChannel(t, pversion.Stable)
	return home, countLegacyMigrations(t)
}

// kafkaMarkerFor names the one nested Kafka key a normal save persists for the fixture's shape.
func kafkaMarkerFor(envContext bool) (string, string) {
	if envContext {
		return "cluster2", "MARKER-kafka-env-contexts-secret"
	}
	return "cluster1", "MARKER-kafka-cluster-configs-secret"
}

// requireMigratedSecretsOnDisk checks that secrets.json holds every cipher marker, verbatim, under
// the key a normal Save uses: tokens and passwords by context name, the rest by identity.
// envContext names the fixture's Kafka shape, which decides the one nested Kafka key saved.
func requireMigratedSecretsOnDisk(t *testing.T, envContext bool) {
	t.Helper()
	p := nativeCipherPrefix()

	file, err := readSecretFileFromDisk(SecretsFilename())
	require.NoError(t, err)

	token := file.Tokens["ctx1"]
	require.NotNil(t, token)
	require.Equal(t, p+"MARKER-auth-token", token.AuthToken)
	require.Equal(t, p+"MARKER-auth-refresh-token", token.AuthRefreshToken)
	require.Equal(t, []byte("MARKER-state-salt"), token.Salt)
	require.Equal(t, []byte("MARKER-state-nonce"), token.Nonce)

	password := file.Passwords["ctx1"]
	require.NotNil(t, password)
	require.Equal(t, p+"MARKER-saved-password", password.Password)
	require.Equal(t, []byte("MARKER-saved-salt"), password.Salt)
	require.Equal(t, []byte("MARKER-saved-nonce"), password.Nonce)

	rec := file.Secrets["cred1"]
	require.NotNil(t, rec)
	require.Equal(t, p+"MARKER-cred-api-secret", rec.Secret)
	require.Equal(t, []byte("MARKER-cred-salt"), rec.SecretSalt)
	require.Equal(t, []byte("MARKER-cred-nonce"), rec.SecretNonce)
	require.Contains(t, rec.GlobalAPIKeys, "global1")
	require.Equal(t, p+"MARKER-global-api-secret", rec.GlobalAPIKeys["global1"].Secret)
	require.Contains(t, rec.SchemaRegistryCredentials["sr1"], "sr-key")
	require.Equal(t, p+"MARKER-sr-credential-secret", rec.SchemaRegistryCredentials["sr1"]["sr-key"].Secret)

	cluster, marker := kafkaMarkerFor(envContext)
	require.Contains(t, rec.KafkaAPIKeys[cluster], "key1")
	require.Equal(t, p+marker, rec.KafkaAPIKeys[cluster]["key1"].Secret)
}

// requireMigratedSecretsInMemory checks that c holds every cipher marker, verbatim.
func requireMigratedSecretsInMemory(t *testing.T, c *Config, envContext bool) {
	t.Helper()
	p := nativeCipherPrefix()

	require.Equal(t, p+"MARKER-auth-token", c.ContextStates["ctx1"].AuthToken)
	require.Equal(t, p+"MARKER-auth-refresh-token", c.ContextStates["ctx1"].AuthRefreshToken)
	require.Equal(t, p+"MARKER-saved-password", c.SavedCredentials["ctx1"].EncryptedPassword)
	require.Equal(t, p+"MARKER-cred-api-secret", c.Credentials["cred1"].APIKeyPair.Secret)

	ctx := c.Contexts["ctx1"]
	require.Contains(t, ctx.GlobalAPIKeys, "global1")
	require.Equal(t, p+"MARKER-global-api-secret", ctx.GlobalAPIKeys["global1"].Secret)
	require.Contains(t, ctx.SchemaRegistryClusters, "sr1")
	require.Equal(t, p+"MARKER-sr-credential-secret", ctx.SchemaRegistryClusters["sr1"].SrCredentials.Secret)

	cluster, marker := kafkaMarkerFor(envContext)
	clusters := allKafkaClusterConfigs(ctx.KafkaClusterContext)
	require.Contains(t, clusters, cluster)
	require.Contains(t, clusters[cluster].APIKeys, "key1")
	require.Equal(t, p+marker, clusters[cluster].APIKeys["key1"].Secret)
}

func TestMigrate_StablePopulatesStores(t *testing.T) {
	tests := []struct {
		name       string
		legacy     []byte
		envContext bool
	}{
		{name: "environment kafka context", legacy: cipherLegacyFixture(), envContext: true},
		{name: "direct kafka context", legacy: directKafkaCipherLegacyFixture(), envContext: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			home := t.TempDir()
			setTestHome(t, home)
			setTestChannel(t, pversion.Stable)
			migrations := countLegacyMigrations(t)
			seedLegacyConfig(t, home, tc.legacy)
			c := New()

			err := loadQuietly(t, c)

			require.NoError(t, err)
			require.Equal(t, int32(1), migrations.Load())
			require.FileExists(t, SettingsFilename())
			require.FileExists(t, ContextsFilename())
			require.FileExists(t, SecretsFilename())
			requireMigratedSecretsOnDisk(t, tc.envContext)
			requireMigratedSecretsInMemory(t, c, tc.envContext)
			require.Equal(t, "ctx1", c.CurrentContext)

			cache := newCacheStore()
			var uc updateCheckCache
			require.True(t, cache.readJSON("update_check.json", &uc))
			require.NotNil(t, uc.LastUpdateCheckAt)
			require.Equal(t, "2020-01-02T03:04:05Z", uc.LastUpdateCheckAt.UTC().Format(time.RFC3339))
			flags := map[string]*FeatureFlags{}
			require.True(t, cache.readJSON("feature_flags.json", &flags))
			require.Equal(t, "MARKER-feature-flags-value", flags["ctx1"].CliValues["marker"])

			backup, err := os.ReadFile(migrationBackupPath(home))
			require.NoError(t, err)
			require.Equal(t, tc.legacy, backup)
			require.NoFileExists(t, migrationMarkerPath(home))
		})
	}
}

func TestMigrate_ReloadKeepsMigratedSecrets(t *testing.T) {
	home, migrations := newStableMigrationTest(t)
	seedLegacyConfig(t, home, cipherLegacyFixture())
	require.NoError(t, loadQuietly(t, New()))
	reloaded := New()

	err := loadQuietly(t, reloaded)

	require.NoError(t, err)
	require.Equal(t, int32(1), migrations.Load())
	requireMigratedSecretsInMemory(t, reloaded, true)
}

func TestMigrate_LeavesLegacyFileFrozen(t *testing.T) {
	home := t.TempDir()
	setTestHome(t, home)
	setTestChannel(t, pversion.Stable)
	// secret-free: the merged Save below decrypts secrets, and the marker fixture's cipher-shaped
	// placeholders can't be decrypted.
	legacy := []byte(`{
		"current_context": "ctx1",
		"platforms": {"platform1": {"name": "platform1", "server": "https://example.com"}},
		"credentials": {"cred1": {"name": "cred1", "username": "test-user", "credential_type": 0}},
		"contexts": {
			"ctx1": {
				"name": "ctx1",
				"platform": "platform1",
				"credential": "cred1",
				"kafka_cluster_context": {"environment_context": false}
			}
		},
		"context_states": {"ctx1": {}}
	}`)
	legacyPath := seedLegacyConfig(t, home, legacy)
	c := New()
	require.NoError(t, loadQuietly(t, c))
	require.Equal(t, "ctx1", c.CurrentContext)

	c.CurrentContext = ""
	err := c.Save()

	require.NoError(t, err)
	reloaded := New()
	require.NoError(t, loadQuietly(t, reloaded))
	require.Equal(t, "", reloaded.CurrentContext)
	got, err := os.ReadFile(legacyPath)
	require.NoError(t, err)
	require.Equal(t, legacy, got)
}

func TestMigrate_IsOneShot(t *testing.T) {
	home, migrations := newStableMigrationTest(t)
	seedLegacyConfig(t, home, cipherLegacyFixture())
	require.NoError(t, loadQuietly(t, New()))
	require.Equal(t, int32(1), migrations.Load())
	for _, path := range []string{SettingsFilename(), ContextsFilename(), SecretsFilename()} {
		require.NoError(t, os.Remove(path))
	}

	c := New()
	err := loadQuietly(t, c)

	require.NoError(t, err)
	require.Equal(t, int32(1), migrations.Load())
	require.Empty(t, c.Contexts)
	require.Equal(t, "", c.CurrentContext)
	require.FileExists(t, SettingsFilename())
	require.FileExists(t, ContextsFilename())
}

func TestMigrate_ResumesAfterInterruptedRun(t *testing.T) {
	home, migrations := newStableMigrationTest(t)
	seedLegacyConfig(t, home, cipherLegacyFixture())
	// a marker and stores that disagree with the legacy file, so a whole re-run is
	// distinguishable from a no-op or a merge; no backup, as a run interrupted before it leaves.
	writeTestStore(t, migrationMarkerPath(home), "")
	writeTestStore(t, SecretsFilename(), fmt.Sprintf(`{"tokens": {"ctx1": {"auth_token": "%sSTALE-auth-token"}}}`, nativeCipherPrefix()))
	writeTestStore(t, SettingsFilename(), `{
		"disable_update_check": true,
		"platforms": {"stale-platform": {"name": "stale-platform", "server": "https://stale"}}
	}`)
	writeTestStore(t, ContextsFilename(), `{
		"current_context": "stale",
		"contexts": {
			"stale": {
				"name": "stale",
				"platform": "stale-platform",
				"credential": "stale-cred",
				"kafka_cluster_context": {"environment_context": false}
			}
		}
	}`)

	c := New()
	err := loadQuietly(t, c)

	require.NoError(t, err)
	require.Equal(t, int32(1), migrations.Load())
	require.Equal(t, "ctx1", c.CurrentContext)
	require.NotContains(t, c.Contexts, "stale")
	require.NotContains(t, c.Platforms, "stale-platform")
	require.False(t, c.DisableUpdateCheck)
	requireMigratedSecretsOnDisk(t, true)
	requireMigratedSecretsInMemory(t, c, true)
	require.FileExists(t, migrationBackupPath(home))
	require.NoFileExists(t, migrationMarkerPath(home))
}

func TestMigrate_ZeroByteSecretsStoreIsNotMigrated(t *testing.T) {
	home, migrations := newStableMigrationTest(t)
	require.NoError(t, loadQuietly(t, New()))
	require.NoError(t, os.Truncate(SecretsFilename(), 0))
	seedLegacyConfig(t, home, cipherLegacyFixture())

	err := loadQuietly(t, New())

	require.ErrorContains(t, err, SecretsFilename())
	require.Equal(t, int32(0), migrations.Load())
	require.NoFileExists(t, migrationBackupPath(home))
}

func TestMigrate_PartialStoreResetDoesNotReimport(t *testing.T) {
	// deleting contexts.json is the documented way to reset contexts (see warnOneSidedStores); no
	// deliberate v5 reset, whole or partial, may re-import the v4 file.
	tests := []struct {
		name    string
		deleted []func() string
	}{
		{name: "contexts.json", deleted: []func() string{ContextsFilename}},
		{name: "secrets.json", deleted: []func() string{SecretsFilename}},
		{name: "settings.json", deleted: []func() string{SettingsFilename}},
		{name: "settings.json and contexts.json", deleted: []func() string{SettingsFilename, ContextsFilename}},
		{name: "settings.json and secrets.json", deleted: []func() string{SettingsFilename, SecretsFilename}},
		{name: "contexts.json and secrets.json", deleted: []func() string{ContextsFilename, SecretsFilename}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			home, migrations := newStableMigrationTest(t)
			require.NoError(t, loadQuietly(t, New()))
			for _, store := range tc.deleted {
				require.NoError(t, os.Remove(store()))
			}
			seedLegacyConfig(t, home, cipherLegacyFixture())

			c := New()
			err := loadQuietly(t, c)

			require.NoError(t, err)
			require.Equal(t, int32(0), migrations.Load())
			require.Empty(t, c.Contexts)
			require.Equal(t, "", c.CurrentContext)
			require.NoFileExists(t, migrationBackupPath(home))
		})
	}
}

func TestMigrate_FailedStoreWriteLeavesNoBackup(t *testing.T) {
	home, migrations := newStableMigrationTest(t)
	seedLegacyConfig(t, home, cipherLegacyFixture())
	// a directory appearing in secrets.json's place after the guard makes the last store write
	// fail; the stores written before it now count as present, so only the marker can resume.
	onLegacyConfigStat(t, func(string) { require.NoError(t, os.MkdirAll(SecretsFilename(), 0700)) })

	err := loadQuietly(t, New())

	require.ErrorContains(t, err, "unable to migrate configuration file")
	require.Equal(t, int32(0), migrations.Load())
	require.FileExists(t, SettingsFilename())
	require.FileExists(t, ContextsFilename())
	require.NoFileExists(t, migrationBackupPath(home))
	require.FileExists(t, migrationMarkerPath(home))

	onLegacyConfigStat(t, func(string) {})
	require.NoError(t, os.Remove(SecretsFilename()))

	err = loadQuietly(t, New())

	require.NoError(t, err)
	require.Equal(t, int32(1), migrations.Load())
	requireMigratedSecretsOnDisk(t, true)
	require.FileExists(t, migrationBackupPath(home))
	require.NoFileExists(t, migrationMarkerPath(home))
}

func TestMigrate_FailedBackupWriteResumes(t *testing.T) {
	home, migrations := newStableMigrationTest(t)
	seedLegacyConfig(t, home, cipherLegacyFixture())
	// a directory appearing in the backup's place after the guard makes the backup write fail.
	onLegacyConfigStat(t, func(string) { require.NoError(t, os.MkdirAll(migrationBackupPath(home), 0700)) })

	err := loadQuietly(t, New())

	require.ErrorContains(t, err, migrationBackupPath(home))
	require.Equal(t, int32(0), migrations.Load())
	require.FileExists(t, migrationMarkerPath(home))

	onLegacyConfigStat(t, func(string) {})
	require.NoError(t, os.Remove(migrationBackupPath(home)))

	c := New()
	err = loadQuietly(t, c)

	require.NoError(t, err)
	require.Equal(t, int32(1), migrations.Load())
	require.Equal(t, "ctx1", c.CurrentContext)
	requireMigratedSecretsOnDisk(t, true)
	require.FileExists(t, migrationBackupPath(home))
	require.NoFileExists(t, migrationMarkerPath(home))
}

func TestMigrate_FailedMarkerWriteWritesNothing(t *testing.T) {
	home, migrations := newStableMigrationTest(t)
	seedLegacyConfig(t, home, cipherLegacyFixture())
	require.NoError(t, os.MkdirAll(migrationMarkerPath(home), 0700))

	err := loadQuietly(t, New())

	require.ErrorContains(t, err, migrationMarkerPath(home))
	require.Equal(t, int32(0), migrations.Load())
	require.NoFileExists(t, SettingsFilename())
	require.NoFileExists(t, ContextsFilename())
	require.NoFileExists(t, SecretsFilename())
	require.NoFileExists(t, migrationBackupPath(home))
}

func TestMigrate_StatErrorIsHardError(t *testing.T) {
	tests := []struct {
		name string
		path func(home string) string
	}{
		{name: "backup", path: migrationBackupPath},
		{name: "marker", path: migrationMarkerPath},
		{name: "settings.json", path: func(string) string { return SettingsFilename() }},
		{name: "contexts.json", path: func(string) string { return ContextsFilename() }},
		{name: "secrets.json", path: func(string) string { return SecretsFilename() }},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			home, migrations := newStableMigrationTest(t)
			seedLegacyConfig(t, home, cipherLegacyFixture())
			// a self-referencing symlink fails stat with ELOOP, which is not "absent".
			looped := tc.path(home)
			if err := os.Symlink(looped, looped); err != nil {
				t.Skip("creating a symlink is not supported on this platform:", err)
			}

			err := loadQuietly(t, New())

			require.ErrorContains(t, err, fmt.Sprintf(`"%s"`, looped))
			require.ErrorContains(t, err, "for a pending configuration migration")
			require.Equal(t, int32(0), migrations.Load())
			for _, store := range []string{SettingsFilename(), ContextsFilename(), SecretsFilename()} {
				if store != looped {
					require.NoFileExists(t, store)
				}
			}
		})
	}
}

func TestMigrate_StaleMarkerWithBackupIsRemoved(t *testing.T) {
	home, migrations := newStableMigrationTest(t)
	seedLegacyConfig(t, home, cipherLegacyFixture())
	require.NoError(t, loadQuietly(t, New()))
	// a crash between the backup write and the marker's removal leaves both behind.
	writeTestStore(t, migrationMarkerPath(home), "")

	err := loadQuietly(t, New())

	require.NoError(t, err)
	require.Equal(t, int32(1), migrations.Load())
	require.NoFileExists(t, migrationMarkerPath(home))
}

func TestMigrate_EmptyReadAfterNonEmptyStatIsRetryable(t *testing.T) {
	home, migrations := newStableMigrationTest(t)
	legacyPath := seedLegacyConfig(t, home, cipherLegacyFixture())
	// v4 truncates config.json before rewriting it, so a read can land between the two.
	onLegacyConfigStat(t, func(path string) { require.NoError(t, os.Truncate(path, 0)) })

	err := loadQuietly(t, New())

	requireLegacyFileChangedError(t, err, legacyPath)
	require.Equal(t, int32(0), migrations.Load())
	require.NoFileExists(t, SettingsFilename())
	require.NoFileExists(t, migrationBackupPath(home))
}

// requireLegacyFileChangedError checks err is the retryable "changed while it was being read"
// error for legacyPath, with the retry advice as its suggestion rather than in the message.
func requireLegacyFileChangedError(t *testing.T, err error, legacyPath string) {
	t.Helper()
	require.EqualError(t, err, fmt.Sprintf(`"%s" changed while it was being read`, legacyPath))
	var withSuggestions perrors.ErrorWithSuggestions
	require.True(t, errors.As(err, &withSuggestions), "expected an error with suggestions")
	require.Equal(t, fmt.Sprintf(`Run the command again. If this keeps happening, restore or remove "%s".`, legacyPath), withSuggestions.GetSuggestionsMsg())
}

func TestMigrate_MarkerWithZeroByteLegacyFileIsRetryable(t *testing.T) {
	home, migrations := newStableMigrationTest(t)
	// v4 mid-rewrite during a resume: the marker must survive for the retry.
	legacyPath := seedLegacyConfig(t, home, nil)
	writeTestStore(t, migrationMarkerPath(home), "")

	err := loadQuietly(t, New())

	requireLegacyFileChangedError(t, err, legacyPath)
	require.Equal(t, int32(0), migrations.Load())
	require.FileExists(t, migrationMarkerPath(home))
}

func TestMigrate_StaleMarkerWithoutLegacyFileIsRemoved(t *testing.T) {
	home, migrations := newStableMigrationTest(t)
	legacyPath := seedLegacyConfig(t, home, cipherLegacyFixture())
	onLegacyConfigStat(t, func(string) { require.NoError(t, os.MkdirAll(SecretsFilename(), 0700)) })
	require.Error(t, loadQuietly(t, New()))
	onLegacyConfigStat(t, func(string) {})
	require.NoError(t, os.Remove(SecretsFilename()))
	// the user moves config.json aside and keeps using v5, which removes the orphaned marker.
	aside := legacyPath + ".aside"
	require.NoError(t, os.Rename(legacyPath, aside))
	c := New()
	require.NoError(t, loadQuietly(t, c))
	require.NoFileExists(t, migrationMarkerPath(home))
	c.DisableUpdateCheck = true
	require.NoError(t, c.Save())
	// a v4 run recreates config.json.
	require.NoError(t, os.Rename(aside, legacyPath))

	reloaded := New()
	err := loadQuietly(t, reloaded)

	require.NoError(t, err)
	require.Equal(t, int32(0), migrations.Load())
	require.True(t, reloaded.DisableUpdateCheck)
}

func TestMigrate_NoLegacyFileSkipsMigrationChecks(t *testing.T) {
	tests := []struct {
		name string
		path func(home string) string
	}{
		{name: "backup", path: migrationBackupPath},
		{name: "marker", path: migrationMarkerPath},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			home, migrations := newStableMigrationTest(t)
			// a v5-only machine: no legacy file, so no migration file is worth a stat error.
			looped := tc.path(home)
			require.NoError(t, os.MkdirAll(filepath.Dir(looped), 0700))
			if err := os.Symlink(looped, looped); err != nil {
				t.Skip("creating a symlink is not supported on this platform:", err)
			}

			err := loadQuietly(t, New())

			require.NoError(t, err)
			require.Equal(t, int32(0), migrations.Load())
			require.FileExists(t, SettingsFilename())
		})
	}
}

func TestMigrate_FreshInstallIsNotMigrated(t *testing.T) {
	home, migrations := newStableMigrationTest(t)
	require.NoError(t, loadQuietly(t, New()))
	require.FileExists(t, SettingsFilename())
	require.FileExists(t, ContextsFilename())
	require.FileExists(t, SecretsFilename())
	seedLegacyConfig(t, home, cipherLegacyFixture())

	c := New()
	err := loadQuietly(t, c)

	require.NoError(t, err)
	require.Equal(t, int32(0), migrations.Load())
	require.Empty(t, c.Contexts)
	require.NoFileExists(t, migrationBackupPath(home))
}

func TestMigrate_EmptyLegacyFileIsIgnored(t *testing.T) {
	home, migrations := newStableMigrationTest(t)
	seedLegacyConfig(t, home, nil)

	c := New()
	err := loadQuietly(t, c)

	require.NoError(t, err)
	require.Equal(t, int32(0), migrations.Load())
	require.Empty(t, c.Contexts)
	require.FileExists(t, SettingsFilename())
	require.FileExists(t, ContextsFilename())
	require.NoFileExists(t, migrationBackupPath(home))
}

func TestMigrate_MalformedLegacyFileIsHardError(t *testing.T) {
	home, migrations := newStableMigrationTest(t)
	legacyPath := seedLegacyConfig(t, home, []byte(`{"contexts":`))

	err := loadQuietly(t, New())

	require.ErrorContains(t, err, readConfigurationFileErrorPrefix)
	require.ErrorContains(t, err, legacyPath)
	require.Equal(t, int32(0), migrations.Load())
	require.NoFileExists(t, SettingsFilename())
	require.NoFileExists(t, ContextsFilename())
	require.NoFileExists(t, SecretsFilename())
	require.NoFileExists(t, migrationBackupPath(home))
}

func TestMigrate_NonStableSeedsReadOnly(t *testing.T) {
	tests := []struct {
		channel pversion.Channel
		dir     string
	}{
		{channel: pversion.Dev, dir: ".confluent-dev"},
		{channel: pversion.Prerelease, dir: ".confluent-prerelease"},
	}

	for _, tc := range tests {
		t.Run(tc.channel.String(), func(t *testing.T) {
			home := t.TempDir()
			setTestHome(t, home)
			setTestChannel(t, tc.channel)
			migrations := countLegacyMigrations(t)
			legacy := cipherLegacyFixture()
			legacyPath := seedLegacyConfig(t, home, legacy)
			frozen := time.Date(2020, 1, 2, 3, 4, 5, 0, time.UTC)
			require.NoError(t, os.Chtimes(legacyPath, frozen, frozen))
			stableDir := filepath.Join(home, ".confluent")

			err := loadQuietly(t, New())

			require.NoError(t, err)
			require.Equal(t, int32(1), migrations.Load())
			stateDir := filepath.Join(home, tc.dir)
			require.FileExists(t, filepath.Join(stateDir, "settings.json"))
			require.FileExists(t, filepath.Join(stateDir, "contexts.json"))
			require.FileExists(t, filepath.Join(stateDir, "secrets.json"))
			requireMigratedSecretsOnDisk(t, true)
			backup, err := os.ReadFile(filepath.Join(stateDir, "config.json.v4-backup"))
			require.NoError(t, err)
			require.Equal(t, legacy, backup)

			entries, err := os.ReadDir(stableDir)
			require.NoError(t, err)
			require.Len(t, entries, 1)
			require.Equal(t, "config.json", entries[0].Name())
			got, err := os.ReadFile(legacyPath)
			require.NoError(t, err)
			require.Equal(t, legacy, got)
			info, err := os.Stat(legacyPath)
			require.NoError(t, err)
			require.True(t, info.ModTime().Equal(frozen), "legacy mtime changed to %v", info.ModTime())
		})
	}
}

func TestMigrate_NonStableSkipsSeedAfterStableMigrated(t *testing.T) {
	home := t.TempDir()
	setTestHome(t, home)
	setTestChannel(t, pversion.Dev)
	migrations := countLegacyMigrations(t)
	legacy := cipherLegacyFixture()
	legacyPath := seedLegacyConfig(t, home, legacy)
	writeTestStore(t, stableBackupPath(home), string(legacy))
	frozen := time.Date(2020, 1, 2, 3, 4, 5, 0, time.UTC)
	for _, path := range []string{legacyPath, stableBackupPath(home)} {
		require.NoError(t, os.Chtimes(path, frozen, frozen))
	}

	c := New()
	err := loadQuietly(t, c)

	require.NoError(t, err)
	require.Equal(t, int32(0), migrations.Load())
	require.Empty(t, c.Contexts)
	require.NoFileExists(t, migrationBackupPath(home))
	entries, err := os.ReadDir(filepath.Join(home, ".confluent"))
	require.NoError(t, err)
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		names = append(names, entry.Name())
	}
	require.ElementsMatch(t, []string{"config.json", "config.json.v4-backup"}, names)
	for _, path := range []string{legacyPath, stableBackupPath(home)} {
		info, err := os.Stat(path)
		require.NoError(t, err)
		require.True(t, info.ModTime().Equal(frozen), "%s mtime changed to %v", path, info.ModTime())
	}
}

func TestMigrate_NonStableSeedFailureStartsFresh(t *testing.T) {
	tests := []struct {
		name   string
		legacy string
	}{
		{name: "malformed legacy file", legacy: `{"contexts":`},
		{name: "context without a platform", legacy: legacyContextWithoutPlatform},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			home := t.TempDir()
			setTestHome(t, home)
			setTestChannel(t, pversion.Dev)
			migrations := countLegacyMigrations(t)
			legacyPath := seedLegacyConfig(t, home, []byte(tc.legacy))
			c := New()

			var err error
			stderr := captureStderr(t, func() { err = c.Load() })

			require.NoError(t, err)
			require.Contains(t, stderr, fmt.Sprintf(`Skipped copying contexts and logins from "%s": `, legacyPath))
			require.Equal(t, 1, strings.Count(stderr, "Skipped copying"))
			require.Equal(t, 1, strings.Count(stderr, legacyPath), "the warning names the legacy file once")
			require.Equal(t, int32(0), migrations.Load())
			require.Empty(t, c.Contexts)
			require.Equal(t, "", c.CurrentContext)
			require.FileExists(t, SettingsFilename())
			require.FileExists(t, ContextsFilename())
			require.NoFileExists(t, migrationBackupPath(home))
			require.NoFileExists(t, migrationMarkerPath(home))
		})
	}
}

func TestMigrate_NonStableStaleMarkerAfterStableMigratedIsRemoved(t *testing.T) {
	home := t.TempDir()
	setTestHome(t, home)
	setTestChannel(t, pversion.Dev)
	migrations := countLegacyMigrations(t)
	legacy := cipherLegacyFixture()
	seedLegacyConfig(t, home, legacy)
	writeTestStore(t, stableBackupPath(home), string(legacy))
	writeTestStore(t, migrationMarkerPath(home), "")
	require.NoError(t, loadQuietly(t, New()))
	require.NoFileExists(t, migrationMarkerPath(home))
	require.NoError(t, os.Remove(stableBackupPath(home)))

	c := New()
	err := loadQuietly(t, c)

	require.NoError(t, err)
	require.Equal(t, int32(0), migrations.Load())
	require.Empty(t, c.Contexts)
}

func TestMigrate_NonStableMarkerWithMalformedLegacyFileWarnsOnce(t *testing.T) {
	home := t.TempDir()
	setTestHome(t, home)
	setTestChannel(t, pversion.Dev)
	seedLegacyConfig(t, home, []byte(`{"contexts":`))
	writeTestStore(t, migrationMarkerPath(home), "")
	stderr := captureStderr(t, func() { require.NoError(t, New().Load()) })
	require.Equal(t, 1, strings.Count(stderr, "Skipped copying"))
	require.NoFileExists(t, migrationMarkerPath(home))

	stderr = captureStderr(t, func() { require.NoError(t, New().Load()) })

	require.NotContains(t, stderr, "Skipped copying")
}

// makeUnreadable chmods path to 000 for the rest of the test. Windows ignores the read bit and
// root ignores the mode, so either skips.
func makeUnreadable(t *testing.T, path string) {
	t.Helper()
	if runtime.GOOS == "windows" || os.Geteuid() == 0 {
		t.Skip("file modes don't block reads on this platform or for this user")
	}
	require.NoError(t, os.Chmod(path, 0))
	t.Cleanup(func() { _ = os.Chmod(path, 0600) })
}

func TestMigrate_NonStableUnreadableStableFileWarnsAndStartsFresh(t *testing.T) {
	tests := []struct {
		name  string
		setup func(t *testing.T, home, legacyPath string)
	}{
		{
			name:  "unreadable legacy file",
			setup: func(t *testing.T, _, legacyPath string) { makeUnreadable(t, legacyPath) },
		},
		{
			name: "unreadable Stable backup",
			setup: func(t *testing.T, home, _ string) {
				// a self-referencing symlink fails stat with ELOOP.
				if err := os.Symlink(stableBackupPath(home), stableBackupPath(home)); err != nil {
					t.Skip("creating a symlink is not supported on this platform:", err)
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			home := t.TempDir()
			setTestHome(t, home)
			setTestChannel(t, pversion.Dev)
			migrations := countLegacyMigrations(t)
			legacyPath := seedLegacyConfig(t, home, cipherLegacyFixture())
			tc.setup(t, home, legacyPath)
			c := New()

			var err error
			stderr := captureStderr(t, func() { err = c.Load() })

			require.NoError(t, err)
			require.Equal(t, 1, strings.Count(stderr, fmt.Sprintf(`Skipped copying contexts and logins from "%s": `, legacyPath)))
			require.Equal(t, int32(0), migrations.Load())
			require.Empty(t, c.Contexts)
			require.FileExists(t, SettingsFilename())
			require.FileExists(t, ContextsFilename())
		})
	}
}

func TestMigrate_StableUnreadableLegacyFileIsHardError(t *testing.T) {
	home, migrations := newStableMigrationTest(t)
	legacyPath := seedLegacyConfig(t, home, cipherLegacyFixture())
	makeUnreadable(t, legacyPath)

	err := loadQuietly(t, New())

	require.ErrorContains(t, err, fmt.Sprintf(`unable to read configuration file "%s"`, legacyPath))
	require.Equal(t, int32(0), migrations.Load())
	require.NoFileExists(t, SettingsFilename())
}

func TestMigrate_TornReadIsRetryable(t *testing.T) {
	tests := []struct {
		name        string
		channel     pversion.Channel
		interrupted bool
	}{
		{name: "stable", channel: pversion.Stable},
		{name: "stable resuming", channel: pversion.Stable, interrupted: true},
		{name: "dev", channel: pversion.Dev},
		{name: "dev resuming", channel: pversion.Dev, interrupted: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			home := t.TempDir()
			setTestHome(t, home)
			setTestChannel(t, tc.channel)
			migrations := countLegacyMigrations(t)
			full := cipherLegacyFixture()
			legacyPath := seedLegacyConfig(t, home, full)
			if tc.interrupted {
				writeTestStore(t, migrationMarkerPath(home), "")
			}
			// v4 truncates then writes, so a read can see a prefix of the new file.
			onLegacyConfigStat(t, func(path string) { require.NoError(t, os.WriteFile(path, full[:len(full)/2], 0600)) })

			var err error
			stderr := captureStderr(t, func() { err = New().Load() })

			requireLegacyFileChangedError(t, err, legacyPath)
			require.NotContains(t, stderr, "Skipped copying")
			require.Equal(t, int32(0), migrations.Load())
			require.NoFileExists(t, SettingsFilename())
			require.NoFileExists(t, ContextsFilename())
			require.NoFileExists(t, SecretsFilename())
			if tc.interrupted {
				require.FileExists(t, migrationMarkerPath(home))
			} else {
				require.NoFileExists(t, migrationMarkerPath(home))
			}

			// once v4 finishes its write, the next load migrates.
			onLegacyConfigStat(t, func(string) {})
			require.NoError(t, os.WriteFile(legacyPath, full, 0600))
			require.NoError(t, loadQuietly(t, New()))
			require.Equal(t, int32(1), migrations.Load())
		})
	}
}

func TestMigrate_UnresolvableHomeSkipsMigration(t *testing.T) {
	cwd := t.TempDir()
	t.Chdir(cwd)
	setTestHome(t, "")
	setTestChannel(t, pversion.Stable)
	migrations := countLegacyMigrations(t)
	// with no home, a naive join resolves the legacy file against the working directory.
	seedLegacyConfig(t, "", cipherLegacyFixture())
	c := New()

	err := loadQuietly(t, c)

	require.NoError(t, err)
	require.Equal(t, int32(0), migrations.Load())
	require.Empty(t, c.Contexts)
	require.NoFileExists(t, filepath.Join(cwd, ".confluent", "config.json.v4-backup"))
}

// sharedLoginLegacy is a v4 config.json whose two contexts share one credential (so one identity
// in secrets.json) yet hold different keys for the same Kafka cluster and Schema Registry.
const sharedLoginLegacy = `{
	"current_context": "ctxA",
	"platforms": {"p1": {"name": "p1", "server": "https://confluent.cloud"}},
	"credentials": {"cred1": {"name": "cred1", "username": "u@example.com", "credential_type": 0}},
	"contexts": {
		"ctxA": {
			"name": "ctxA", "platform": "p1", "credential": "cred1",
			"kafka_cluster_context": {
				"environment_context": false,
				"active_kafka": "lkc-1",
				"kafka_cluster_configs": {
					"lkc-1": {"id": "lkc-1", "name": "c", "bootstrap_servers": "b:9092", "api_key": "kA",
						"api_keys": {"kA": {"api_key": "kA", "api_secret": "secretA"}}}
				}
			},
			"schema_registry_clusters": {
				"sr1": {"id": "sr1", "schema_registry_endpoint": "https://sr", "schema_registry_credentials": {"api_key": "srA", "api_secret": "srSecretA"}}
			}
		},
		"ctxB": {
			"name": "ctxB", "platform": "p1", "credential": "cred1",
			"kafka_cluster_context": {
				"environment_context": false,
				"active_kafka": "lkc-1",
				"kafka_cluster_configs": {
					"lkc-1": {"id": "lkc-1", "name": "c", "bootstrap_servers": "b:9092", "api_key": "kB",
						"api_keys": {"kB": {"api_key": "kB", "api_secret": "secretB"}}}
				}
			},
			"schema_registry_clusters": {
				"sr1": {"id": "sr1", "schema_registry_endpoint": "https://sr", "schema_registry_credentials": {"api_key": "srB", "api_secret": "srSecretB"}}
			}
		}
	},
	"context_states": {"ctxA": {}, "ctxB": {}}
}`

// requireDecryptsTo checks that pair's stored secret decrypts to want, leaving pair untouched.
func requireDecryptsTo(t *testing.T, pair *APIKeyPair, want, label string) {
	t.Helper()
	require.NotNil(t, pair, label)
	shadow := &APIKeyPair{Key: pair.Key, Secret: pair.Secret, Salt: pair.Salt, Nonce: pair.Nonce}
	require.NoError(t, shadow.DecryptSecret(), label)
	require.Equal(t, want, shadow.Secret, label)
}

// requireSharedLoginSecrets checks that each sharedLoginLegacy context kept its own keys.
func requireSharedLoginSecrets(t *testing.T, c *Config) {
	t.Helper()
	kafkaKey := func(ctx, key string) *APIKeyPair {
		return c.Contexts[ctx].KafkaClusterContext.KafkaClusterConfigs["lkc-1"].APIKeys[key]
	}
	requireDecryptsTo(t, kafkaKey("ctxA", "kA"), "secretA", "ctxA kafka key")
	requireDecryptsTo(t, kafkaKey("ctxB", "kB"), "secretB", "ctxB kafka key")
	requireDecryptsTo(t, c.Contexts["ctxA"].SchemaRegistryClusters["sr1"].SrCredentials, "srSecretA", "ctxA sr key")
	requireDecryptsTo(t, c.Contexts["ctxB"].SchemaRegistryClusters["sr1"].SrCredentials, "srSecretB", "ctxB sr key")
}

func TestMigrate_SharedLoginKeepsEveryKey(t *testing.T) {
	home, migrations := newStableMigrationTest(t)
	seedLegacyConfig(t, home, []byte(sharedLoginLegacy))
	migrated := New()

	require.NoError(t, loadQuietly(t, migrated))

	require.Equal(t, int32(1), migrations.Load())
	requireSharedLoginSecrets(t, migrated)

	// two merged save and reload cycles, each from a freshly loaded config.
	migrated.DisableUpdateCheck = true
	require.NoError(t, migrated.Save())
	reloaded := New()
	require.NoError(t, loadQuietly(t, reloaded))
	requireSharedLoginSecrets(t, reloaded)

	reloaded.CurrentContext = "ctxB"
	require.NoError(t, reloaded.Save())
	again := New()
	require.NoError(t, loadQuietly(t, again))
	requireSharedLoginSecrets(t, again)
}

func TestMigrate_ConcurrentFirstRunsMigrateOnce(t *testing.T) {
	const loaders = 4
	home, migrations := newStableMigrationTest(t)
	seedLegacyConfig(t, home, cipherLegacyFixture())
	configs := make([]*Config, loaders)
	for i := range configs {
		configs[i] = New()
	}
	errs := make([]error, loaders)

	captureStderr(t, func() {
		var wg sync.WaitGroup
		for i, c := range configs {
			wg.Add(1)
			go func(i int, c *Config) {
				defer wg.Done()
				errs[i] = c.Load()
			}(i, c)
		}
		wg.Wait()
	})

	require.Equal(t, int32(1), migrations.Load())
	for i, c := range configs {
		require.NoError(t, errs[i], "loader %d", i)
		require.Equal(t, "ctx1", c.CurrentContext, "loader %d", i)
	}
}

// plaintextV4Values are the secrets stateful_cloud.json holds in plaintext, as v4 could leave them.
var plaintextV4Values = []string{"def-secret-456", "eyJ.eyJ.abc", "v1.abc"}

// requireNoPlaintextInStores checks that no plaintext v4 secret reached any store file.
func requireNoPlaintextInStores(t *testing.T) {
	t.Helper()
	for _, path := range []string{SettingsFilename(), ContextsFilename(), SecretsFilename()} {
		data, err := os.ReadFile(path)
		require.NoError(t, err)
		for _, value := range plaintextV4Values {
			require.NotContains(t, string(data), value, "plaintext %q in %s", value, path)
		}
	}
}

// seedStatefulCloudLegacy copies the plaintext stateful_cloud.json fixture in as home's v4 file.
func seedStatefulCloudLegacy(t *testing.T, home string) {
	t.Helper()
	legacy, err := os.ReadFile(filepath.Join("test_json", "stateful_cloud.json"))
	require.NoError(t, err)
	seedLegacyConfig(t, home, legacy)
}

// statefulCloudKafkaPair is the fixture context's nested Kafka API key.
func statefulCloudKafkaPair(c *Config) *APIKeyPair {
	kcc := c.Contexts["my-context"].KafkaClusterContext
	return kcc.KafkaEnvContexts["env-123456"].KafkaClusterConfigs["anonymous-id"].APIKeys["abc-key-123"]
}

// loadAndDecrypt loads the stores and decrypts the context's tokens and nested Kafka key in place,
// as PreRun and ResolveKafkaAPIKey would.
func loadAndDecrypt(t *testing.T) *Config {
	t.Helper()
	c := New()
	require.NoError(t, loadQuietly(t, c))
	require.NoError(t, c.DecryptContextStates())
	require.NoError(t, c.DecryptCredentials())
	require.NoError(t, statefulCloudKafkaPair(c).DecryptSecret())
	return c
}

// requireStatefulCloudPlaintext checks that c holds the fixture's original plaintext secrets.
func requireStatefulCloudPlaintext(t *testing.T, c *Config) {
	t.Helper()
	require.Equal(t, "eyJ.eyJ.abc", c.ContextStates["my-context"].AuthToken)
	require.Equal(t, "v1.abc", c.ContextStates["my-context"].AuthRefreshToken)
	require.Equal(t, "def-secret-456", statefulCloudKafkaPair(c).Secret)
}

func TestMigrate_PlaintextV4FileIsEncrypted(t *testing.T) {
	home := t.TempDir()
	setTestHome(t, home)
	setTestChannel(t, pversion.Stable)
	seedStatefulCloudLegacy(t, home)

	err := loadQuietly(t, New())

	require.NoError(t, err)
	requireNoPlaintextInStores(t)
	file, err := readSecretFileFromDisk(SecretsFilename())
	require.NoError(t, err)
	p := nativeCipherPrefix()
	token := file.Tokens["my-context"]
	require.NotNil(t, token)
	require.True(t, strings.HasPrefix(token.AuthToken, p), "auth token not encrypted: %q", token.AuthToken)
	require.True(t, strings.HasPrefix(token.AuthRefreshToken, p), "refresh token not encrypted: %q", token.AuthRefreshToken)
	identity := file.Secrets["username-test-user"]
	require.NotNil(t, identity)
	require.Contains(t, identity.KafkaAPIKeys["anonymous-id"], "abc-key-123")
	kafka := identity.KafkaAPIKeys["anonymous-id"]["abc-key-123"]
	require.True(t, strings.HasPrefix(kafka.Secret, p), "kafka secret not encrypted: %q", kafka.Secret)
	// no context references this credential, so it is saved but never loaded back: decrypt the
	// stored record directly instead.
	cred := file.Secrets["api-key-abc-key-123"]
	require.NotNil(t, cred)
	require.True(t, strings.HasPrefix(cred.Secret, p), "credential secret not encrypted: %q", cred.Secret)
	credPair := &APIKeyPair{Key: "abc-key-123", Secret: cred.Secret, Salt: cred.SecretSalt, Nonce: cred.SecretNonce}
	require.NoError(t, credPair.DecryptSecret())
	require.Equal(t, "def-secret-456", credPair.Secret)
}

func TestMigrate_PlaintextV4FileDecryptsAfterReload(t *testing.T) {
	home := t.TempDir()
	setTestHome(t, home)
	setTestChannel(t, pversion.Stable)
	seedStatefulCloudLegacy(t, home)
	require.NoError(t, loadQuietly(t, New()))

	c := loadAndDecrypt(t)

	requireStatefulCloudPlaintext(t, c)
}

func TestMigrate_PlaintextV4FileSurvivesMergedSave(t *testing.T) {
	home := t.TempDir()
	setTestHome(t, home)
	setTestChannel(t, pversion.Stable)
	seedStatefulCloudLegacy(t, home)
	require.NoError(t, loadQuietly(t, New()))
	c := loadAndDecrypt(t)

	c.DisableUpdateCheck = true
	err := c.Save()

	require.NoError(t, err)
	requireNoPlaintextInStores(t)
	reloaded := loadAndDecrypt(t)
	require.True(t, reloaded.DisableUpdateCheck)
	requireStatefulCloudPlaintext(t, reloaded)
}

// legacyContextWithoutPlatform is a v4 config.json that parses but fails wireContexts: its one
// context names no platform.
const legacyContextWithoutPlatform = `{
	"current_context": "ctx1",
	"credentials": {"cred1": {"name": "cred1", "username": "test-user", "credential_type": 0}},
	"contexts": {
		"ctx1": {
			"name": "ctx1",
			"credential": "cred1",
			"kafka_cluster_context": {"environment_context": false}
		}
	},
	"context_states": {"ctx1": {}}
}`

// TestMigrate_WireContextsFailureIsHardError covers a legacy file whose context has no platform:
// it fails wireContexts, and the caller must still learn which legacy file caused it.
func TestMigrate_WireContextsFailureIsHardError(t *testing.T) {
	home, migrations := newStableMigrationTest(t)
	legacyPath := seedLegacyConfig(t, home, []byte(legacyContextWithoutPlatform))

	err := loadQuietly(t, New())

	require.ErrorContains(t, err, "unable to migrate configuration file")
	require.ErrorContains(t, err, fmt.Sprintf("%q", legacyPath))
	require.Equal(t, int32(0), migrations.Load())
	require.NoFileExists(t, migrationBackupPath(home))
	require.NoFileExists(t, SecretsFilename())

	// migrationErrorMsg's %w must keep the underlying typed error reachable, not just its text.
	var corrupted *perrors.CorruptedConfigError
	require.True(t, errors.As(err, &corrupted), "expected a *errors.CorruptedConfigError in the chain")
}

func TestMigrate_AnnouncesOnce(t *testing.T) {
	home := t.TempDir()
	setTestHome(t, home)
	setTestChannel(t, pversion.Stable)
	seedLegacyConfig(t, home, cipherLegacyFixture())
	c := New()

	stderr := captureStderr(t, func() {
		require.NoError(t, c.Load())
	})

	require.Contains(t, stderr, migrationAnnouncementMsg)

	stderr = captureStderr(t, func() {
		require.NoError(t, New().Load())
	})

	require.NotContains(t, stderr, migrationAnnouncementMsg)
}

func TestMigrate_NoAnnouncementWithoutContexts(t *testing.T) {
	home := t.TempDir()
	setTestHome(t, home)
	setTestChannel(t, pversion.Stable)
	// a legacy file with no contexts: migration and backup still happen, but there is nothing
	// worth telling the user about.
	seedLegacyConfig(t, home, []byte(`{"disable_update_check": true}`))

	stderr := captureStderr(t, func() {
		require.NoError(t, New().Load())
	})

	require.NotContains(t, stderr, migrationAnnouncementMsg)
	require.FileExists(t, migrationBackupPath(home))
}

func TestMigrate_NonStableSeedAnnouncement(t *testing.T) {
	home := t.TempDir()
	setTestHome(t, home)
	setTestChannel(t, pversion.Dev)
	legacyPath := seedLegacyConfig(t, home, cipherLegacyFixture())

	stderr := captureStderr(t, func() {
		require.NoError(t, New().Load())
	})

	stateDir, err := StateDir()
	require.NoError(t, err)
	require.Contains(t, stderr, fmt.Sprintf(migrationSeedMsg, legacyPath, stateDir))
}

func TestMigrate_WarnsOncePerLegacyWrite(t *testing.T) {
	home := t.TempDir()
	setTestHome(t, home)
	setTestChannel(t, pversion.Stable)
	legacyPath := seedLegacyConfig(t, home, cipherLegacyFixture())
	require.NoError(t, loadQuietly(t, New()))

	// silent immediately after migration: the legacy file has not changed since the stamp.
	stderr := captureStderr(t, func() {
		require.NoError(t, New().Load())
	})
	require.NotContains(t, stderr, "changed after your configuration moved")

	rewriteLegacyFile(t, legacyPath)

	stderr = captureStderr(t, func() {
		require.NoError(t, New().Load())
	})
	require.Contains(t, stderr, fmt.Sprintf(legacyFileChangedWarningMsg, legacyPath))

	// silent again: the new stamp now matches.
	stderr = captureStderr(t, func() {
		require.NoError(t, New().Load())
	})
	require.NotContains(t, stderr, "changed after your configuration moved")

	rewriteLegacyFile(t, legacyPath)

	stderr = captureStderr(t, func() {
		require.NoError(t, New().Load())
	})
	require.Contains(t, stderr, fmt.Sprintf(legacyFileChangedWarningMsg, legacyPath))
}

func TestMigrate_NoDowngradeWarningOnNonStable(t *testing.T) {
	home := t.TempDir()
	setTestHome(t, home)
	setTestChannel(t, pversion.Stable)
	legacyPath := seedLegacyConfig(t, home, cipherLegacyFixture())
	require.NoError(t, loadQuietly(t, New()))
	setTestChannel(t, pversion.Dev)
	rewriteLegacyFile(t, legacyPath)

	stderr := captureStderr(t, func() {
		require.NoError(t, New().Load())
	})

	require.NotContains(t, stderr, "changed after your configuration moved")
}

func TestMigrate_MissingStampRecordsSilently(t *testing.T) {
	home := t.TempDir()
	setTestHome(t, home)
	setTestChannel(t, pversion.Stable)
	legacyPath := seedLegacyConfig(t, home, cipherLegacyFixture())
	require.NoError(t, loadQuietly(t, New()))
	require.NoError(t, os.Remove(filepath.Join(CacheDir(), "legacy_config.json")))
	rewriteLegacyFile(t, legacyPath)

	stderr := captureStderr(t, func() {
		require.NoError(t, New().Load())
	})

	require.NotContains(t, stderr, "changed after your configuration moved")
	require.FileExists(t, filepath.Join(CacheDir(), "legacy_config.json"))

	// the stamp was recreated, so a further reload without another write stays silent.
	stderr = captureStderr(t, func() {
		require.NoError(t, New().Load())
	})
	require.NotContains(t, stderr, "changed after your configuration moved")
}

// rewriteLegacyFile overwrites path with new bytes and advances its mtime, simulating a v4
// downgrade writing to the frozen legacy file.
func rewriteLegacyFile(t *testing.T, path string) {
	t.Helper()
	info, err := os.Stat(path)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(path, []byte(`{"current_context":"","disable_update_check":true}`), 0600))
	require.NoError(t, os.Chtimes(path, time.Time{}, info.ModTime().Add(time.Second)))
}
