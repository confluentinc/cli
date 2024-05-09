package logout

import (
	"os"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	"github.com/confluentinc/mds-sdk-go-public/mdsv1"

	climock "github.com/confluentinc/cli/v3/mock"
	pauth "github.com/confluentinc/cli/v3/pkg/auth"
	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/config"
)

const (
	testToken      = "y0ur.jwt.T0kEn"
	promptUser     = "prompt-user@confluent.io"
	promptPassword = " prompt-password "
	ccloudURL      = "https://confluent.cloud"
)

var (
	orgManagerImpl           = pauth.NewLoginOrganizationManagerImpl()
	LoginOrganizationManager = &climock.LoginOrganizationManager{
		GetLoginOrganizationFromFlagFunc: func(cmd *cobra.Command) func() string {
			return orgManagerImpl.GetLoginOrganizationFromFlag(cmd)
		},
		GetLoginOrganizationFromEnvironmentVariableFunc: func() func() string {
			return orgManagerImpl.GetLoginOrganizationFromEnvironmentVariable()
		},
	}
	AuthTokenHandler = &climock.AuthTokenHandler{
		GetCCloudTokensFunc: func(_ pauth.CCloudClientFactory, _ string, _ *pauth.Credentials, _ bool, _ string) (string, string, error) {
			return testToken, "refreshToken", nil
		},
		GetConfluentTokenFunc: func(_ *mdsv1.APIClient, _ *pauth.Credentials) (string, error) {
			return testToken, nil
		},
	}
)

func TestLogout(t *testing.T) {
	req := require.New(t)
	clearCCloudDeprecatedEnvVar(req)
	cfg := config.AuthenticatedCloudConfigMock()
	contextName := cfg.Context().Name
	logoutCmd, cfg := newLogoutCmd(cfg)
	_, err := pcmd.ExecuteCommand(logoutCmd)
	req.NoError(err)
	verifyLoggedOutState(t, cfg, contextName)
}

func newLogoutCmd(cfg *config.Config) (*cobra.Command, *config.Config) {
	logoutCmd := New(cfg, climock.NewPreRunnerMock(nil, nil, nil, nil, cfg))
	return logoutCmd, cfg
}

func verifyLoggedOutState(t *testing.T, cfg *config.Config, loggedOutContext string) {
	req := require.New(t)
	state := cfg.Contexts[loggedOutContext].State
	req.Empty(state.AuthToken)
	req.Empty(state.Auth)
}

func clearCCloudDeprecatedEnvVar(req *require.Assertions) {
	req.NoError(os.Unsetenv(pauth.DeprecatedConfluentCloudEmail))
}
