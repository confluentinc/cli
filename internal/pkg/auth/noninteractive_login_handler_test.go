package auth
import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"
	"github.com/confluentinc/ccloud-sdk-go"
	"github.com/confluentinc/mds-sdk-go/mdsv1"

	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/confluentinc/cli/internal/pkg/mock"
	"github.com/confluentinc/cli/internal/pkg/netrc"
)

const (
	username = "chrissy"
	password = "password"
	deprecatedEnvUser = "deprecated-chrissy"
	deprecatedEnvPassword = "deprecated-password"
	refreshToken = "refresh-token"
	ssoAuthToken = "ccloud-sso-auth-token"
	ccloudCredentialsAuthToken = "ccloud-credentials-auth-token"
	confluentAuthToken = "confluent-auth-token"

	netrcFileName = ".netrc"
)

var (
	usernamePasswordCredentials = &Credentials{
		Username:     username,
		Password:     password,
	}
	deprecateEnvCredentials = &Credentials{
		Username:     deprecatedEnvUser,
		Password:     deprecatedEnvPassword,
	}
	ssoCredentials = &Credentials{
		Username:     username,
		RefreshToken: refreshToken,
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

type NonInteractiveLoginHandlerTestSuite struct {
	suite.Suite
	require *require.Assertions

	ccloudClient *ccloud.Client
	mdsClient    *mdsv1.APIClient
	logger       *log.Logger
	authTokenHandler AuthTokenHandler
	netrcHandler  netrc.NetrcHandler

	nonInteractiveLoginHandler NonInteractiveLoginHandler
}

func (suite *NonInteractiveLoginHandlerTestSuite) SetupSuite() {
	suite.ccloudClient = &ccloud.Client{}
	suite.mdsClient = &mdsv1.APIClient{}
	suite.logger = log.New()
	suite.authTokenHandler = &mock.MockAuthTokenHandler{
		GetCCloudUserSSOFunc: func(client *ccloud.Client, email string) (*orgv1.User,error) {
			return &orgv1.User{}, nil
		},
		GetCCloudCredentialsTokenFunc: func(client *ccloud.Client, email, password string) (string,error) {
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
}

func (suite *NonInteractiveLoginHandlerTestSuite) SetupTest() {
	suite.require = require.New(suite.T())
	suite.clearCCloudEnvironmentVariables()
	suite.clearConfluentEnvironmentVariables()
	suite.nonInteractiveLoginHandler = NewNonInteractiveLoginHandler(suite.authTokenHandler, suite.netrcHandler, suite.logger)
}

func (suite *NonInteractiveLoginHandlerTestSuite) TestGetCCloudTokenAndCredentialsFromEnvVar() {
	// incomplete credentials, setting on username but not password
	suite.require.NoError(os.Setenv(CCloudEmailEnvVar, username))
	token, creds, err := suite.nonInteractiveLoginHandler.GetCCloudTokenAndCredentialsFromEnvVar(suite.ccloudClient)
	suite.require.NoError(err)
	suite.require.Empty(token)
	suite.require.Nil(creds)

	suite.setCCloudEnvironmentVariables()
	token, creds, err = suite.nonInteractiveLoginHandler.GetCCloudTokenAndCredentialsFromEnvVar(suite.ccloudClient)
	suite.require.NoError(err)
	suite.require.Equal(ccloudCredentialsAuthToken, token)
	suite.compareCredentials(usernamePasswordCredentials, creds)
}

func (suite *NonInteractiveLoginHandlerTestSuite) TestGetCCloudTokenAndCredentialsFromDeprecatedEnvVar() {
	// incomplete credentials, setting on username but not password
	suite.require.NoError(os.Setenv(CCloudEmailDeprecatedEnvVar, username))
	token, creds, err := suite.nonInteractiveLoginHandler.GetCCloudTokenAndCredentialsFromEnvVar(suite.ccloudClient)
	suite.require.NoError(err)
	suite.require.Empty(token)
	suite.require.Nil(creds)

	suite.setCCloudDeprecatedEnvironmentVariables()
	token, creds, err = suite.nonInteractiveLoginHandler.GetCCloudTokenAndCredentialsFromEnvVar(suite.ccloudClient)
	suite.require.NoError(err)
	suite.require.Equal(ccloudCredentialsAuthToken, token)
	suite.compareCredentials(deprecateEnvCredentials, creds)
}

func (suite *NonInteractiveLoginHandlerTestSuite) TestGetCCloudTokenAndCredentialsFromEnvVarOrderOfPrecedence() {
	suite.setCCloudEnvironmentVariables()
	suite.setCCloudDeprecatedEnvironmentVariables()
	token, creds, err := suite.nonInteractiveLoginHandler.GetCCloudTokenAndCredentialsFromEnvVar(suite.ccloudClient)
	suite.require.NoError(err)
	suite.require.Equal(ccloudCredentialsAuthToken, token)
	suite.compareCredentials(usernamePasswordCredentials, creds)
}

func (suite *NonInteractiveLoginHandlerTestSuite) TestGetConfluentTokenAndCredentialsFromEnvVar() {
	// incomplete credentials, setting on username but not password
	suite.require.NoError(os.Setenv(ConfluentUsernameEnvVar, username))
	token, creds, err := suite.nonInteractiveLoginHandler.GetConfluentTokenAndCredentialsFromEnvVar(suite.mdsClient)
	suite.require.NoError(err)
	suite.require.Empty(token)
	suite.require.Nil(creds)

	suite.setConfluentEnvironmentVariables()
	token, creds, err = suite.nonInteractiveLoginHandler.GetConfluentTokenAndCredentialsFromEnvVar(suite.mdsClient)
	suite.require.NoError(err)
	suite.require.Equal(confluentAuthToken, token)
	suite.compareCredentials(usernamePasswordCredentials, creds)
}

func (suite *NonInteractiveLoginHandlerTestSuite) TestGetConfluentTokenAndCredentialsFromDeprecatedEnvVar() {
	// incomplete credentials, setting on username but not password
	suite.require.NoError(os.Setenv(ConfluentUsernameDeprecatedEnvVar, username))
	token, creds, err := suite.nonInteractiveLoginHandler.GetConfluentTokenAndCredentialsFromEnvVar(suite.mdsClient)
	suite.require.NoError(err)
	suite.require.Empty(token)
	suite.require.Nil(creds)

	suite.setConfluentDeprecatedEnvironmentVariables()
	token, creds, err = suite.nonInteractiveLoginHandler.GetConfluentTokenAndCredentialsFromEnvVar(suite.mdsClient)
	suite.require.NoError(err)
	suite.require.Equal(confluentAuthToken, token)
	suite.compareCredentials(deprecateEnvCredentials, creds)
}

func (suite *NonInteractiveLoginHandlerTestSuite) TestGetConfluentTokenAndCredentialsFromEnvVarOrderOfPrecedence() {
	suite.setConfluentEnvironmentVariables()
	suite.setConfluentDeprecatedEnvironmentVariables()
	token, creds, err := suite.nonInteractiveLoginHandler.GetConfluentTokenAndCredentialsFromEnvVar(suite.mdsClient)
	suite.require.NoError(err)
	suite.require.Equal(confluentAuthToken, token)
	suite.compareCredentials(usernamePasswordCredentials, creds)
}

func (suite *NonInteractiveLoginHandlerTestSuite) TestGetCCloudTokenAndCredentialsFromNetrcUsernamePassword() {
	token, creds, err := suite.nonInteractiveLoginHandler.GetCCloudTokenAndCredentialsFromNetrc(suite.ccloudClient, "", netrc.GetMatchingNetrcMachineParams{
		CLIName: "ccloud",
		IsSSO:   false,
	})
	suite.require.NoError(err)
	suite.require.Equal(ccloudCredentialsAuthToken, token)
	suite.compareCredentials(usernamePasswordCredentials, creds)
}

func (suite *NonInteractiveLoginHandlerTestSuite) TestGetCCloudTokenAndCredentialsFromNetrcSSO() {
	token, creds, err := suite.nonInteractiveLoginHandler.GetCCloudTokenAndCredentialsFromNetrc(suite.ccloudClient, "", netrc.GetMatchingNetrcMachineParams{
		CLIName: "ccloud",
		IsSSO:   true,
	})
	suite.require.NoError(err)
	suite.require.Equal(ssoAuthToken, token)
	suite.compareCredentials(ssoCredentials, creds)
}

func (suite *NonInteractiveLoginHandlerTestSuite) TestGetConfluentTokenAndCredentialsFromNetrc() {
	token, creds, err := suite.nonInteractiveLoginHandler.GetConfluentTokenAndCredentialsFromNetrc(suite.mdsClient, netrc.GetMatchingNetrcMachineParams{
		CLIName: "ccloud",
		IsSSO:   false,
	})
	suite.require.NoError(err)
	suite.require.Equal(confluentAuthToken, token)
	suite.compareCredentials(usernamePasswordCredentials, creds)
}

func (suite *NonInteractiveLoginHandlerTestSuite) compareCredentials(expect, actual *Credentials) {
	suite.require.Equal(expect.Username, actual.Username)
	suite.require.Equal(expect.Password, actual.Password)
	suite.require.Equal(expect.RefreshToken, actual.RefreshToken)
}

func (suite *NonInteractiveLoginHandlerTestSuite) clearCCloudEnvironmentVariables() {
	suite.require.NoError(os.Setenv(CCloudEmailEnvVar, ""))
	suite.require.NoError(os.Setenv(CCloudPasswordEnvVar, ""))
	suite.require.NoError(os.Setenv(CCloudEmailDeprecatedEnvVar, ""))
	suite.require.NoError(os.Setenv(CCloudPasswordDeprecatedEnvVar, ""))
}

func (suite *NonInteractiveLoginHandlerTestSuite) setCCloudEnvironmentVariables() {
	suite.require.NoError(os.Setenv(CCloudEmailEnvVar, username))
	suite.require.NoError(os.Setenv(CCloudPasswordEnvVar, password))
}

func (suite *NonInteractiveLoginHandlerTestSuite) setCCloudDeprecatedEnvironmentVariables() {
	suite.require.NoError(os.Setenv(CCloudEmailDeprecatedEnvVar, deprecatedEnvUser))
	suite.require.NoError(os.Setenv(CCloudPasswordDeprecatedEnvVar, deprecatedEnvPassword))
}

func (suite *NonInteractiveLoginHandlerTestSuite) setConfluentEnvironmentVariables() {
	suite.require.NoError(os.Setenv(ConfluentUsernameEnvVar, username))
	suite.require.NoError(os.Setenv(ConfluentPasswordEnvVar, password))
}

func (suite *NonInteractiveLoginHandlerTestSuite) setConfluentDeprecatedEnvironmentVariables() {
	suite.require.NoError(os.Setenv(ConfluentUsernameDeprecatedEnvVar, deprecatedEnvUser))
	suite.require.NoError(os.Setenv(ConfluentPasswordDeprecatedEnvVar, deprecatedEnvPassword))
}

func (suite *NonInteractiveLoginHandlerTestSuite) clearConfluentEnvironmentVariables() {
	suite.require.NoError(os.Setenv(ConfluentUsernameEnvVar, ""))
	suite.require.NoError(os.Setenv(ConfluentPasswordEnvVar, ""))
	suite.require.NoError(os.Setenv(ConfluentUsernameDeprecatedEnvVar, ""))
	suite.require.NoError(os.Setenv(ConfluentPasswordDeprecatedEnvVar, ""))
}

func TestNonInteractiveLoginHandler(t *testing.T) {
	suite.Run(t, new(NonInteractiveLoginHandlerTestSuite))
}
