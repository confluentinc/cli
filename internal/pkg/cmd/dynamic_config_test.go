package cmd_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/form"
	pmock "github.com/confluentinc/cli/internal/pkg/mock"
)

func TestDynamicConfig_ParseFlagsIntoConfig(t *testing.T) {
	config := v3.AuthenticatedCloudConfigMock()
	dynamicConfigBase := pcmd.NewDynamicConfig(config, &pcmd.FlagResolverImpl{
		Prompt: &form.RealPrompt{},
		Out:    os.Stdout,
	}, pmock.NewClientMock())

	config = v3.AuthenticatedCloudConfigMock()
	dynamicConfigFlag := pcmd.NewDynamicConfig(config, &pcmd.FlagResolverImpl{
		Prompt: &form.RealPrompt{},
		Out:    os.Stdout,
	}, pmock.NewClientMock())
	dynamicConfigFlag.Contexts["test-context"] = &v3.Context{
		Name: "test-context",
	}
	tests := []struct {
		name           string
		context        string
		dConfig        *pcmd.DynamicConfig
		errMsg         string
		suggestionsMsg string
	}{
		{
			name:    "read context from config",
			dConfig: dynamicConfigBase,
		},
		{
			name:    "read context from flag",
			context: "test-context",
			dConfig: dynamicConfigFlag,
		},
		{
			name:    "bad-context specified with flag",
			context: "bad-context",
			dConfig: dynamicConfigFlag,
			errMsg:  fmt.Sprintf(errors.ContextNotExistErrorMsg, "bad-context"),
		},
	}
	for _, tt := range tests {
		cmd := &cobra.Command{
			Run: func(cmd *cobra.Command, args []string) {},
		}
		cmd.Flags().String("context", "", "Context name.")
		err := cmd.ParseFlags([]string{"--context", tt.context})
		require.NoError(t, err)
		initialCurrentContext := tt.dConfig.CurrentContext
		err = tt.dConfig.ParseFlagsIntoConfig(cmd)
		if tt.errMsg != "" {
			require.Error(t, err)
			require.Equal(t, tt.errMsg, err.Error())
			if tt.suggestionsMsg != "" {
				errors.VerifyErrorAndSuggestions(require.New(t), err, tt.errMsg, tt.suggestionsMsg)
			}
		} else {
			require.NoError(t, err)
			ctx, err := tt.dConfig.Context(cmd)
			require.NoError(t, err)
			if tt.context != "" {
				require.Equal(t, tt.context, ctx.Name)
			} else {
				require.Equal(t, initialCurrentContext, ctx.Name)
			}
		}
	}
}
