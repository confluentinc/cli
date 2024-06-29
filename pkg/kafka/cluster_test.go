package kafka

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	cmkv2 "github.com/confluentinc/ccloud-sdk-go-v2/cmk/v2"
	cmkmock "github.com/confluentinc/ccloud-sdk-go-v2/cmk/v2/mock"

	"github.com/confluentinc/cli/v3/pkg/ccloudv2"
	"github.com/confluentinc/cli/v3/pkg/config"
)

func TestFindCluster_Unexpired(t *testing.T) {
	update := time.Now()

	ctx := &config.Context{
		KafkaClusterContext: &config.KafkaClusterContext{
			KafkaClusterConfigs: map[string]*config.KafkaClusterConfig{
				"lkc-123456": {
					LastUpdate: update,
					Bootstrap:  "pkc-abc12.us-west-2.aws.confluent.cloud:1234",
				},
			},
		},
	}

	config, err := FindCluster(nil, ctx, "lkc-123456")
	require.NoError(t, err)
	require.True(t, config.LastUpdate.Equal(update))
}

func TestFindCluster_Expired(t *testing.T) {
	update := time.Now().Add(-7 * 24 * time.Hour)

	ctx := &config.Context{
		CurrentEnvironment:  "env-123456",
		KafkaClusterContext: &config.KafkaClusterContext{KafkaClusterConfigs: map[string]*config.KafkaClusterConfig{"lkc-123456": {LastUpdate: update}}},
		Credential:          &config.Credential{CredentialType: config.Username},
		State:               &config.ContextState{AuthToken: "token"},
		Config:              &config.Config{},
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

	config, err := FindCluster(client, ctx, "lkc-123456")
	require.NoError(t, err)
	require.True(t, config.LastUpdate.After(update))
}
