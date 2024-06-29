package logout

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	ccloudv1 "github.com/confluentinc/ccloud-sdk-go-v1-public"
	ccloudv1mock "github.com/confluentinc/ccloud-sdk-go-v1-public/mock"
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
		GetConfluentTokenFunc: func(_ *mdsv1.APIClient, _ *pauth.Credentials, _ bool) (string, string, error) {
			return testToken, "", nil
		},
	}
)

func TestLogout(t *testing.T) {
	req := require.New(t)
	cfg := config.AuthenticatedConfigMockWithContextName(config.MockContextName)
	contextName := cfg.Context().Name
	logoutCmd, cfg := newLogoutCmd(getAuthMock(), nil, true, req, AuthTokenHandler, contextName)
	_, err := pcmd.ExecuteCommand(logoutCmd)
	req.NoError(err)
	verifyLoggedOutState(t, cfg, contextName)
}

func newLogoutCmd(auth *ccloudv1mock.Auth, userInterface *ccloudv1mock.UserInterface, isCloud bool, req *require.Assertions, authTokenHandler pauth.AuthTokenHandler, contextName string) (*cobra.Command, *config.Config) {
	config.SetTempHomeDir()
	cfg := config.AuthenticatedConfigMockWithContextName(contextName)
	var prerunner pcmd.PreRunner

	if !isCloud {
		mdsClient := climock.NewMdsClientMock(testToken)
		prerunner = climock.NewPreRunnerMock(nil, nil, mdsClient, nil, cfg)
	} else {
		ccloudClientFactory := climock.NewCCloudClientFactoryMock(auth, userInterface, req)
		prerunner = climock.NewPreRunnerMock(ccloudClientFactory.AnonHTTPClientFactory(ccloudURL), nil, nil, nil, cfg)
	}
	logoutCmd := New(cfg, prerunner, authTokenHandler)
	return logoutCmd, cfg
}

func verifyLoggedOutState(t *testing.T, cfg *config.Config, loggedOutContext string) {
	req := require.New(t)
	state := cfg.Contexts[loggedOutContext].State
	req.Empty(state.AuthToken)
	req.Empty(state.Auth)
}

func getAuthMock() *ccloudv1mock.Auth {
	return &ccloudv1mock.Auth{
		LoginFunc: func(_ *ccloudv1.AuthenticateRequest) (*ccloudv1.AuthenticateReply, error) {
			return &ccloudv1.AuthenticateReply{Token: testToken}, nil
		},
		UserFunc: func() (*ccloudv1.GetMeReply, error) {
			return &ccloudv1.GetMeReply{
				User: &ccloudv1.User{
					Id:        23,
					Email:     promptUser,
					FirstName: "Cody",
				},
				Organization: &ccloudv1.Organization{ResourceId: "o-123"},
				Accounts:     []*ccloudv1.Account{{Id: "a-595", Name: "Default"}},
			}, nil
		},
	}
}
