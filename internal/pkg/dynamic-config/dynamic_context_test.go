package dynamicconfig

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	ccloudv1 "github.com/confluentinc/ccloud-sdk-go-v1-public"
	ccloudv1mock "github.com/confluentinc/ccloud-sdk-go-v1-public/mock"
	cmkv2 "github.com/confluentinc/ccloud-sdk-go-v2/cmk/v2"
	cmkmock "github.com/confluentinc/ccloud-sdk-go-v2/cmk/v2/mock"
	"github.com/hashicorp/go-version"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	"github.com/confluentinc/cli/internal/pkg/ccloudv2"
	"github.com/confluentinc/cli/internal/pkg/config"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/errors"
	pmock "github.com/confluentinc/cli/internal/pkg/mock"
)

var (
	flagEnvironment  = "env-test"
	flagCluster      = "lkc-0001"
	flagClusterInEnv = "lkc-0002"
	badFlagEnv       = "bad-env"
	apiEnvironment   = "env-from-api-call"
)

func TestFindKafkaCluster_Unexpired(t *testing.T) {
	t.Parallel()

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
	t.Parallel()

	update := time.Now().Add(-7 * 24 * time.Hour)

	d := &DynamicContext{
		Context: &v1.Context{
			KafkaClusterContext: &v1.KafkaClusterContext{
				KafkaClusterConfigs: map[string]*v1.KafkaClusterConfig{
					"lkc-123456": {LastUpdate: update},
				},
			},
			Credential: &v1.Credential{CredentialType: v1.Username},
			State: &v1.ContextState{
				Auth:      &v1.AuthConfig{Account: &ccloudv1.Account{Id: "env-123456"}},
				AuthToken: "token",
			},
			Config: &v1.Config{BaseConfig: &config.BaseConfig{Ver: config.Version{Version: &version.Version{}}}},
		},
		V2Client: &ccloudv2.Client{
			CmkClient: &cmkv2.APIClient{
				ClustersCmkV2Api: &cmkmock.ClustersCmkV2Api{
					GetCmkV2ClusterFunc: func(_ context.Context, _ string) cmkv2.ApiGetCmkV2ClusterRequest {
						return cmkv2.ApiGetCmkV2ClusterRequest{}
					},
					GetCmkV2ClusterExecuteFunc: func(_ cmkv2.ApiGetCmkV2ClusterRequest) (cmkv2.CmkV2Cluster, *http.Response, error) {
						cluster := cmkv2.CmkV2Cluster{
							Id: stringPtr("lkc-123456"),
							Spec: &cmkv2.CmkV2ClusterSpec{
								DisplayName:            stringPtr(""),
								KafkaBootstrapEndpoint: stringPtr(""),
								HttpEndpoint:           stringPtr(""),
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

func stringPtr(s string) *string {
	return &s
}

func TestDynamicContext_ParseFlagsIntoContext(t *testing.T) {
	t.Parallel()

	client := buildCcloudMockClient()
	tests := []struct {
		name           string
		ctx            *DynamicContext
		cluster        string
		environment    string
		errMsg         string
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
			name:        "environment not found",
			environment: badFlagEnv,
			ctx:         getEnvFlagContext(),
			errMsg:      fmt.Sprintf(errors.EnvironmentNotFoundErrorMsg, badFlagEnv, getEnvFlagContext().Name),
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
	for _, tt := range tests {
		cmd := &cobra.Command{
			Run: func(cmd *cobra.Command, args []string) {},
		}
		cmd.Flags().String("environment", "", "Environment ID.")
		cmd.Flags().String("cluster", "", "Kafka cluster ID.")
		err := cmd.ParseFlags([]string{"--cluster", tt.cluster, "--environment", tt.environment})
		require.NoError(t, err)
		initialEnvId := tt.ctx.GetEnvironment().GetId()
		initialActiveKafkaId := tt.ctx.KafkaClusterContext.GetActiveKafkaClusterId()
		err = tt.ctx.ParseFlagsIntoContext(cmd, client)
		if tt.errMsg != "" {
			require.Error(t, err)
			require.Equal(t, tt.errMsg, err.Error())
			if tt.suggestionsMsg != "" {
				errors.VerifyErrorAndSuggestions(require.New(t), err, tt.errMsg, tt.suggestionsMsg)
			}
		} else {
			require.NoError(t, err)
			finalEnv := tt.ctx.GetEnvironment().GetId()
			finalCluster := tt.ctx.KafkaClusterContext.GetActiveKafkaClusterId()
			if tt.environment != "" {
				require.Equal(t, tt.environment, finalEnv)
			} else {
				require.Equal(t, initialEnvId, finalEnv)
			}
			if tt.cluster != "" {
				require.Equal(t, tt.cluster, finalCluster)
			} else if tt.environment == "" {
				require.Equal(t, initialActiveKafkaId, finalCluster)
			}
		}
	}
}

func buildCcloudMockClient() *ccloudv1.Client {
	client := pmock.NewClientMock()
	client.Account = &ccloudv1mock.AccountInterface{ListFunc: func(ctx context.Context, account *ccloudv1.Account) ([]*ccloudv1.Account, error) {
		return []*ccloudv1.Account{{Id: apiEnvironment}}, nil
	}}
	return client
}

func getBaseContext() *DynamicContext {
	cfg := v1.AuthenticatedCloudConfigMock()
	return NewDynamicContext(cfg.Context(), pmock.NewClientMock(), pmock.NewV2ClientMock())
}

func getClusterFlagContext() *DynamicContext {
	config := v1.AuthenticatedCloudConfigMock()
	clusterFlagContext := NewDynamicContext(config.Context(), pmock.NewClientMock(), pmock.NewV2ClientMock())
	// create cluster that will be used in "--cluster" flag value
	clusterFlagContext.KafkaClusterContext.KafkaEnvContexts["testAccount"].KafkaClusterConfigs[flagCluster] = &v1.KafkaClusterConfig{
		ID:   flagCluster,
		Name: "miles",
	}
	return clusterFlagContext
}

func getEnvFlagContext() *DynamicContext {
	config := v1.AuthenticatedCloudConfigMock()
	envFlagContext := NewDynamicContext(config.Context(), pmock.NewClientMock(), pmock.NewV2ClientMock())
	envFlagContext.State.Auth.Accounts = append(envFlagContext.State.Auth.Accounts, &ccloudv1.Account{Name: flagEnvironment, Id: flagEnvironment})
	return envFlagContext
}

func getEnvAndClusterFlagContext() *DynamicContext {
	config := v1.AuthenticatedCloudConfigMock()
	envAndClusterFlagContext := NewDynamicContext(config.Context(), pmock.NewClientMock(), pmock.NewV2ClientMock())

	envAndClusterFlagContext.State.Auth.Accounts = append(envAndClusterFlagContext.State.Auth.Accounts, &ccloudv1.Account{Name: flagEnvironment, Id: flagEnvironment})
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
