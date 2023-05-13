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
	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
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
	LoginCredentialsManager = &cliMock.LoginCredentialsManager{
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

		GetCredentialsFromNetrcFunc: func(cmd *cobra.Command, filterParams netrc.GetMatchingNetrcMachineParams) func() (*pauth.Credentials, error) {
			return func() (*pauth.Credentials, error) {
				return nil, nil
			}
		},

		GetCredentialsFromConfigFunc: func(_ *v3.Config, _ netrc.GetMatchingNetrcMachineParams) func() (*pauth.Credentials, error) {
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
	NetrcHandler = &pmock.NetrcHandler{
		GetFileNameFunc: func() string { return netrcFile },
		RemoveNetrcCredentialsFunc: func(cliName string, ctxName string) (string, error) {
			return "", nil
		},
		CheckCredentialExistFunc: func(cliName string, ctxName string) (bool, error) {
			return false, nil
		},
	}
)

func TestLogout(t *testing.T) {
	req := require.New(t)
	clearCCloudDeprecatedEnvVar(req)
	cfg := v3.AuthenticatedCloudConfigMock()
	contextName := cfg.Context().Name
	logoutCmd, cfg := newLogoutCmd("ccloud", cfg, NetrcHandler)
	output, err := pcmd.ExecuteCommand(logoutCmd.Command)
	req.NoError(err)
	req.Contains(output, errors.LoggedOutMsg)
	exist, err := NetrcHandler.CheckCredentialExistFunc("ccloud", contextName)
	req.NoError(err)
	req.Equal(exist, false)
	verifyLoggedOutState(t, cfg, contextName)
}

func TestRemoveNetrcCredentials(t *testing.T) {
	req := require.New(t)
	clearCCloudDeprecatedEnvVar(req)
	cfg := v3.AuthenticatedCloudConfigMock()
	contextName := cfg.Context().Name
	logoutCmd, _ := newLogoutCmd("ccloud", cfg, NetrcHandler)
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
				Accounts: []*orgv1.Account{{Id: "a-595", Name: "Default"}},
			}, nil
		},
	}
	user := &sdkMock.User{}
	suite := []struct {
		cliName string
		args    []string
		setEnv  bool
	}{
		{
			cliName: "ccloud",
			args:    []string{},
		},
	}
	loginCmd, _ := newLoginCmd(auth, user, suite[0].cliName, req, NetrcHandler, mockAuthTokenHandler, LoginCredentialsManager)
	_, err := pcmd.ExecuteCommand(loginCmd.Command, suite[0].args...)
	req.NoError(err)

	_, err = logoutCmd.netrcHandler.RemoveNetrcCredentials(logoutCmd.cliName, contextName)
	req.NoError(err)
	exist, err := NetrcHandler.CheckCredentialExistFunc("ccloud", contextName)
	req.NoError(err)
	req.Equal(exist, false)
}

func newLoginCmd(auth *sdkMock.Auth, user *sdkMock.User, cliName string, req *require.Assertions, netrcHandler netrc.NetrcHandler,
	authTokenHandler pauth.AuthTokenHandler, loginCredentialsManager pauth.LoginCredentialsManager) (*login.Command, *v3.Config) {
	cfg := v3.New(&config.Params{
		CLIName:    cliName,
		MetricSink: nil,
		Logger:     nil,
	})
	var mdsClient *mds.APIClient
	if cliName == "confluent" {
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
		GetMDSClientFunc: func(url string, caCertPath string, logger *log.Logger) (client *mds.APIClient, e error) {
			return mdsClient, nil
		},
	}
	prerunner := cliMock.NewPreRunnerMock(ccloudClientFactory.AnonHTTPClientFactory(ccloudURL), mdsClient, nil, cfg)
	loginCmd := login.New(cliName, cfg, prerunner, log.New(), ccloudClientFactory, mdsClientManager,
		cliMock.NewDummyAnalyticsMock(), netrcHandler, loginCredentialsManager, authTokenHandler)
	return loginCmd, cfg
}

func newLogoutCmd(cliName string, cfg *v3.Config, netrcHandler netrc.NetrcHandler) (*Command, *v3.Config) {
	logoutCmd := New(cliName, cliMock.NewPreRunnerMock(nil, nil, nil, cfg), cliMock.NewDummyAnalyticsMock(), netrcHandler)
	return logoutCmd, cfg
}

func verifyLoggedOutState(t *testing.T, cfg *v3.Config, loggedOutContext string) {
	req := require.New(t)
	state := cfg.Contexts[loggedOutContext].State
	req.Empty(state.AuthToken)
	req.Empty(state.Auth)
}

// XX_CCLOUD_EMAIL is used for integration test hack
func clearCCloudDeprecatedEnvVar(req *require.Assertions) {
	req.NoError(os.Setenv(pauth.CCloudEmailDeprecatedEnvVar, ""))
}
