package logout

import (
	"context"
	"net/http"
	"os"
	"testing"

	flowv1 "github.com/confluentinc/cc-structs/kafka/flow/v1"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"
	"github.com/confluentinc/ccloud-sdk-go-v1"
	sdkMock "github.com/confluentinc/ccloud-sdk-go-v1/mock"
	mds "github.com/confluentinc/mds-sdk-go/mdsv1"
	mdsMock "github.com/confluentinc/mds-sdk-go/mdsv1/mock"

	"github.com/confluentinc/cli/internal/cmd/login"
	pauth "github.com/confluentinc/cli/internal/pkg/auth"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/config"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/errors"
	pmock "github.com/confluentinc/cli/internal/pkg/mock"
	"github.com/confluentinc/cli/internal/pkg/netrc"
	cliMock "github.com/confluentinc/cli/mock"
)

const (
	testToken      = "y0ur.jwt.T0kEn"
	promptUser     = "prompt-user@confluent.io"
	promptPassword = " prompt-password "
	netrcFile      = "netrc-file"
	ccloudURL      = "https://confluent.cloud"
)

var (
	mockLoginCredentialsManager = &cliMock.LoginCredentialsManager{
		GetCloudCredentialsFromEnvVarFunc: func(_ *cobra.Command, orgResourceId string) func() (*pauth.Credentials, error) {
			return func() (*pauth.Credentials, error) {
				return nil, nil
			}
		},
		GetCloudCredentialsFromPromptFunc: func(_ *cobra.Command, orgResourceId string) func() (*pauth.Credentials, error) {
			return func() (*pauth.Credentials, error) {
				return &pauth.Credentials{
					Username: promptUser,
					Password: promptPassword,
				}, nil
			}
		},
		GetOnPremCredentialsFromEnvVarFunc: func(_ *cobra.Command) func() (*pauth.Credentials, error) {
			return func() (*pauth.Credentials, error) {
				return nil, nil
			}
		},
		GetOnPremCredentialsFromPromptFunc: func(_ *cobra.Command) func() (*pauth.Credentials, error) {
			return func() (*pauth.Credentials, error) {
				return &pauth.Credentials{
					Username: promptUser,
					Password: promptPassword,
				}, nil
			}
		},
		GetSsoCredentialsFromConfigFunc: func(_ *v1.Config) func() (*pauth.Credentials, error) {
			return func() (*pauth.Credentials, error) {
				return nil, nil
			}
		},
		GetCredentialsFromNetrcFunc: func(_ *cobra.Command, _ netrc.NetrcMachineParams) func() (*pauth.Credentials, error) {
			return func() (*pauth.Credentials, error) {
				return nil, nil
			}
		},
		GetCredentialsFromKeychainFunc: func(_ *v1.Config, _ bool, _, _ string) func() (*pauth.Credentials, error) {
			return func() (*pauth.Credentials, error) {
				return nil, nil
			}
		},
		GetCredentialsFromConfigFunc: func(_ *v1.Config, _ netrc.NetrcMachineParams) func() (*pauth.Credentials, error) {
			return func() (*pauth.Credentials, error) {
				return nil, nil
			}
		},
		SetCloudClientFunc: func(_ *ccloud.Client) {},
	}
	orgManagerImpl               = pauth.NewLoginOrganizationManagerImpl()
	mockLoginOrganizationManager = &cliMock.MockLoginOrganizationManager{
		GetLoginOrganizationFromArgsFunc: func(cmd *cobra.Command) func() (string, error) {
			return orgManagerImpl.GetLoginOrganizationFromArgs(cmd)
		},
		GetLoginOrganizationFromEnvVarFunc: func(cmd *cobra.Command) func() (string, error) {
			return orgManagerImpl.GetLoginOrganizationFromEnvVar(cmd)
		},
		GetDefaultLoginOrganizationFunc: func() func() (string, error) {
			return orgManagerImpl.GetDefaultLoginOrganization()
		},
	}
	mockAuthTokenHandler = &cliMock.MockAuthTokenHandler{
		GetCCloudTokensFunc: func(_ pauth.CCloudClientFactory, _ string, _ *pauth.Credentials, _ bool, _ string) (string, string, error) {
			return testToken, "refreshToken", nil
		},
		GetConfluentTokenFunc: func(_ *mds.APIClient, _ *pauth.Credentials) (string, error) {
			return testToken, nil
		},
	}
	mockNetrcHandler = &pmock.NetrcHandler{
		GetFileNameFunc: func() string { return netrcFile },
		RemoveNetrcCredentialsFunc: func(_ bool, _ string) (string, error) {
			return "", nil
		},
		CheckCredentialExistFunc: func(_ bool, _ string) (bool, error) {
			return false, nil
		},
	}
)

func TestLogout(t *testing.T) {
	req := require.New(t)
	clearCCloudDeprecatedEnvVar(req)
	cfg := v1.AuthenticatedCloudConfigMock()
	contextName := cfg.Context().Name
	logoutCmd, cfg := newLogoutCmd(cfg, mockNetrcHandler)
	output, err := pcmd.ExecuteCommand(logoutCmd.Command)
	req.NoError(err)
	req.Contains(output, errors.LoggedOutMsg)
	exist, err := mockNetrcHandler.CheckCredentialExistFunc(true, contextName)
	req.NoError(err)
	req.Equal(exist, false)
	verifyLoggedOutState(t, cfg, contextName)
}

func TestRemoveNetrcCredentials(t *testing.T) {
	req := require.New(t)
	clearCCloudDeprecatedEnvVar(req)
	cfg := v1.AuthenticatedCloudConfigMock()
	contextName := cfg.Context().GetNetrcMachineName()
	logoutCmd, _ := newLogoutCmd(cfg, mockNetrcHandler)
	// run login command
	auth := &sdkMock.Auth{
		LoginFunc: func(_ context.Context, _, _, _, _ string) (string, error) {
			return testToken, nil
		},
		UserFunc: func(_ context.Context) (*flowv1.GetMeReply, error) {
			return &flowv1.GetMeReply{
				User: &orgv1.User{
					Id:        23,
					Email:     promptUser,
					FirstName: "Cody",
				},
				Organization: &orgv1.Organization{ResourceId: "o-123"},
				Accounts:     []*orgv1.Account{{Id: "a-595", Name: "Default"}},
			}, nil
		},
	}
	user := &sdkMock.User{}
	loginCmd, _ := newLoginCmd(auth, user, true, req, mockNetrcHandler, mockAuthTokenHandler, mockLoginCredentialsManager, mockLoginOrganizationManager)
	_, err := pcmd.ExecuteCommand(loginCmd.Command)
	req.NoError(err)

	_, err = logoutCmd.netrcHandler.RemoveNetrcCredentials(true, contextName)
	req.NoError(err)
	exist, err := mockNetrcHandler.CheckCredentialExistFunc(true, contextName)
	req.NoError(err)
	req.Equal(exist, false)
}

func newLoginCmd(auth *sdkMock.Auth, user *sdkMock.User, isCloud bool, req *require.Assertions, netrcHandler netrc.NetrcHandler,
	authTokenHandler pauth.AuthTokenHandler, loginCredentialsManager pauth.LoginCredentialsManager,
	loginOrganizationManager pauth.LoginOrganizationManager) (*login.Command, *v1.Config) {
	cfg := v1.New(new(config.Params))
	var mdsClient *mds.APIClient
	if !isCloud {
		mdsConfig := mds.NewConfiguration()
		mdsClient = mds.NewAPIClient(mdsConfig)
		mdsClient.TokensAndAuthenticationApi = &mdsMock.TokensAndAuthenticationApi{
			GetTokenFunc: func(ctx context.Context) (mds.AuthenticationResponse, *http.Response, error) {
				return mds.AuthenticationResponse{
					AuthToken: testToken,
					TokenType: "JWT",
					ExpiresIn: 100,
				}, nil, nil
			},
		}
	}
	ccloudClientFactory := &cliMock.MockCCloudClientFactory{
		AnonHTTPClientFactoryFunc: func(baseURL string) *ccloud.Client {
			req.Equal("https://confluent.cloud", baseURL)
			return &ccloud.Client{Params: &ccloud.Params{HttpClient: new(http.Client)}, Auth: auth, User: user}
		},
		JwtHTTPClientFactoryFunc: func(ctx context.Context, jwt, baseURL string) *ccloud.Client {
			return &ccloud.Client{Auth: auth, User: user}
		},
	}
	mdsClientManager := &cliMock.MockMDSClientManager{
		GetMDSClientFunc: func(url string, caCertPath string) (client *mds.APIClient, e error) {
			return mdsClient, nil
		},
	}
	prerunner := cliMock.NewPreRunnerMock(ccloudClientFactory.AnonHTTPClientFactory(ccloudURL), mdsClient, nil, cfg)
	loginCmd := login.New(prerunner, ccloudClientFactory, mdsClientManager, netrcHandler, loginCredentialsManager, authTokenHandler, true)
	return loginCmd, cfg
}

func newLogoutCmd(cfg *v1.Config, netrcHandler netrc.NetrcHandler) (*Command, *v1.Config) {
	logoutCmd := New(cfg, cliMock.NewPreRunnerMock(nil, nil, nil, cfg), netrcHandler)
	return logoutCmd, cfg
}

func verifyLoggedOutState(t *testing.T, cfg *v1.Config, loggedOutContext string) {
	req := require.New(t)
	state := cfg.Contexts[loggedOutContext].State
	req.Empty(state.AuthToken)
	req.Empty(state.Auth)
}

func clearCCloudDeprecatedEnvVar(req *require.Assertions) {
	req.NoError(os.Unsetenv(pauth.DeprecatedConfluentCloudEmail))
}
