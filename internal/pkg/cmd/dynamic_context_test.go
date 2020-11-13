package cmd_test

import (
	"fmt"
	"os"
	"testing"

	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/form"
	pmock "github.com/confluentinc/cli/internal/pkg/mock"
)

var (
	flagEnvironment  = "env-test"
	flagCluster      = "lkc-0001"
	flagClusterInEnv = "lkc-0002"
	badFlagEnv       = "bad-env"
)

func TestDynamicContext_ParseFlagsIntoContext(t *testing.T) {
	tests := []struct {
		name           string
		ctx            *pcmd.DynamicContext
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
	}
	for _, tt := range tests {
		cmd := &cobra.Command{
			Run: func(cmd *cobra.Command, args []string) {},
		}
		cmd.Flags().String("environment", "", "Environment ID.")
		cmd.Flags().String("cluster", "", "Kafka cluster ID.")
		err := cmd.ParseFlags([]string{"--cluster", tt.cluster, "--environment", tt.environment})
		require.NoError(t, err)
		initialEnvId := tt.ctx.GetCurrentEnvironmentId()
		initialActiveKafkaId := tt.ctx.KafkaClusterContext.GetActiveKafkaClusterId()
		err = tt.ctx.ParseFlagsIntoContext(cmd)
		if tt.errMsg != "" {
			require.Error(t, err)
			require.Equal(t, tt.errMsg, err.Error())
			if tt.suggestionsMsg != "" {
				errors.VerifyErrorAndSuggestions(require.New(t), err, tt.errMsg, tt.suggestionsMsg)
			}
		} else {
			require.NoError(t, err)
			finalEnv := tt.ctx.GetCurrentEnvironmentId()
			finalCluster := tt.ctx.KafkaClusterContext.GetActiveKafkaClusterId()
			if tt.environment != "" {
				require.Equal(t, tt.environment, finalEnv)
			} else {
				require.Equal(t, initialEnvId, finalEnv)
			}
			if tt.cluster != "" {
				require.Equal(t, tt.cluster, finalCluster)
			} else if tt.environment == ""{
				require.Equal(t, initialActiveKafkaId, finalCluster)
			}
		}
	}
}

func getBaseContext() *pcmd.DynamicContext {
	config := v3.AuthenticatedCloudConfigMock()
	context := pcmd.NewDynamicContext(config.Context(), &pcmd.FlagResolverImpl{
		Prompt: &form.RealPrompt{},
		Out:    os.Stdout,
	}, pmock.NewClientMock())
	return context
}

func getClusterFlagContext() *pcmd.DynamicContext {
	config := v3.AuthenticatedCloudConfigMock()
	clusterFlagContext := pcmd.NewDynamicContext(config.Context(), &pcmd.FlagResolverImpl{
		Prompt: &form.RealPrompt{},
		Out:    os.Stdout,
	}, pmock.NewClientMock())
	// create cluster that will be used in "--cluster" flag value
	clusterFlagContext.KafkaClusterContext.KafkaEnvContexts["testAccount"].KafkaClusterConfigs[flagCluster] = &v1.KafkaClusterConfig{
		ID:   flagCluster,
		Name: "miles",
	}
	return clusterFlagContext
}

func getEnvFlagContext() *pcmd.DynamicContext {
	config := v3.AuthenticatedCloudConfigMock()
	envFlagContext := pcmd.NewDynamicContext(config.Context(), &pcmd.FlagResolverImpl{
		Prompt: &form.RealPrompt{},
		Out:    os.Stdout,
	}, pmock.NewClientMock())
	envFlagContext.State.Auth.Accounts = append(envFlagContext.State.Auth.Accounts, &orgv1.Account{Name: flagEnvironment, Id: flagEnvironment})
	return envFlagContext
}

func getEnvAndClusterFlagContext() *pcmd.DynamicContext {
	config := v3.AuthenticatedCloudConfigMock()
	envAndClusterFlagContext := pcmd.NewDynamicContext(config.Context(), &pcmd.FlagResolverImpl{
		Prompt: &form.RealPrompt{},
		Out:    os.Stdout,
	}, pmock.NewClientMock())

	envAndClusterFlagContext.State.Auth.Accounts = append(envAndClusterFlagContext.State.Auth.Accounts, &orgv1.Account{Name: flagEnvironment, Id: flagEnvironment})
	envAndClusterFlagContext.KafkaClusterContext.KafkaEnvContexts[flagEnvironment] = &v3.KafkaEnvContext{
		ActiveKafkaCluster:  "",
		KafkaClusterConfigs: map[string]*v1.KafkaClusterConfig{},
	}
	envAndClusterFlagContext.KafkaClusterContext.KafkaEnvContexts[flagEnvironment].KafkaClusterConfigs[flagClusterInEnv] = &v1.KafkaClusterConfig{
		ID:   flagClusterInEnv,
		Name: "miles2",
	}
	return envAndClusterFlagContext
}
