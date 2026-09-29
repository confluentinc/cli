package config

import (
	"bytes"
	"encoding/base64"
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

	"github.com/confluentinc/cli/v4/pkg/secret"
	pversion "github.com/confluentinc/cli/v4/pkg/version"
)

// legacyConfigFileCase is one readLegacyConfigFile scenario: a path and the (data, found) it
// should return.
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

	cases := []legacyConfigFileCase{
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
	cases = append(cases, legacyConfigFileCase{name: "symlink to file", path: symlinkToFile, wantFound: true, wantData: want})

	danglingSymlink := filepath.Join(dir, "dangling-symlink.json")
	require.NoError(t, os.Symlink(filepath.Join(dir, "does-not-exist.json"), danglingSymlink))
	cases = append(cases, legacyConfigFileCase{name: "dangling symlink", path: danglingSymlink, wantFound: false})

	symlinkToDir := filepath.Join(dir, "symlink-to-dir.json")
	require.NoError(t, os.Symlink(dir, symlinkToDir))
	cases = append(cases, legacyConfigFileCase{name: "symlink to directory", path: symlinkToDir, wantFound: false})

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			data, found, err := readLegacyConfigFile(tc.path)

			require.NoError(t, err)
			require.Equal(t, tc.wantFound, found)
			require.Equal(t, tc.wantData, data)
		})
	}
}

func TestApplyLegacyConfig_Malformed(t *testing.T) {
	cases := []struct {
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

	for _, tc := range cases {
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
	cases := []struct {
		fixture string
		// kafkaViaEnvContext is true when the fixture's Kafka cluster config (and its nested
		// API key) live under kafka_environment_contexts rather than kafka_cluster_configs.
		kafkaViaEnvContext bool
	}{
		{fixture: "stateful_cloud.json", kafkaViaEnvContext: true},
		{fixture: "stateful_onprem.json", kafkaViaEnvContext: false},
	}

	for _, tc := range cases {
		t.Run(tc.fixture, func(t *testing.T) {
			home := t.TempDir()
			setTestHome(t, home)
			c := New()

			src, err := os.ReadFile(filepath.Join("test_json", tc.fixture))
			require.NoError(t, err)

			path := filepath.Join(home, "config.json")
			require.NoError(t, os.WriteFile(path, src, 0600))

			data, found, err := readLegacyConfigFile(path)
			require.NoError(t, err)
			require.True(t, found)

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
// overlay type. It also fails if an allowlist entry is never visited (a stale entry could mask
// real drift elsewhere). It only verifies that a retagged field's TYPE has a matching overlay
// field, not that the overlay's wrapper structs reach every holder along the way (e.g. both
// KafkaClusterConfigs and KafkaEnvContexts) - TestApplyLegacyConfig_CopiesSecretsVerbatim covers
// that with real markers.
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

func requireFileExists(t *testing.T, path string) {
	t.Helper()
	_, err := os.Stat(path)
	require.NoError(t, err, "expected %s to exist", path)
}

func requireFileAbsent(t *testing.T, path string) {
	t.Helper()
	_, err := os.Stat(path)
	require.True(t, os.IsNotExist(err), "expected %s to be absent, got err=%v", path, err)
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
	require.Contains(t, rec.SchemaRegistryCredentials, "sr1")
	require.Equal(t, p+"MARKER-sr-credential-secret", rec.SchemaRegistryCredentials["sr1"].Secret)

	cluster, marker := "cluster1", "MARKER-kafka-cluster-configs-secret"
	if envContext {
		cluster, marker = "cluster2", "MARKER-kafka-env-contexts-secret"
	}
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

	cluster, marker := "cluster1", "MARKER-kafka-cluster-configs-secret"
	if envContext {
		cluster, marker = "cluster2", "MARKER-kafka-env-contexts-secret"
	}
	clusters := allKafkaClusterConfigs(ctx.KafkaClusterContext)
	require.Contains(t, clusters, cluster)
	require.Contains(t, clusters[cluster].APIKeys, "key1")
	require.Equal(t, p+marker, clusters[cluster].APIKeys["key1"].Secret)
}

func TestMigrate_StablePopulatesStores(t *testing.T) {
	cases := []struct {
		name       string
		legacy     []byte
		envContext bool
	}{
		{name: "environment kafka context", legacy: cipherLegacyFixture(), envContext: true},
		{name: "direct kafka context", legacy: directKafkaCipherLegacyFixture(), envContext: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			home := t.TempDir()
			setTestHome(t, home)
			setTestChannel(t, pversion.Stable)
			migrations := countLegacyMigrations(t)
			seedLegacyConfig(t, home, tc.legacy)
			c := New()

			err := c.Load()

			require.NoError(t, err)
			require.Equal(t, int32(1), migrations.Load())
			requireFileExists(t, SettingsFilename())
			requireFileExists(t, ContextsFilename())
			requireFileExists(t, SecretsFilename())
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
		})
	}
}

func TestMigrate_ReloadKeepsMigratedSecrets(t *testing.T) {
	home := t.TempDir()
	setTestHome(t, home)
	setTestChannel(t, pversion.Stable)
	migrations := countLegacyMigrations(t)
	seedLegacyConfig(t, home, cipherLegacyFixture())
	require.NoError(t, New().Load())
	reloaded := New()

	err := reloaded.Load()

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
	require.NoError(t, c.Load())
	require.Equal(t, "ctx1", c.CurrentContext)

	c.CurrentContext = ""
	err := c.Save()

	require.NoError(t, err)
	reloaded := New()
	require.NoError(t, reloaded.Load())
	require.Equal(t, "", reloaded.CurrentContext)
	got, err := os.ReadFile(legacyPath)
	require.NoError(t, err)
	require.Equal(t, legacy, got)
}

func TestMigrate_IsOneShot(t *testing.T) {
	home := t.TempDir()
	setTestHome(t, home)
	setTestChannel(t, pversion.Stable)
	migrations := countLegacyMigrations(t)
	seedLegacyConfig(t, home, cipherLegacyFixture())
	require.NoError(t, New().Load())
	require.Equal(t, int32(1), migrations.Load())
	for _, path := range []string{SettingsFilename(), ContextsFilename(), SecretsFilename()} {
		require.NoError(t, os.Remove(path))
	}

	c := New()
	err := c.Load()

	require.NoError(t, err)
	require.Equal(t, int32(1), migrations.Load())
	require.Empty(t, c.Contexts)
	require.Equal(t, "", c.CurrentContext)
	requireFileExists(t, SettingsFilename())
	requireFileExists(t, ContextsFilename())
}

func TestMigrate_ResumesAfterInterruptedRun(t *testing.T) {
	home := t.TempDir()
	setTestHome(t, home)
	setTestChannel(t, pversion.Stable)
	migrations := countLegacyMigrations(t)
	seedLegacyConfig(t, home, cipherLegacyFixture())
	// stores that disagree with the legacy file, so a whole re-run is distinguishable from a
	// no-op or a merge; no secrets.json and no backup, as an interrupted run leaves them.
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
	err := c.Load()

	require.NoError(t, err)
	require.Equal(t, int32(1), migrations.Load())
	require.Equal(t, "ctx1", c.CurrentContext)
	require.NotContains(t, c.Contexts, "stale")
	require.NotContains(t, c.Platforms, "stale-platform")
	require.False(t, c.DisableUpdateCheck)
	requireMigratedSecretsOnDisk(t, true)
	requireMigratedSecretsInMemory(t, c, true)
	requireFileExists(t, migrationBackupPath(home))
}

func TestMigrate_ZeroByteSecretsStoreIsMigrated(t *testing.T) {
	home := t.TempDir()
	setTestHome(t, home)
	setTestChannel(t, pversion.Stable)
	migrations := countLegacyMigrations(t)
	require.NoError(t, New().Load())
	require.NoError(t, os.Truncate(SecretsFilename(), 0))
	seedLegacyConfig(t, home, cipherLegacyFixture())

	c := New()
	err := c.Load()

	require.NoError(t, err)
	require.Equal(t, int32(1), migrations.Load())
	require.Equal(t, "ctx1", c.CurrentContext)
	requireMigratedSecretsOnDisk(t, true)
	requireFileExists(t, migrationBackupPath(home))
}

func TestMigrate_ContextsResetDoesNotReimport(t *testing.T) {
	home := t.TempDir()
	setTestHome(t, home)
	setTestChannel(t, pversion.Stable)
	migrations := countLegacyMigrations(t)
	require.NoError(t, New().Load())
	// deleting contexts.json is the documented way to reset contexts (see warnOneSidedStores).
	require.NoError(t, os.Remove(ContextsFilename()))
	seedLegacyConfig(t, home, cipherLegacyFixture())

	c := New()
	err := c.Load()

	require.NoError(t, err)
	require.Equal(t, int32(0), migrations.Load())
	require.Empty(t, c.Contexts)
	require.Equal(t, "", c.CurrentContext)
	requireFileAbsent(t, ContextsFilename())
	requireFileAbsent(t, migrationBackupPath(home))
}

func TestMigrate_FailedStoreWriteLeavesNoBackup(t *testing.T) {
	home := t.TempDir()
	setTestHome(t, home)
	setTestChannel(t, pversion.Stable)
	migrations := countLegacyMigrations(t)
	seedLegacyConfig(t, home, cipherLegacyFixture())
	// a directory in secrets.json's place makes the last store write fail.
	require.NoError(t, os.MkdirAll(SecretsFilename(), 0700))

	err := New().Load()

	require.ErrorContains(t, err, "unable to migrate configuration file")
	require.Equal(t, int32(0), migrations.Load())
	requireFileAbsent(t, migrationBackupPath(home))

	require.NoError(t, os.Remove(SecretsFilename()))

	err = New().Load()

	require.NoError(t, err)
	require.Equal(t, int32(1), migrations.Load())
	requireMigratedSecretsOnDisk(t, true)
	requireFileExists(t, migrationBackupPath(home))
}

func TestMigrate_FreshInstallIsNotMigrated(t *testing.T) {
	home := t.TempDir()
	setTestHome(t, home)
	setTestChannel(t, pversion.Stable)
	migrations := countLegacyMigrations(t)
	require.NoError(t, New().Load())
	requireFileExists(t, SettingsFilename())
	requireFileExists(t, ContextsFilename())
	requireFileExists(t, SecretsFilename())
	seedLegacyConfig(t, home, cipherLegacyFixture())

	c := New()
	err := c.Load()

	require.NoError(t, err)
	require.Equal(t, int32(0), migrations.Load())
	require.Empty(t, c.Contexts)
	requireFileAbsent(t, migrationBackupPath(home))
}

func TestMigrate_EmptyLegacyFileIsIgnored(t *testing.T) {
	home := t.TempDir()
	setTestHome(t, home)
	setTestChannel(t, pversion.Stable)
	migrations := countLegacyMigrations(t)
	seedLegacyConfig(t, home, nil)

	c := New()
	err := c.Load()

	require.NoError(t, err)
	require.Equal(t, int32(0), migrations.Load())
	require.Empty(t, c.Contexts)
	requireFileExists(t, SettingsFilename())
	requireFileExists(t, ContextsFilename())
	requireFileAbsent(t, migrationBackupPath(home))
}

func TestMigrate_MalformedLegacyFileIsHardError(t *testing.T) {
	home := t.TempDir()
	setTestHome(t, home)
	setTestChannel(t, pversion.Stable)
	migrations := countLegacyMigrations(t)
	legacyPath := seedLegacyConfig(t, home, []byte(`{"contexts":`))

	err := New().Load()

	require.ErrorContains(t, err, readConfigurationFileErrorPrefix)
	require.ErrorContains(t, err, legacyPath)
	require.Equal(t, int32(0), migrations.Load())
	requireFileAbsent(t, SettingsFilename())
	requireFileAbsent(t, ContextsFilename())
	requireFileAbsent(t, SecretsFilename())
	requireFileAbsent(t, migrationBackupPath(home))
}

func TestMigrate_NonStableSeedsReadOnly(t *testing.T) {
	cases := []struct {
		channel pversion.Channel
		dir     string
	}{
		{channel: pversion.Dev, dir: ".confluent-dev"},
		{channel: pversion.Prerelease, dir: ".confluent-prerelease"},
	}

	for _, tc := range cases {
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

			err := New().Load()

			require.NoError(t, err)
			require.Equal(t, int32(1), migrations.Load())
			stateDir := filepath.Join(home, tc.dir)
			requireFileExists(t, filepath.Join(stateDir, "settings.json"))
			requireFileExists(t, filepath.Join(stateDir, "contexts.json"))
			requireFileExists(t, filepath.Join(stateDir, "secrets.json"))
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

func TestMigrate_ConcurrentFirstRunsMigrateOnce(t *testing.T) {
	const loaders = 4
	home := t.TempDir()
	setTestHome(t, home)
	setTestChannel(t, pversion.Stable)
	migrations := countLegacyMigrations(t)
	seedLegacyConfig(t, home, cipherLegacyFixture())
	configs := make([]*Config, loaders)
	for i := range configs {
		configs[i] = New()
	}
	errs := make([]error, loaders)

	var wg sync.WaitGroup
	for i, c := range configs {
		wg.Add(1)
		go func(i int, c *Config) {
			defer wg.Done()
			errs[i] = c.Load()
		}(i, c)
	}
	wg.Wait()

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
	require.NoError(t, c.Load())
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

	err := New().Load()

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
	require.NoError(t, New().Load())

	c := loadAndDecrypt(t)

	requireStatefulCloudPlaintext(t, c)
}

func TestMigrate_PlaintextV4FileSurvivesMergedSave(t *testing.T) {
	home := t.TempDir()
	setTestHome(t, home)
	setTestChannel(t, pversion.Stable)
	seedStatefulCloudLegacy(t, home)
	require.NoError(t, New().Load())
	c := loadAndDecrypt(t)

	c.DisableUpdateCheck = true
	err := c.Save()

	require.NoError(t, err)
	requireNoPlaintextInStores(t)
	reloaded := loadAndDecrypt(t)
	require.True(t, reloaded.DisableUpdateCheck)
	requireStatefulCloudPlaintext(t, reloaded)
}
