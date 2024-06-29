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
