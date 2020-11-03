package auth

import (
	"context"
	"os"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"
	"github.com/confluentinc/ccloud-sdk-go"
	sdkMock "github.com/confluentinc/ccloud-sdk-go/mock"
	"github.com/confluentinc/mds-sdk-go/mdsv1"

	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/confluentinc/cli/internal/pkg/mock"
	"github.com/confluentinc/cli/internal/pkg/netrc"
)

const (
	username                   = "chrissy"
	password                   = "password"
	deprecatedEnvUser          = "deprecated-chrissy"
	deprecatedEnvPassword      = "deprecated-password"
	refreshToken               = "refresh-token"
	ssoAuthToken               = "ccloud-sso-auth-token"
	promptUsername             = "prompt-chrissy"
	promptPassword             = "  prompt-password  "
	ccloudCredentialsAuthToken = "ccloud-credentials-auth-token"
	confluentAuthToken         = "confluent-auth-token"

	netrcFileName = ".netrc"
)

var (
	usernamePasswordCredentials = &Credentials{
		Username: username,
		Password: password,
	}
	deprecateEnvCredentials = &Credentials{
		Username: deprecatedEnvUser,
		Password: deprecatedEnvPassword,
	}
	ssoCredentials = &Credentials{
		Username:     username,
		RefreshToken: refreshToken,
	}
	promptCredentials = &Credentials{
		Username: promptUsername,
		Password: promptPassword,
	}

	ccloudCredMachine = &netrc.Machine{
		Name:     "ccloud-cred",
		User:     username,
		Password: password,
		IsSSO:    false,
	}
	ccloudSSOMachine = &netrc.Machine{
		Name:     "ccloud-sso",
		User:     username,
		Password: refreshToken,
		IsSSO:    true,
	}
	confluentMachine = &netrc.Machine{
		Name:     "confluent",
		User:     username,
		Password: confluentAuthToken,
		IsSSO:    false,
	}
)

type LoginTokenHandlerTestSuite struct {
	suite.Suite
	require *require.Assertions

	ccloudClient     *ccloud.Client
	mdsClient        *mdsv1.APIClient
	logger           *log.Logger
	authTokenHandler AuthTokenHandler
	netrcHandler     netrc.NetrcHandler
	prompt           mock.Prompt

	loginTokenHandler LoginTokenHandler
}

func (suite *LoginTokenHandlerTestSuite) SetupSuite() {
	suite.ccloudClient = &ccloud.Client{
		User: &sdkMock.User{
			CheckEmailFunc: func(ctx context.Context, user *orgv1.User) (*orgv1.User, error) {
				return &orgv1.User{
					Email: "",
				}, nil
			},
		},
	}
	suite.mdsClient = &mdsv1.APIClient{}
	suite.logger = log.New()
	suite.authTokenHandler = &mock.MockAuthTokenHandler{
		GetCCloudUserSSOFunc: func(client *ccloud.Client, email string) (*orgv1.User, error) {
			return &orgv1.User{}, nil
		},
		GetCCloudCredentialsTokenFunc: func(client *ccloud.Client, email, password string) (string, error) {
			return ccloudCredentialsAuthToken, nil
		},
		GetCCloudSSOTokenFunc: func(client *ccloud.Client, url string, noBrowser bool, email string, logger *log.Logger) (string, string, error) {
			return ssoAuthToken, refreshToken, nil
		},
		RefreshCCloudSSOTokenFunc: func(client *ccloud.Client, refreshToken, url string, logger *log.Logger) (string, error) {
			return ssoAuthToken, nil
		},
		GetConfluentAuthTokenFunc: func(mdsClient *mdsv1.APIClient, username, password string) (string, error) {
			return confluentAuthToken, nil
		},
	}
	suite.netrcHandler = &mock.MockNetrcHandler{
		GetMatchingNetrcMachineFunc: func(params netrc.GetMatchingNetrcMachineParams) (*netrc.Machine, error) {
			if params.CLIName == "ccloud" {
				if params.IsSSO {
					return ccloudSSOMachine, nil
				}
				return ccloudCredMachine, nil
			} else {
				return confluentMachine, nil
			}
		},
		GetFileNameFunc: func() string {
			return netrcFileName
		},
	}
	suite.prompt = mock.Prompt{
		ReadLineFunc: func() (string, error) {
			return promptUsername, nil
		},
		ReadLineMaskedFunc: func() (string, error) {
			return promptPassword, nil
		},
	}
}

func (suite *LoginTokenHandlerTestSuite) SetupTest() {
	suite.require = require.New(suite.T())
	suite.clearCCloudEnvironmentVariables()
	suite.clearConfluentEnvironmentVariables()
	suite.loginTokenHandler = NewLoginTokenHandler(suite.authTokenHandler, suite.netrcHandler, &suite.prompt, suite.logger)
}

func (suite *LoginTokenHandlerTestSuite) TestGetCCloudTokenAndCredentialsFromEnvVar() {
	// incomplete credentials, setting on username but not password
	suite.require.NoError(os.Setenv(CCloudEmailEnvVar, username))
	token, creds, err := suite.loginTokenHandler.GetCCloudTokenAndCredentialsFromEnvVar(&cobra.Command{}, suite.ccloudClient)
	suite.require.NoError(err)
	suite.require.Empty(token)
	suite.require.Nil(creds)

	suite.setCCloudEnvironmentVariables()
	token, creds, err = suite.loginTokenHandler.GetCCloudTokenAndCredentialsFromEnvVar(&cobra.Command{}, suite.ccloudClient)
	suite.require.NoError(err)
	suite.require.Equal(ccloudCredentialsAuthToken, token)
	suite.compareCredentials(usernamePasswordCredentials, creds)
}

func (suite *LoginTokenHandlerTestSuite) TestGetCCloudTokenAndCredentialsFromDeprecatedEnvVar() {
	// incomplete credentials, setting on username but not password
	suite.require.NoError(os.Setenv(CCloudEmailDeprecatedEnvVar, username))
	token, creds, err := suite.loginTokenHandler.GetCCloudTokenAndCredentialsFromEnvVar(&cobra.Command{}, suite.ccloudClient)
	suite.require.NoError(err)
	suite.require.Empty(token)
	suite.require.Nil(creds)

	suite.setCCloudDeprecatedEnvironmentVariables()
	token, creds, err = suite.loginTokenHandler.GetCCloudTokenAndCredentialsFromEnvVar(&cobra.Command{}, suite.ccloudClient)
	suite.require.NoError(err)
	suite.require.Equal(ccloudCredentialsAuthToken, token)
	suite.compareCredentials(deprecateEnvCredentials, creds)
}

func (suite *LoginTokenHandlerTestSuite) TestGetCCloudTokenAndCredentialsFromEnvVarOrderOfPrecedence() {
	suite.setCCloudEnvironmentVariables()
	suite.setCCloudDeprecatedEnvironmentVariables()
	token, creds, err := suite.loginTokenHandler.GetCCloudTokenAndCredentialsFromEnvVar(&cobra.Command{}, suite.ccloudClient)
	suite.require.NoError(err)
	suite.require.Equal(ccloudCredentialsAuthToken, token)
	suite.compareCredentials(usernamePasswordCredentials, creds)
}

func (suite *LoginTokenHandlerTestSuite) TestGetConfluentTokenAndCredentialsFromEnvVar() {
	// incomplete credentials, setting on username but not password
	suite.require.NoError(os.Setenv(ConfluentUsernameEnvVar, username))
	token, creds, err := suite.loginTokenHandler.GetConfluentTokenAndCredentialsFromEnvVar(&cobra.Command{}, suite.mdsClient)
	suite.require.NoError(err)
	suite.require.Empty(token)
	suite.require.Nil(creds)

	suite.setConfluentEnvironmentVariables()
	token, creds, err = suite.loginTokenHandler.GetConfluentTokenAndCredentialsFromEnvVar(&cobra.Command{}, suite.mdsClient)
	suite.require.NoError(err)
	suite.require.Equal(confluentAuthToken, token)
	suite.compareCredentials(usernamePasswordCredentials, creds)
}

func (suite *LoginTokenHandlerTestSuite) TestGetConfluentTokenAndCredentialsFromDeprecatedEnvVar() {
	// incomplete credentials, setting on username but not password
	suite.require.NoError(os.Setenv(ConfluentUsernameDeprecatedEnvVar, username))
	token, creds, err := suite.loginTokenHandler.GetConfluentTokenAndCredentialsFromEnvVar(&cobra.Command{}, suite.mdsClient)
	suite.require.NoError(err)
	suite.require.Empty(token)
	suite.require.Nil(creds)

	suite.setConfluentDeprecatedEnvironmentVariables()
	token, creds, err = suite.loginTokenHandler.GetConfluentTokenAndCredentialsFromEnvVar(&cobra.Command{}, suite.mdsClient)
	suite.require.NoError(err)
	suite.require.Equal(confluentAuthToken, token)
	suite.compareCredentials(deprecateEnvCredentials, creds)
}

func (suite *LoginTokenHandlerTestSuite) TestGetConfluentTokenAndCredentialsFromEnvVarOrderOfPrecedence() {
	suite.setConfluentEnvironmentVariables()
	suite.setConfluentDeprecatedEnvironmentVariables()
	token, creds, err := suite.loginTokenHandler.GetConfluentTokenAndCredentialsFromEnvVar(&cobra.Command{}, suite.mdsClient)
	suite.require.NoError(err)
	suite.require.Equal(confluentAuthToken, token)
	suite.compareCredentials(usernamePasswordCredentials, creds)
}

func (suite *LoginTokenHandlerTestSuite) TestGetCCloudTokenAndCredentialsFromNetrcUsernamePassword() {
	token, creds, err := suite.loginTokenHandler.GetCCloudTokenAndCredentialsFromNetrc(&cobra.Command{}, suite.ccloudClient, "", netrc.GetMatchingNetrcMachineParams{
		CLIName: "ccloud",
		IsSSO:   false,
	})
	suite.require.NoError(err)
	suite.require.Equal(ccloudCredentialsAuthToken, token)
	suite.compareCredentials(usernamePasswordCredentials, creds)
}

func (suite *LoginTokenHandlerTestSuite) TestGetCCloudTokenAndCredentialsFromNetrcSSO() {
	token, creds, err := suite.loginTokenHandler.GetCCloudTokenAndCredentialsFromNetrc(&cobra.Command{}, suite.ccloudClient, "", netrc.GetMatchingNetrcMachineParams{
		CLIName: "ccloud",
		IsSSO:   true,
	})
	suite.require.NoError(err)
	suite.require.Equal(ssoAuthToken, token)
	suite.compareCredentials(ssoCredentials, creds)
}

func (suite *LoginTokenHandlerTestSuite) TestGetConfluentTokenAndCredentialsFromNetrc() {
	token, creds, err := suite.loginTokenHandler.GetConfluentTokenAndCredentialsFromNetrc(&cobra.Command{}, suite.mdsClient, netrc.GetMatchingNetrcMachineParams{
		CLIName: "ccloud",
		IsSSO:   false,
	})
	suite.require.NoError(err)
	suite.require.Equal(confluentAuthToken, token)
	suite.compareCredentials(usernamePasswordCredentials, creds)
}

func (suite *LoginTokenHandlerTestSuite) TestGetCCloudTokenAndCredentialsFromPrompt() {
	token, creds, err := suite.loginTokenHandler.GetCCloudTokenAndCredentialsFromPrompt(&cobra.Command{}, suite.ccloudClient, "")
	suite.require.NoError(err)
	suite.require.Equal(ccloudCredentialsAuthToken, token)
	suite.compareCredentials(promptCredentials, creds)
}

func (suite *LoginTokenHandlerTestSuite) TestGetConfluentTokenAndCredentialsFromPrompt() {
	token, creds, err := suite.loginTokenHandler.GetConfluentTokenAndCredentialsFromPrompt(&cobra.Command{}, suite.mdsClient)
	suite.require.NoError(err)
	suite.require.Equal(confluentAuthToken, token)
	suite.compareCredentials(promptCredentials, creds)
}

func (suite *LoginTokenHandlerTestSuite) compareCredentials(expect, actual *Credentials) {
	suite.require.Equal(expect.Username, actual.Username)
	suite.require.Equal(expect.Password, actual.Password)
	suite.require.Equal(expect.RefreshToken, actual.RefreshToken)
}

func (suite *LoginTokenHandlerTestSuite) clearCCloudEnvironmentVariables() {
	suite.require.NoError(os.Setenv(CCloudEmailEnvVar, ""))
	suite.require.NoError(os.Setenv(CCloudPasswordEnvVar, ""))
	suite.require.NoError(os.Setenv(CCloudEmailDeprecatedEnvVar, ""))
	suite.require.NoError(os.Setenv(CCloudPasswordDeprecatedEnvVar, ""))
}

func (suite *LoginTokenHandlerTestSuite) setCCloudEnvironmentVariables() {
	suite.require.NoError(os.Setenv(CCloudEmailEnvVar, username))
	suite.require.NoError(os.Setenv(CCloudPasswordEnvVar, password))
}

func (suite *LoginTokenHandlerTestSuite) setCCloudDeprecatedEnvironmentVariables() {
	suite.require.NoError(os.Setenv(CCloudEmailDeprecatedEnvVar, deprecatedEnvUser))
	suite.require.NoError(os.Setenv(CCloudPasswordDeprecatedEnvVar, deprecatedEnvPassword))
}

func (suite *LoginTokenHandlerTestSuite) setConfluentEnvironmentVariables() {
	suite.require.NoError(os.Setenv(ConfluentUsernameEnvVar, username))
	suite.require.NoError(os.Setenv(ConfluentPasswordEnvVar, password))
}

func (suite *LoginTokenHandlerTestSuite) setConfluentDeprecatedEnvironmentVariables() {
	suite.require.NoError(os.Setenv(ConfluentUsernameDeprecatedEnvVar, deprecatedEnvUser))
	suite.require.NoError(os.Setenv(ConfluentPasswordDeprecatedEnvVar, deprecatedEnvPassword))
}

func (suite *LoginTokenHandlerTestSuite) clearConfluentEnvironmentVariables() {
	suite.require.NoError(os.Setenv(ConfluentUsernameEnvVar, ""))
	suite.require.NoError(os.Setenv(ConfluentPasswordEnvVar, ""))
	suite.require.NoError(os.Setenv(ConfluentUsernameDeprecatedEnvVar, ""))
	suite.require.NoError(os.Setenv(ConfluentPasswordDeprecatedEnvVar, ""))
}

func TestLoginTokenHandler(t *testing.T) {
	suite.Run(t, new(LoginTokenHandlerTestSuite))
}
