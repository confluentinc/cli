package dynamicconfig

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/hashicorp/go-version"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	cmkv2 "github.com/confluentinc/ccloud-sdk-go-v2/cmk/v2"
	cmkmock "github.com/confluentinc/ccloud-sdk-go-v2/cmk/v2/mock"

	"github.com/confluentinc/cli/internal/pkg/ccloudv2"
	"github.com/confluentinc/cli/internal/pkg/config"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	pmock "github.com/confluentinc/cli/internal/pkg/mock"
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
		Context: &v1.Context{
			KafkaClusterContext: &v1.KafkaClusterContext{
				KafkaClusterConfigs: map[string]*v1.KafkaClusterConfig{
					"lkc-123456": {LastUpdate: update, Bootstrap: "pkc-abc12.us-west-2.aws.confluent.cloud:1234"},
				},
			},
		},
	}

	config, err := d.FindKafkaCluster("lkc-123456")
	require.NoError(t, err)
	require.True(t, config.LastUpdate.Equal(update))
}

func TestFindKafkaCluster_Expired(t *testing.T) {
	update := time.Now().Add(-7 * 24 * time.Hour)

	d := &DynamicContext{
		Context: &v1.Context{
			CurrentEnvironment:  "env-123456",
			KafkaClusterContext: &v1.KafkaClusterContext{KafkaClusterConfigs: map[string]*v1.KafkaClusterConfig{"lkc-123456": {LastUpdate: update}}},
			Credential:          &v1.Credential{CredentialType: v1.Username},
			State:               &v1.ContextState{AuthToken: "token"},
			Config:              &v1.Config{BaseConfig: &config.BaseConfig{Ver: config.Version{Version: &version.Version{}}}},
		},
		V2Client: &ccloudv2.Client{
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
		},
	}

	config, err := d.FindKafkaCluster("lkc-123456")
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
	cfg := v1.AuthenticatedCloudConfigMock()
	return NewDynamicContext(cfg.Context(), pmock.NewV2ClientMock())
}

func getClusterFlagContext() *DynamicContext {
	config := v1.AuthenticatedCloudConfigMock()
	clusterFlagContext := NewDynamicContext(config.Context(), pmock.NewV2ClientMock())
	// create cluster that will be used in "--cluster" flag value
	clusterFlagContext.KafkaClusterContext.KafkaEnvContexts["testAccount"].KafkaClusterConfigs[flagCluster] = &v1.KafkaClusterConfig{
		ID:   flagCluster,
		Name: "miles",
	}
	return clusterFlagContext
}

func getEnvFlagContext() *DynamicContext {
	config := v1.AuthenticatedCloudConfigMock()
	envFlagContext := NewDynamicContext(config.Context(), pmock.NewV2ClientMock())
	envFlagContext.Environments[flagEnvironment] = &v1.EnvironmentContext{}
	return envFlagContext
}

func getEnvAndClusterFlagContext() *DynamicContext {
	envAndClusterFlagContext := getEnvFlagContext()

	envAndClusterFlagContext.KafkaClusterContext.KafkaEnvContexts[flagEnvironment] = &v1.KafkaEnvContext{
		ActiveKafkaCluster:  "",
		KafkaClusterConfigs: map[string]*v1.KafkaClusterConfig{},
	}
	envAndClusterFlagContext.KafkaClusterContext.KafkaEnvContexts[flagEnvironment].KafkaClusterConfigs[flagClusterInEnv] = &v1.KafkaClusterConfig{
		ID:   flagClusterInEnv,
		Name: "miles2",
	}
	return envAndClusterFlagContext
}
