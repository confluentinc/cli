package dynamicconfig

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/hashicorp/go-version"
	"github.com/stretchr/testify/require"

	cmkv2 "github.com/confluentinc/ccloud-sdk-go-v2/cmk/v2"
	cmkmock "github.com/confluentinc/ccloud-sdk-go-v2/cmk/v2/mock"

	"github.com/confluentinc/cli/internal/pkg/ccloudv2"
	"github.com/confluentinc/cli/internal/pkg/config"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
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
