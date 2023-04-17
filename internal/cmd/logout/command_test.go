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
	mds "github.com/confluentinc/mds-sdk-go-public/mdsv1"
	mdsMock "github.com/confluentinc/mds-sdk-go-public/mdsv1/mock"

	"github.com/confluentinc/cli/internal/cmd/login"
	pauth "github.com/confluentinc/cli/internal/pkg/auth"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	pmock "github.com/confluentinc/cli/internal/pkg/mock"
	"github.com/confluentinc/cli/internal/pkg/netrc"
	climock "github.com/confluentinc/cli/mock"
)

const (
	testToken      = "y0ur.jwt.T0kEn"
	promptUser     = "prompt-user@confluent.io"
	promptPassword = " prompt-password "
	netrcFile      = "netrc-file"
	ccloudURL      = "https://confluent.cloud"
)

var (
	mockLoginCredentialsManager = &climock.LoginCredentialsManager{
		GetCloudCredentialsFromEnvVarFunc: func(_ string) func() (*pauth.Credentials, error) {
			return func() (*pauth.Credentials, error) {
				return nil, nil
			}
		},
		GetCloudCredentialsFromPromptFunc: func(_ string) func() (*pauth.Credentials, error) {
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
		GetOnPremCredentialsFromPromptFunc: func() func() (*pauth.Credentials, error) {
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
		GetCredentialsFromNetrcFunc: func(_ netrc.NetrcMachineParams) func() (*pauth.Credentials, error) {
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
		SetCloudClientFunc: func(_ *ccloudv1.Client) {},
	}
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
	_, err := pcmd.ExecuteCommand(logoutCmd)
	req.NoError(err)
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
	userInterface := &ccloudv1Mock.UserInterface{}
	loginCmd, _ := newLoginCmd(auth, userInterface, true, req, mockNetrcHandler, AuthTokenHandler, mockLoginCredentialsManager, LoginOrganizationManager)
	_, err := pcmd.ExecuteCommand(loginCmd)
	req.NoError(err)

	_, err = mockNetrcHandler.RemoveNetrcCredentials(true, contextName)
	req.NoError(err)
	exist, err := mockNetrcHandler.CheckCredentialExistFunc(true, contextName)
	req.NoError(err)
	req.Equal(exist, false)
}

func newLoginCmd(auth *ccloudv1Mock.Auth, userInterface *ccloudv1Mock.UserInterface, isCloud bool, req *require.Assertions, netrcHandler netrc.NetrcHandler, authTokenHandler pauth.AuthTokenHandler, loginCredentialsManager pauth.LoginCredentialsManager, loginOrganizationManager pauth.LoginOrganizationManager) (*cobra.Command, *v1.Config) {
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
	ccloudClientFactory := &climock.CCloudClientFactory{
		AnonHTTPClientFactoryFunc: func(baseURL string) *ccloudv1.Client {
			req.Equal("https://confluent.cloud", baseURL)
			return &ccloudv1.Client{Params: &ccloudv1.Params{HttpClient: new(http.Client)}, Auth: auth, User: userInterface}
		},
		JwtHTTPClientFactoryFunc: func(ctx context.Context, jwt, baseURL string) *ccloudv1.Client {
			return &ccloudv1.Client{Growth: &ccloudv1Mock.Growth{
				GetFreeTrialInfoFunc: func(_ context.Context, orgId int32) ([]*ccloudv1.GrowthPromoCodeClaim, error) {
					var claims []*ccloudv1.GrowthPromoCodeClaim
					return claims, nil
				},
			}, Auth: auth, User: userInterface}
		},
	}
	mdsClientManager := &climock.MDSClientManager{
		GetMDSClientFunc: func(_, _ string, _ bool) (*mds.APIClient, error) {
			return mdsClient, nil
		},
	}
	prerunner := climock.NewPreRunnerMock(ccloudClientFactory.AnonHTTPClientFactory(ccloudURL), nil, mdsClient, nil, cfg)
	loginCmd := login.New(cfg, prerunner, ccloudClientFactory, mdsClientManager, netrcHandler, loginCredentialsManager, loginOrganizationManager, authTokenHandler)
	return loginCmd, cfg
}

func newLogoutCmd(cfg *v1.Config, netrcHandler netrc.NetrcHandler) (*cobra.Command, *v1.Config) {
	logoutCmd := New(cfg, climock.NewPreRunnerMock(nil, nil, nil, nil, cfg), netrcHandler)
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
