package dynamicconfig

import (
	"fmt"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/errors"
	pmock "github.com/confluentinc/cli/internal/pkg/mock"
)

func TestDynamicConfig_ParseFlagsIntoConfig(t *testing.T) {
	config := v1.AuthenticatedCloudConfigMock()
	dynamicConfigBase := New(config, pmock.NewV2ClientMock())

	config = v1.AuthenticatedCloudConfigMock()
	dynamicConfigFlag := New(config, pmock.NewV2ClientMock())
	dynamicConfigFlag.Contexts["test-context"] = &v1.Context{
		Name: "test-context",
	}
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
	for _, tt := range tests {
		cmd := &cobra.Command{
			Run: func(cmd *cobra.Command, args []string) {},
		}
		cmd.Flags().String("context", "", "Context name.")
		err := cmd.ParseFlags([]string{"--context", tt.context})
		require.NoError(t, err)
		initialCurrentContext := tt.dynamicConfig.CurrentContext
		err = tt.dynamicConfig.ParseFlagsIntoConfig(cmd)
		if tt.errMsg != "" {
			require.Error(t, err)
			require.Equal(t, tt.errMsg, err.Error())
			if tt.suggestionsMsg != "" {
				errors.VerifyErrorAndSuggestions(require.New(t), err, tt.errMsg, tt.suggestionsMsg)
			}
		} else {
			require.NoError(t, err)
			ctx := tt.dynamicConfig.Context()
			if tt.context != "" {
				require.Equal(t, tt.context, ctx.Name)
			} else {
				require.Equal(t, initialCurrentContext, ctx.Name)
			}
		}
	}
}
