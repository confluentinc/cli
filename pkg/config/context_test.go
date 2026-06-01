package config

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

var (
	flagEnvironment  = "env-test"
	flagCluster      = "lkc-0001"
	flagClusterInEnv = "lkc-0002"
)

func TestParseFlagsIntoContext(t *testing.T) {
	tests := []struct {
		name           string
		ctx            *Context
		cluster        string
		environment    string
		suggestionsMsg string
	}{
		{
			name: "read cluster from config",
			ctx:  getBaseContext(),
		},
		{
			name:    "read cluster from flag",
			ctx:     getClusterFlagContext(),
			cluster: flagCluster,
		},
		{
			name: "read environment from config",
			ctx:  getEnvFlagContext(),
		},
		{
			name:        "read environment from flag",
			environment: flagEnvironment,
			ctx:         getEnvFlagContext(),
		},
		{
			name:        "pass cluster and environment",
			cluster:     flagClusterInEnv,
			environment: flagEnvironment,
			ctx:         getEnvAndClusterFlagContext(),
		},
		{
			name:        "find environment from api call",
			cluster:     flagClusterInEnv,
			environment: "env-from-api-call",
			ctx:         getEnvFlagContext(),
		},
	}
	for _, test := range tests {
		cmd := &cobra.Command{Run: func(cmd *cobra.Command, args []string) {}}
		cmd.Flags().String("environment", "", "Environment ID.")
		cmd.Flags().String("cluster", "", "Kafka cluster ID.")
		err := cmd.ParseFlags([]string{"--cluster", test.cluster, "--environment", test.environment})
		require.NoError(t, err)
		initialEnvId := test.ctx.GetCurrentEnvironment()
		initialActiveKafkaId := test.ctx.KafkaClusterContext.GetActiveKafkaClusterId()
		err = test.ctx.ParseFlagsIntoContext(cmd)
		require.NoError(t, err)
		finalEnv := test.ctx.GetCurrentEnvironment()
		finalCluster := test.ctx.KafkaClusterContext.GetActiveKafkaClusterId()
		if test.environment != "" {
			require.Equal(t, test.environment, finalEnv)
		} else {
			require.Equal(t, initialEnvId, finalEnv)
		}
		if test.cluster != "" {
			require.Equal(t, test.cluster, finalCluster)
		} else if test.environment == "" {
			require.Equal(t, initialActiveKafkaId, finalCluster)
		}
	}
}

func getBaseContext() *Context {
	return AuthenticatedCloudConfigMock().Context()
}

func getClusterFlagContext() *Context {
	ctx := getBaseContext()
	// create cluster that will be used in "--cluster" flag value
	ctx.KafkaClusterContext.KafkaEnvContexts["testAccount"].KafkaClusterConfigs[flagCluster] = &KafkaClusterConfig{
		ID:   flagCluster,
		Name: "miles",
	}
	return ctx
}

func getEnvFlagContext() *Context {
	ctx := getBaseContext()
	ctx.Environments[flagEnvironment] = &EnvironmentContext{}
	return ctx
}

func getEnvAndClusterFlagContext() *Context {
	ctx := getEnvFlagContext()

	ctx.KafkaClusterContext.KafkaEnvContexts[flagEnvironment] = &KafkaEnvContext{
		KafkaClusterConfigs: map[string]*KafkaClusterConfig{
			flagClusterInEnv: {
				ID:   flagClusterInEnv,
				Name: "miles2",
			},
		},
	}

	return ctx
}

func TestContext_GlobalAPIKeyStorage(t *testing.T) {
	ctx := getBaseContext()
	require.NotNil(t, ctx.GlobalAPIKeys, "validate() should initialize GlobalAPIKeys")

	pair := &APIKeyPair{Key: "GLOBAL-KEY-1", Secret: "plain-secret"}
	require.NoError(t, ctx.StoreGlobalAPIKey(pair))

	require.True(t, ctx.HasGlobalAPIKey("GLOBAL-KEY-1"))
	require.False(t, ctx.HasGlobalAPIKey("missing"))

	// After Store, the secret is encrypted in place. Resolve must decrypt it.
	stored := ctx.GlobalAPIKeys["GLOBAL-KEY-1"]
	require.NotEqual(t, "plain-secret", stored.Secret, "secret should be encrypted at rest")

	require.NoError(t, ctx.SetActiveGlobalAPIKey("GLOBAL-KEY-1"))
	require.Equal(t, "GLOBAL-KEY-1", ctx.GetActiveGlobalAPIKey())

	// SetActiveGlobalAPIKey must reject keys that aren't stored.
	require.Error(t, ctx.SetActiveGlobalAPIKey("unknown-key"))
}

func TestContext_ResolveKafkaAPIKey_PrefersClusterScoped(t *testing.T) {
	ctx := getBaseContext()

	// Pre-populate a cluster-scoped key on the active cluster.
	kcc := ctx.KafkaClusterContext.GetActiveKafkaClusterConfig()
	require.NotNil(t, kcc)
	clusterKey := &APIKeyPair{Key: "CLUSTER-KEY", Secret: "cluster-secret"}
	require.NoError(t, clusterKey.EncryptSecret())
	kcc.APIKeys = map[string]*APIKeyPair{"CLUSTER-KEY": clusterKey}
	kcc.APIKey = "CLUSTER-KEY"

	// Also set up a global key + active marker.
	require.NoError(t, ctx.StoreGlobalAPIKey(&APIKeyPair{Key: "GLOBAL-KEY", Secret: "global-secret"}))
	require.NoError(t, ctx.SetActiveGlobalAPIKey("GLOBAL-KEY"))

	key, secret, err := ctx.ResolveKafkaAPIKey(kcc)
	require.NoError(t, err)
	require.Equal(t, "CLUSTER-KEY", key)
	require.Equal(t, "cluster-secret", secret)
}

func TestContext_ResolveKafkaAPIKey_FallsBackToGlobal(t *testing.T) {
	ctx := getBaseContext()

	kcc := ctx.KafkaClusterContext.GetActiveKafkaClusterConfig()
	require.NotNil(t, kcc)
	kcc.APIKey = ""
	kcc.APIKeys = map[string]*APIKeyPair{}

	require.NoError(t, ctx.StoreGlobalAPIKey(&APIKeyPair{Key: "GLOBAL-KEY", Secret: "global-secret"}))
	require.NoError(t, ctx.SetActiveGlobalAPIKey("GLOBAL-KEY"))

	key, secret, err := ctx.ResolveKafkaAPIKey(kcc)
	require.NoError(t, err)
	require.Equal(t, "GLOBAL-KEY", key)
	require.Equal(t, "global-secret", secret)
}

func TestContext_ResolveKafkaAPIKey_NoCredentialsConfigured(t *testing.T) {
	ctx := getBaseContext()
	kcc := ctx.KafkaClusterContext.GetActiveKafkaClusterConfig()
	require.NotNil(t, kcc)
	kcc.APIKey = ""
	kcc.APIKeys = map[string]*APIKeyPair{}

	key, secret, err := ctx.ResolveKafkaAPIKey(kcc)
	require.NoError(t, err)
	require.Empty(t, key)
	require.Empty(t, secret)
}

func TestContext_DeleteGlobalAPIKey_ClearsActive(t *testing.T) {
	ctx := getBaseContext()
	require.NoError(t, ctx.StoreGlobalAPIKey(&APIKeyPair{Key: "GLOBAL-KEY", Secret: "global-secret"}))
	require.NoError(t, ctx.SetActiveGlobalAPIKey("GLOBAL-KEY"))

	ctx.DeleteGlobalAPIKey("GLOBAL-KEY")

	require.False(t, ctx.HasGlobalAPIKey("GLOBAL-KEY"))
	require.Empty(t, ctx.GetActiveGlobalAPIKey(), "deleting active key should clear ActiveGlobalAPIKey")
}

func TestContext_ValidateGlobalAPIKeys_RemovesOrphanedActive(t *testing.T) {
	ctx := getBaseContext()
	// Active key references a pair that doesn't exist in the map.
	ctx.ActiveGlobalAPIKey = "ghost-key"
	ctx.validateGlobalAPIKeys()
	require.Empty(t, ctx.ActiveGlobalAPIKey, "validate should clear active key when not present in map")
}

func TestContext_StoreGlobalAPIKey_RejectsIncompletePair(t *testing.T) {
	ctx := getBaseContext()

	require.Error(t, ctx.StoreGlobalAPIKey(nil), "nil pair should be rejected")
	require.Error(t, ctx.StoreGlobalAPIKey(&APIKeyPair{Secret: "secret-only"}), "missing key should be rejected")
	require.Error(t, ctx.StoreGlobalAPIKey(&APIKeyPair{Key: "key-only"}), "missing secret should be rejected")

	// A malformed pair must not be persisted.
	require.False(t, ctx.HasGlobalAPIKey("key-only"))
	require.Empty(t, ctx.GlobalAPIKeys)
}

func TestContext_ResolveKafkaAPIKey_ErrorsWhenClusterSecretMissing(t *testing.T) {
	ctx := getBaseContext()

	// Cluster has an active key set, but its secret is not stored locally.
	kcc := ctx.KafkaClusterContext.GetActiveKafkaClusterConfig()
	require.NotNil(t, kcc)
	kcc.APIKey = "CLUSTER-KEY"
	kcc.APIKeys = map[string]*APIKeyPair{}

	// A usable Global key exists, but it must NOT be silently used to mask the broken cluster config.
	require.NoError(t, ctx.StoreGlobalAPIKey(&APIKeyPair{Key: "GLOBAL-KEY", Secret: "global-secret"}))
	require.NoError(t, ctx.SetActiveGlobalAPIKey("GLOBAL-KEY"))

	_, _, err := ctx.ResolveKafkaAPIKey(kcc)
	require.Error(t, err, "should surface the broken cluster-scoped config rather than fall back to Global")
}
