package dynamicconfig

import (
	"fmt"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	"github.com/confluentinc/cli/v3/pkg/config"
	"github.com/confluentinc/cli/v3/pkg/errors"
)

func TestDynamicConfig_ParseFlagsIntoConfig(t *testing.T) {
	dynamicConfigBase := New(config.AuthenticatedCloudConfigMock())

	dynamicConfigFlag := New(config.AuthenticatedCloudConfigMock())
	dynamicConfigFlag.Contexts["test-context"] = &config.Context{Name: "test-context"}

	tests := []struct {
		name           string
		context        string
		dynamicConfig  *DynamicConfig
		errMsg         string
		suggestionsMsg string
	}{
		{
			name:          "read context from config",
			dynamicConfig: dynamicConfigBase,
		},
		{
			name:          "read context from flag",
			context:       "test-context",
			dynamicConfig: dynamicConfigFlag,
		},
		{
			name:          "bad-context specified with flag",
			context:       "bad-context",
			dynamicConfig: dynamicConfigFlag,
			errMsg:        fmt.Sprintf(errors.ContextDoesNotExistErrorMsg, "bad-context"),
		},
	}
	for _, test := range tests {
		cmd := &cobra.Command{Run: func(cmd *cobra.Command, args []string) {}}
		cmd.Flags().String("context", "", "Context name.")
		err := cmd.ParseFlags([]string{"--context", test.context})
		require.NoError(t, err)
		initialCurrentContext := test.dynamicConfig.CurrentContext
		err = test.dynamicConfig.ParseFlagsIntoConfig(cmd)
		if test.errMsg != "" {
			require.Error(t, err)
			require.Equal(t, test.errMsg, err.Error())
			if test.suggestionsMsg != "" {
				errors.VerifyErrorAndSuggestions(require.New(t), err, test.errMsg, test.suggestionsMsg)
			}
		} else {
			require.NoError(t, err)
			ctx := test.dynamicConfig.Context()
			if test.context != "" {
				require.Equal(t, test.context, ctx.Name)
			} else {
				require.Equal(t, initialCurrentContext, ctx.Name)
			}
		}
	}
}
