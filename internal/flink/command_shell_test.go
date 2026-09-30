package flink

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"

	cmfsdk "github.com/confluentinc/cmf-sdk-go/v1"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/config"
	"github.com/confluentinc/cli/v4/pkg/version"
)

const (
	testInitialAuthToken = "initial-auth-token"
	testUpdatedAuthToken = "updated-auth-token"
	testRefreshToken     = "refresh-token"
)

func TestAuthenticatedOnPrem_IsOnPremLogin(t *testing.T) {
	cfg := config.AuthenticatedOnPremConfigMock()
	ctx := cfg.Context()
	ctx.State = &config.ContextState{
		AuthToken:        testInitialAuthToken,
		AuthRefreshToken: testRefreshToken,
	}

	// These flags must be defined to create the cmf client
	cmd := &cobra.Command{}
	cmd.Flags().Bool("unsafe-trace", false, "")
	addCmfFlagSet(cmd)

	c := &command{
		AuthenticatedCLICommand: &pcmd.AuthenticatedCLICommand{
			CLICommand: &pcmd.CLICommand{
				Command: cmd,
				Config:  cfg,
				Version: &version.Version{},
			},
			Context: ctx,
		},
	}
	assert.True(t, c.Config.IsOnPremLogin())

	// Set up the client and set the initial auth token as in startFlinkSqlClientOnPrem
	cmfClient, err := c.GetCmfClient(cmd)
	assert.NoError(t, err)

	cmfClient.AuthToken = c.Context.GetAuthToken()
	assert.Equal(t, testInitialAuthToken, cmfClient.AuthToken)

	// the actual prerunner.AuthenticatedWithMDS updates the tokens, so we emulate that here
	mockAuthenticatedWithMDS := func(_ *cobra.Command, _ []string) error {
		return c.Context.UpdateAuthTokens(testUpdatedAuthToken, testRefreshToken)
	}

	err = c.authenticatedOnPrem(mockAuthenticatedWithMDS, cmd)()
	assert.NoError(t, err)

	// Verify that the auth token update was propagated to the cmf client
	assert.Equal(t, testUpdatedAuthToken, cmfClient.AuthToken)

	// Sanity check the CmfApiContext call used in the Store methods
	// And also compare it to the c.createContext call used by the non-shell commands
	cmfContext := cmfClient.CmfApiContext()
	cmfCtxAuthTokenValue, ok := cmfContext.Value(cmfsdk.ContextAccessToken).(string)
	assert.True(t, ok)
	assert.Equal(t, testUpdatedAuthToken, cmfCtxAuthTokenValue)

	assert.Equal(t, c.createContext(), cmfContext)
}
