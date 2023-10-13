package dynamicconfig

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	cmkv2 "github.com/confluentinc/ccloud-sdk-go-v2/cmk/v2"
	cmkmock "github.com/confluentinc/ccloud-sdk-go-v2/cmk/v2/mock"

	"github.com/confluentinc/cli/v3/pkg/ccloudv2"
	"github.com/confluentinc/cli/v3/pkg/config"
)

var (
	flagEnvironment  = "env-test"
	flagCluster      = "lkc-0001"
	flagClusterInEnv = "lkc-0002"
	apiEnvironment   = "env-from-api-call"
)

func TestFindKafkaCluster_Unexpired(t *testing.T) {
	update := time.Now()

	d := &DynamicContext{
		Context: &config.Context{
			KafkaClusterContext: &config.KafkaClusterContext{
				KafkaClusterConfigs: map[string]*config.KafkaClusterConfig{
					"lkc-123456": {LastUpdate: update, Bootstrap: "pkc-abc12.us-west-2.aws.confluent.cloud:1234"},
				},
			},
		},
	}

	config, err := d.FindKafkaCluster(nil, "lkc-123456")
	require.NoError(t, err)
	require.True(t, config.LastUpdate.Equal(update))
}

func TestFindKafkaCluster_Expired(t *testing.T) {
	update := time.Now().Add(-7 * 24 * time.Hour)

	d := &DynamicContext{
		Context: &config.Context{
			CurrentEnvironment:  "env-123456",
			KafkaClusterContext: &config.KafkaClusterContext{KafkaClusterConfigs: map[string]*config.KafkaClusterConfig{"lkc-123456": {LastUpdate: update}}},
			Credential:          &config.Credential{CredentialType: config.Username},
			State:               &config.ContextState{AuthToken: "token"},
			Config:              &config.Config{},
		},
	}

	client := &ccloudv2.Client{
		CmkClient: &cmkv2.APIClient{
			ClustersCmkV2Api: &cmkmock.ClustersCmkV2Api{
				GetCmkV2ClusterFunc: func(_ context.Context, _ string) cmkv2.ApiGetCmkV2ClusterRequest {
					return cmkv2.ApiGetCmkV2ClusterRequest{}
				},
				GetCmkV2ClusterExecuteFunc: func(_ cmkv2.ApiGetCmkV2ClusterRequest) (cmkv2.CmkV2Cluster, *http.Response, error) {
					cluster := cmkv2.CmkV2Cluster{
						Id: cmkv2.PtrString("lkc-123456"),
						Spec: &cmkv2.CmkV2ClusterSpec{
							DisplayName:            cmkv2.PtrString(""),
							KafkaBootstrapEndpoint: cmkv2.PtrString(""),
							HttpEndpoint:           cmkv2.PtrString(""),
						},
					}
					return cluster, nil, nil
				},
			},
		},
	}

	config, err := d.FindKafkaCluster(client, "lkc-123456")
	require.NoError(t, err)
	require.True(t, config.LastUpdate.After(update))
}

func TestDynamicContext_ParseFlagsIntoContext(t *testing.T) {
	tests := []struct {
		name           string
		ctx            *DynamicContext
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
			environment: apiEnvironment,
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

func getBaseContext() *DynamicContext {
	cfg := config.AuthenticatedCloudConfigMock()
	return NewDynamicContext(cfg.Context())
}

func getClusterFlagContext() *DynamicContext {
	cfg := config.AuthenticatedCloudConfigMock()
	clusterFlagContext := NewDynamicContext(cfg.Context())
	// create cluster that will be used in "--cluster" flag value
	clusterFlagContext.KafkaClusterContext.KafkaEnvContexts["testAccount"].KafkaClusterConfigs[flagCluster] = &config.KafkaClusterConfig{
		ID:   flagCluster,
		Name: "miles",
	}
	return clusterFlagContext
}

func getEnvFlagContext() *DynamicContext {
	cfg := config.AuthenticatedCloudConfigMock()
	envFlagContext := NewDynamicContext(cfg.Context())
	envFlagContext.Environments[flagEnvironment] = &config.EnvironmentContext{}
	return envFlagContext
}

func getEnvAndClusterFlagContext() *DynamicContext {
	envAndClusterFlagContext := getEnvFlagContext()

	envAndClusterFlagContext.KafkaClusterContext.KafkaEnvContexts[flagEnvironment] = &config.KafkaEnvContext{
		ActiveKafkaCluster:  "",
		KafkaClusterConfigs: map[string]*config.KafkaClusterConfig{},
	}
	envAndClusterFlagContext.KafkaClusterContext.KafkaEnvContexts[flagEnvironment].KafkaClusterConfigs[flagClusterInEnv] = &config.KafkaClusterConfig{
		ID:   flagClusterInEnv,
		Name: "miles2",
	}
	return envAndClusterFlagContext
}
