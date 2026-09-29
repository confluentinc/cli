package config

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReadLegacyConfigFile_Missing(t *testing.T) {
	dir := t.TempDir()

	data, found, err := readLegacyConfigFile(filepath.Join(dir, "missing.json"))
	require.NoError(t, err)
	require.False(t, found)
	require.Nil(t, data)

	data, found, err = readLegacyConfigFile(dir)
	require.NoError(t, err)
	require.False(t, found)
	require.Nil(t, data)

	zero := filepath.Join(dir, "zero.json")
	require.NoError(t, os.WriteFile(zero, nil, 0600))
	data, found, err = readLegacyConfigFile(zero)
	require.NoError(t, err)
	require.False(t, found)
	require.Nil(t, data)
}

func TestApplyLegacyConfig_Malformed(t *testing.T) {
	setTestHome(t, t.TempDir())
	c := New()
	path := filepath.Join(t.TempDir(), "config.json")

	err := c.applyLegacyConfig(path, []byte(`{"contexts":`))

	require.ErrorContains(t, err, readConfigurationFileErrorPrefix)
	require.ErrorContains(t, err, path)
	require.Equal(t, New(), c)
}

// legacyFixture builds a v4 config.json (inline, not a committed fixture) carrying a distinct
// marker string in every secret path, plus last_update_check_at and a context's feature_flags.
func legacyFixture() []byte {
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
				"active_global_api_key": "global1",
				"feature_flags": {
					"values": {"marker": "%s"},
					"ccloud_values": {},
					"last_update_time": 42
				},
				"global_api_keys": {
					"global1": {
						"api_key": "global-key",
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
		"MARKER-cred-api-secret", b64("MARKER-cred-salt"), b64("MARKER-cred-nonce"),
		"MARKER-feature-flags-value",
		"MARKER-global-api-secret",
		"MARKER-kafka-cluster-configs-secret",
		"MARKER-kafka-env-contexts-secret",
		"MARKER-sr-credential-secret",
		"MARKER-auth-token", "MARKER-auth-refresh-token", b64("MARKER-state-salt"), b64("MARKER-state-nonce"),
		"MARKER-saved-password", b64("MARKER-saved-salt"), b64("MARKER-saved-nonce"),
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

	envCfg := ctx.KafkaClusterContext.KafkaEnvContexts["env1"].KafkaClusterConfigs["cluster2"]
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

	// non-secret fields survive the decode untouched.
	require.Equal(t, "ctx1", c.CurrentContext)
	require.Equal(t, "platform1", ctx.PlatformName)
	require.Equal(t, "cluster1", ctx.KafkaClusterContext.ActiveKafkaCluster)
	require.Equal(t, "global1", ctx.ActiveGlobalAPIKey)
}

func TestApplyLegacyConfig_StatefulFixtures(t *testing.T) {
	for _, fixture := range []string{"stateful_cloud.json", "stateful_onprem.json"} {
		t.Run(fixture, func(t *testing.T) {
			setTestHome(t, t.TempDir())
			c := New()
			path := filepath.Join(t.TempDir(), "config.json")

			data, err := os.ReadFile(filepath.Join("test_json", fixture))
			require.NoError(t, err)

			require.NoError(t, c.applyLegacyConfig(path, data))

			ctx := c.Contexts["my-context"]
			require.Equal(t, "def-secret-456", c.Credentials["api-key-abc-key-123"].APIKeyPair.Secret)

			var kafkaKey *APIKeyPair
			if ctx.KafkaClusterContext.KafkaClusterConfigs != nil {
				kafkaKey = ctx.KafkaClusterContext.KafkaClusterConfigs["anonymous-id"].APIKeys["abc-key-123"]
			} else {
				kafkaKey = ctx.KafkaClusterContext.KafkaEnvContexts["env-123456"].KafkaClusterConfigs["anonymous-id"].APIKeys["abc-key-123"]
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
// persisted (non json:"-") fields, and fails if any json:"-" field it finds under pkg/config's
// own types is covered by neither an overlay type nor the allowlist above. A future secret field
// retagged without a matching overlay entry fails this test.
func TestLegacyOverlay_CoversEveryRetaggedField(t *testing.T) {
	pkgPath := reflect.TypeOf(Config{}).PkgPath()
	visited := map[reflect.Type]bool{}

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
}
