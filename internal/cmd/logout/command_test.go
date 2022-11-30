package logout

import (
	"context"
	"net/http"
	"os"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	ccloudv1 "github.com/confluentinc/ccloud-sdk-go-v1-public"
	ccloudv1Mock "github.com/confluentinc/ccloud-sdk-go-v1-public/mock"
	mds "github.com/confluentinc/mds-sdk-go/mdsv1"
	mdsMock "github.com/confluentinc/mds-sdk-go/mdsv1/mock"

	"github.com/confluentinc/cli/internal/cmd/login"
	pauth "github.com/confluentinc/cli/internal/pkg/auth"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
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
		GetCloudCredentialsFromEnvVarFunc: func(orgResourceId string) func() (*pauth.Credentials, error) {
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
		GetOnPremCredentialsFromEnvVarFunc: func() func() (*pauth.Credentials, error) {
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
		GetCredentialsFromConfigFunc: func(_ *v1.Config) func() (*pauth.Credentials, error) {
			return func() (*pauth.Credentials, error) {
				return nil, nil
			}
		},
		GetCredentialsFromNetrcFunc: func(_ netrc.NetrcMachineParams) func() (*pauth.Credentials, error) {
			return func() (*pauth.Credentials, error) {
				return nil, nil
			}
		},
		SetCloudClientFunc: func(_ *ccloudv1.Client) {},
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
	mockNetrcHandler = &pmock.MockNetrcHandler{
		GetFileNameFunc: func() string { return netrcFile },
		WriteNetrcCredentialsFunc: func(_ bool, _, _, _ string) error {
			return nil
		},
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
	output, err := pcmd.ExecuteCommand(logoutCmd)
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
	contextName := cfg.Context().NetrcMachineName
	// run login command
	auth := &ccloudv1Mock.Auth{
		LoginFunc: func(_ context.Context, _ *ccloudv1.AuthenticateRequest) (*ccloudv1.AuthenticateReply, error) {
			return &ccloudv1.AuthenticateReply{Token: testToken}, nil
		},
		UserFunc: func(_ context.Context) (*ccloudv1.GetMeReply, error) {
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
	loginRealm := &ccloudv1Mock.LoginRealm{}
	loginCmd, _ := newLoginCmd(auth, loginRealm, true, req, mockNetrcHandler, mockAuthTokenHandler, mockLoginCredentialsManager, mockLoginOrganizationManager)
	_, err := pcmd.ExecuteCommand(loginCmd)
	req.NoError(err)

	_, err = mockNetrcHandler.RemoveNetrcCredentials(true, contextName)
	req.NoError(err)
	exist, err := mockNetrcHandler.CheckCredentialExistFunc(true, contextName)
	req.NoError(err)
	req.Equal(exist, false)
}

func newLoginCmd(auth *ccloudv1Mock.Auth, loginRealm *ccloudv1Mock.LoginRealm, isCloud bool, req *require.Assertions, netrcHandler netrc.NetrcHandler,
	authTokenHandler pauth.AuthTokenHandler, loginCredentialsManager pauth.LoginCredentialsManager,
	loginOrganizationManager pauth.LoginOrganizationManager) (*cobra.Command, *v1.Config) {
	cfg := v1.New()
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
	ccloudClientFactory := &cliMock.CCloudClientFactory{
		AnonHTTPClientFactoryFunc: func(baseURL string) *ccloudv1.Client {
			req.Equal("https://confluent.cloud", baseURL)
			return &ccloudv1.Client{Params: &ccloudv1.Params{HttpClient: new(http.Client)}, Auth: auth, LoginRealm: loginRealm}
		},
		JwtHTTPClientFactoryFunc: func(ctx context.Context, jwt, baseURL string) *ccloudv1.Client {
			return &ccloudv1.Client{Billing: &ccloudv1Mock.Billing{
				GetClaimedPromoCodesFunc: func(_ context.Context, _ *ccloudv1.Organization, _ bool) ([]*ccloudv1.PromoCodeClaim, error) {
					var claims []*ccloudv1.PromoCodeClaim
					return claims, nil
				},
			}, Auth: auth, LoginRealm: loginRealm}
		},
	}
	mdsClientManager := &cliMock.MockMDSClientManager{
		GetMDSClientFunc: func(_, _ string, _ bool) (*mds.APIClient, error) {
			return mdsClient, nil
		},
	}
	prerunner := cliMock.NewPreRunnerMock(nil, ccloudClientFactory.AnonHTTPClientFactory(ccloudURL), nil, mdsClient, nil, cfg)
	loginCmd := login.New(cfg, prerunner, ccloudClientFactory, mdsClientManager, netrcHandler, loginCredentialsManager, loginOrganizationManager, authTokenHandler)
	return loginCmd, cfg
}

func newLogoutCmd(cfg *v1.Config, netrcHandler netrc.NetrcHandler) (*cobra.Command, *v1.Config) {
	logoutCmd := New(cfg, cliMock.NewPreRunnerMock(nil, nil, nil, nil, nil, cfg), netrcHandler)
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
