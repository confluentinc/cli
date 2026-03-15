package config

import (
	"fmt"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	ccloudv1 "github.com/confluentinc/ccloud-sdk-go-v1-public"
	ccloudv1mock "github.com/confluentinc/ccloud-sdk-go-v1-public/mock"
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

func TestSwitchOrganization(t *testing.T) {
	cfg := AuthenticatedCloudConfigMock()
	ctx := cfg.Context()

	oldRefreshToken := "old-refresh-token"
	ctx.State.AuthRefreshToken = oldRefreshToken

	newOrgId := "org-new-id"
	newToken := "new-auth-token"
	newRefreshToken := "new-refresh-token"

	auth := &ccloudv1mock.Auth{
		LoginFunc: func(req *ccloudv1.AuthenticateRequest) (*ccloudv1.AuthenticateReply, error) {
			require.Equal(t, newOrgId, req.OrgResourceId)
			require.Equal(t, oldRefreshToken, req.RefreshToken)
			return &ccloudv1.AuthenticateReply{
				Token:        newToken,
				RefreshToken: newRefreshToken,
			}, nil
		},
	}
	client := &ccloudv1.Client{Auth: auth}

	err := ctx.SwitchOrganization(client, newOrgId)
	require.NoError(t, err)

	require.Equal(t, newOrgId, ctx.LastOrgId)
	require.Equal(t, newToken, ctx.State.AuthToken)
	require.Equal(t, newRefreshToken, ctx.State.AuthRefreshToken)
}

func TestSwitchOrganization_ClearsEnvironmentState(t *testing.T) {
	cfg := AuthenticatedCloudConfigMock()
	ctx := cfg.Context()
	ctx.State.AuthRefreshToken = "refresh-token"

	// Pre-populate environment and Kafka state
	ctx.CurrentEnvironment = "env-123"
	ctx.Environments["env-123"] = &EnvironmentContext{CurrentFlinkComputePool: "pool-1"}
	ctx.KafkaClusterContext.KafkaEnvContexts["env-123"] = &KafkaEnvContext{
		ActiveKafkaCluster:  "lkc-999",
		KafkaClusterConfigs: map[string]*KafkaClusterConfig{"lkc-999": {ID: "lkc-999"}},
	}

	auth := &ccloudv1mock.Auth{
		LoginFunc: func(_ *ccloudv1.AuthenticateRequest) (*ccloudv1.AuthenticateReply, error) {
			return &ccloudv1.AuthenticateReply{Token: "t", RefreshToken: "r"}, nil
		},
	}
	client := &ccloudv1.Client{Auth: auth}

	err := ctx.SwitchOrganization(client, "org-other")
	require.NoError(t, err)

	require.Empty(t, ctx.CurrentEnvironment)
	require.Empty(t, ctx.Environments)
	require.Empty(t, ctx.KafkaClusterContext.ActiveKafkaCluster)
	require.Empty(t, ctx.KafkaClusterContext.ActiveKafkaClusterEndpoint)
	require.Empty(t, ctx.KafkaClusterContext.KafkaClusterConfigs)
	require.Empty(t, ctx.KafkaClusterContext.KafkaEnvContexts)
}

func TestSwitchOrganization_RollbackOnFailure(t *testing.T) {
	cfg := AuthenticatedCloudConfigMock()
	ctx := cfg.Context()

	oldOrgId := ctx.LastOrgId
	oldToken := ctx.State.AuthToken
	oldRefreshToken := "old-refresh-token"
	ctx.State.AuthRefreshToken = oldRefreshToken

	// Pre-populate environment state to verify it is NOT cleared on failure
	ctx.CurrentEnvironment = "env-123"
	ctx.Environments["env-123"] = &EnvironmentContext{}

	auth := &ccloudv1mock.Auth{
		LoginFunc: func(_ *ccloudv1.AuthenticateRequest) (*ccloudv1.AuthenticateReply, error) {
			return nil, fmt.Errorf("auth failed")
		},
	}
	client := &ccloudv1.Client{Auth: auth}

	err := ctx.SwitchOrganization(client, "org-other")
	require.Error(t, err)
	require.Contains(t, err.Error(), "auth failed")

	// Verify full rollback
	require.Equal(t, oldOrgId, ctx.LastOrgId)
	require.Equal(t, oldToken, ctx.State.AuthToken)
	require.Equal(t, oldRefreshToken, ctx.State.AuthRefreshToken)

	// Verify environment state was NOT touched
	require.Equal(t, "env-123", ctx.CurrentEnvironment)
	require.Contains(t, ctx.Environments, "env-123")
}

func TestSwitchOrganization_NilKafkaClusterContext(t *testing.T) {
	cfg := AuthenticatedCloudConfigMock()
	ctx := cfg.Context()
	ctx.State.AuthRefreshToken = "refresh-token"
	ctx.KafkaClusterContext = nil

	auth := &ccloudv1mock.Auth{
		LoginFunc: func(_ *ccloudv1.AuthenticateRequest) (*ccloudv1.AuthenticateReply, error) {
			return &ccloudv1.AuthenticateReply{Token: "t", RefreshToken: "r"}, nil
		},
	}
	client := &ccloudv1.Client{Auth: auth}

	err := ctx.SwitchOrganization(client, "org-other")
	require.NoError(t, err)

	require.Equal(t, "org-other", ctx.LastOrgId)
	require.Empty(t, ctx.CurrentEnvironment)
	require.Nil(t, ctx.KafkaClusterContext)
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
