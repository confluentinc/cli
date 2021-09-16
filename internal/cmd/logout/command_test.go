package logout

import (
	"context"
	"net/http"
	"os"
	"testing"

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
	"github.com/confluentinc/cli/internal/pkg/log"
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
	mockLoginCredentialsManager = &cliMock.MockLoginCredentialsManager{
		GetCCloudCredentialsFromEnvVarFunc: func(cmd *cobra.Command) func() (*pauth.Credentials, error) {
			return func() (*pauth.Credentials, error) {
				return nil, nil
			}
		},
		GetCCloudCredentialsFromPromptFunc: func(cmd *cobra.Command) func() (*pauth.Credentials, error) {
			return func() (*pauth.Credentials, error) {
				return &pauth.Credentials{
					Username: promptUser,
					Password: promptPassword,
				}, nil
			}
		},

		GetConfluentCredentialsFromEnvVarFunc: func(cmd *cobra.Command) func() (*pauth.Credentials, error) {
			return func() (*pauth.Credentials, error) {
				return nil, nil
			}
		},
		GetConfluentCredentialsFromPromptFunc: func(cmd *cobra.Command) func() (*pauth.Credentials, error) {
			return func() (*pauth.Credentials, error) {
				return &pauth.Credentials{
					Username: promptUser,
					Password: promptPassword,
				}, nil
			}
		},

		GetCredentialsFromNetrcFunc: func(cmd *cobra.Command, filterParams netrc.NetrcMachineParams) func() (*pauth.Credentials, error) {
			return func() (*pauth.Credentials, error) {
				return nil, nil
			}
		},
	}
	mockAuthTokenHandler = &cliMock.MockAuthTokenHandler{
		GetCCloudTokensFunc: func(client *ccloud.Client, credentials *pauth.Credentials, noBrowser bool) (s string, s2 string, e error) {
			return testToken, "refreshToken", nil
		},
		GetConfluentTokenFunc: func(mdsClient *mds.APIClient, credentials *pauth.Credentials) (s string, e error) {
			return testToken, nil
		},
	}
	mockNetrcHandler = &pmock.MockNetrcHandler{
		GetFileNameFunc: func() string { return netrcFile },
		WriteNetrcCredentialsFunc: func(isCloud, isSSO bool, ctxName, username, password string) error {
			return nil
		},
		RemoveNetrcCredentialsFunc: func(isCloud bool, ctxName string) (string, error) {
			return "", nil
		},
		CheckCredentialExistFunc: func(isCloud bool, ctxName string) (bool, error) {
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
	contextName := cfg.Context().Name
	logoutCmd, _ := newLogoutCmd(cfg, mockNetrcHandler)
	// run login command
	auth := &sdkMock.Auth{
		LoginFunc: func(ctx context.Context, idToken string, username string, password string) (string, error) {
			return testToken, nil
		},
		UserFunc: func(ctx context.Context) (*orgv1.GetUserReply, error) {
			return &orgv1.GetUserReply{
				User: &orgv1.User{
					Id:        23,
					Email:     promptUser,
					FirstName: "Cody",
				},
				Accounts: []*orgv1.Account{{Id: "a-595", Name: "Default"}},
			}, nil
		},
	}
	user := &sdkMock.User{}
	loginCmd, _ := newLoginCmd(auth, user, true, req, mockNetrcHandler, mockAuthTokenHandler, mockLoginCredentialsManager)
	_, err := pcmd.ExecuteCommand(loginCmd.Command)
	req.NoError(err)

	_, err = logoutCmd.netrcHandler.RemoveNetrcCredentials(true, contextName)
	req.NoError(err)
	exist, err := mockNetrcHandler.CheckCredentialExistFunc(true, contextName)
	req.NoError(err)
	req.Equal(exist, false)
}

func newLoginCmd(auth *sdkMock.Auth, user *sdkMock.User, isCloud bool, req *require.Assertions, netrcHandler netrc.NetrcHandler,
	authTokenHandler pauth.AuthTokenHandler, loginCredentialsManager pauth.LoginCredentialsManager) (*login.Command, *v1.Config) {
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
			return &ccloud.Client{Auth: auth, User: user}
		},
		JwtHTTPClientFactoryFunc: func(ctx context.Context, jwt, baseURL string) *ccloud.Client {
			return &ccloud.Client{Auth: auth, User: user}
		},
	}
	mdsClientManager := &cliMock.MockMDSClientManager{
		GetMDSClientFunc: func(url string, caCertPath string, logger *log.Logger) (client *mds.APIClient, e error) {
			return mdsClient, nil
		},
	}
	prerunner := cliMock.NewPreRunnerMock(ccloudClientFactory.AnonHTTPClientFactory(ccloudURL), mdsClient, nil, cfg)
	loginCmd := login.New(prerunner, log.New(), ccloudClientFactory, mdsClientManager,
		cliMock.NewDummyAnalyticsMock(), netrcHandler, loginCredentialsManager, authTokenHandler, true)
	return loginCmd, cfg
}

func newLogoutCmd(cfg *v1.Config, netrcHandler netrc.NetrcHandler) (*Command, *v1.Config) {
	logoutCmd := New(cfg, cliMock.NewPreRunnerMock(nil, nil, nil, cfg), cliMock.NewDummyAnalyticsMock(), netrcHandler)
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
