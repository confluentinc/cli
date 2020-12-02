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

	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/confluentinc/cli/internal/pkg/mock"
	"github.com/confluentinc/cli/internal/pkg/netrc"
)

const (
	deprecatedEnvUser     = "deprecated-chrissy"
	deprecatedEnvPassword = "deprecated-password"

	envUsername = "env-username"
	envPassword = "env-password"

	netrcUsername = "netrc-username"
	netrcPassword = "netrc-password"

	ssoUsername  = "sso-username"
	refreshToken = "refresh-token"

	promptUsername = "prompt-chrissy"
	promptPassword = "  prompt-password  "

	netrcFileName = ".netrc"
)

var (
	envCredentials = &Credentials{
		Username: envUsername,
		Password: envPassword,
	}
	deprecateEnvCredentials = &Credentials{
		Username: deprecatedEnvUser,
		Password: deprecatedEnvPassword,
	}
	netrcCredentials = &Credentials{
		Username: netrcUsername,
		Password: netrcPassword,
		IsSSO:    false,
	}
	ssoCredentials = &Credentials{
		Username: ssoUsername,
		Password: refreshToken,
		IsSSO:    true,
	}
	promptCredentials = &Credentials{
		Username: promptUsername,
		Password: promptPassword,
	}

	ccloudCredMachine = &netrc.Machine{
		Name:     "ccloud-cred",
		User:     netrcUsername,
		Password: netrcPassword,
		IsSSO:    false,
	}
	ccloudSSOMachine = &netrc.Machine{
		Name:     "ccloud-sso",
		User:     ssoUsername,
		Password: refreshToken,
		IsSSO:    true,
	}
	confluentMachine = &netrc.Machine{
		Name:     "confluent",
		User:     netrcUsername,
		Password: netrcPassword,
		IsSSO:    false,
	}
)

type LoginCredentialsManagerTestSuite struct {
	suite.Suite
	require *require.Assertions

	ccloudClient *ccloud.Client
	logger       *log.Logger
	netrcHandler netrc.NetrcHandler
	prompt       *mock.Prompt

	loginCredentialsManager LoginCredentialsManager
}

func (suite *LoginCredentialsManagerTestSuite) SetupSuite() {
	suite.ccloudClient = &ccloud.Client{
		User: &sdkMock.User{
			CheckEmailFunc: func(ctx context.Context, user *orgv1.User) (*orgv1.User, error) {
				return &orgv1.User{
					Email: "",
				}, nil
			},
		},
	}
	suite.logger = log.New()
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
	suite.prompt = &mock.Prompt{
		ReadLineFunc: func() (string, error) {
			return promptUsername, nil
		},
		ReadLineMaskedFunc: func() (string, error) {
			return promptPassword, nil
		},
	}
}

func (suite *LoginCredentialsManagerTestSuite) SetupTest() {
	suite.require = require.New(suite.T())
	suite.clearCCloudEnvironmentVariables()
	suite.clearConfluentEnvironmentVariables()
	suite.loginCredentialsManager = NewLoginCredentialsManager(suite.netrcHandler, suite.prompt, suite.logger)
}

func (suite *LoginCredentialsManagerTestSuite) TestGetCCloudCredentialsFromEnvVar() {
	// incomplete credentials, setting on username but not password
	suite.require.NoError(os.Setenv(CCloudEmailEnvVar, envUsername))
	creds, err := suite.loginCredentialsManager.GetCCloudCredentialsFromEnvVar(&cobra.Command{})()
	suite.require.NoError(err)
	suite.require.Nil(creds)

	suite.setCCloudEnvironmentVariables()
	creds, err = suite.loginCredentialsManager.GetCCloudCredentialsFromEnvVar(&cobra.Command{})()
	suite.require.NoError(err)
	suite.compareCredentials(envCredentials, creds)
}

func (suite *LoginCredentialsManagerTestSuite) TestGetCCloudCredentialsFromDeprecatedEnvVar() {
	// incomplete credentials, setting on username but not password
	suite.require.NoError(os.Setenv(CCloudEmailDeprecatedEnvVar, deprecatedEnvUser))
	creds, err := suite.loginCredentialsManager.GetCCloudCredentialsFromEnvVar(&cobra.Command{})()
	suite.require.NoError(err)
	suite.require.Nil(creds)

	suite.setCCloudDeprecatedEnvironmentVariables()
	creds, err = suite.loginCredentialsManager.GetCCloudCredentialsFromEnvVar(&cobra.Command{})()
	suite.require.NoError(err)
	suite.compareCredentials(deprecateEnvCredentials, creds)
}

func (suite *LoginCredentialsManagerTestSuite) TestGetCCloudCredentialsFromEnvVarOrderOfPrecedence() {
	suite.setCCloudEnvironmentVariables()
	suite.setCCloudDeprecatedEnvironmentVariables()
	creds, err := suite.loginCredentialsManager.GetCCloudCredentialsFromEnvVar(&cobra.Command{})()
	suite.require.NoError(err)
	suite.compareCredentials(envCredentials, creds)
}

func (suite *LoginCredentialsManagerTestSuite) TestGetConfluentTokenAndCredentialsFromEnvVar() {
	// incomplete credentials, setting on username but not password
	suite.require.NoError(os.Setenv(ConfluentUsernameEnvVar, deprecatedEnvUser))
	creds, err := suite.loginCredentialsManager.GetConfluentCredentialsFromEnvVar(&cobra.Command{})()
	suite.require.NoError(err)
	suite.require.Nil(creds)

	suite.setConfluentEnvironmentVariables()
	creds, err = suite.loginCredentialsManager.GetConfluentCredentialsFromEnvVar(&cobra.Command{})()
	suite.require.NoError(err)
	suite.compareCredentials(envCredentials, creds)
}

func (suite *LoginCredentialsManagerTestSuite) TestGetConfluentCredentialsFromDeprecatedEnvVar() {
	// incomplete credentials, setting on username but not password
	suite.require.NoError(os.Setenv(ConfluentUsernameDeprecatedEnvVar, deprecatedEnvUser))
	creds, err := suite.loginCredentialsManager.GetConfluentCredentialsFromEnvVar(&cobra.Command{})()
	suite.require.NoError(err)
	suite.require.Nil(creds)

	suite.setConfluentDeprecatedEnvironmentVariables()
	creds, err = suite.loginCredentialsManager.GetConfluentCredentialsFromEnvVar(&cobra.Command{})()
	suite.require.NoError(err)
	suite.compareCredentials(deprecateEnvCredentials, creds)
}

func (suite *LoginCredentialsManagerTestSuite) TestGetConfluentCredentialsFromEnvVarOrderOfPrecedence() {
	suite.setConfluentEnvironmentVariables()
	suite.setConfluentDeprecatedEnvironmentVariables()
	creds, err := suite.loginCredentialsManager.GetConfluentCredentialsFromEnvVar(&cobra.Command{})()
	suite.require.NoError(err)
	suite.compareCredentials(envCredentials, creds)
}

func (suite *LoginCredentialsManagerTestSuite) TestCCloudUsernamePasswordGetCredentialsFromNetrc() {
	creds, err := suite.loginCredentialsManager.GetCredentialsFromNetrc(&cobra.Command{}, netrc.GetMatchingNetrcMachineParams{
		CLIName: "ccloud",
		IsSSO:   false,
	})()
	suite.require.NoError(err)
	suite.compareCredentials(netrcCredentials, creds)
}

func (suite *LoginCredentialsManagerTestSuite) TestCCloudSSOGetCCloudCredentialsFromNetrc() {
	creds, err := suite.loginCredentialsManager.GetCredentialsFromNetrc(&cobra.Command{}, netrc.GetMatchingNetrcMachineParams{
		CLIName: "ccloud",
		IsSSO:   true,
	})()
	suite.require.NoError(err)
	suite.compareCredentials(ssoCredentials, creds)
}

func (suite *LoginCredentialsManagerTestSuite) TestConfluentGetCredentialsFromNetrc() {
	creds, err := suite.loginCredentialsManager.GetCredentialsFromNetrc(&cobra.Command{}, netrc.GetMatchingNetrcMachineParams{
		CLIName: "confluent",
		IsSSO:   false,
	})()
	suite.require.NoError(err)
	suite.compareCredentials(netrcCredentials, creds)
}

func (suite *LoginCredentialsManagerTestSuite) TestGetCCloudCredentialsFromPrompt() {
	creds, err := suite.loginCredentialsManager.GetCCloudCredentialsFromPrompt(&cobra.Command{}, suite.ccloudClient)()
	suite.require.NoError(err)
	suite.compareCredentials(promptCredentials, creds)
}

func (suite *LoginCredentialsManagerTestSuite) TestGetConfluentCredentialsFromPrompt() {
	creds, err := suite.loginCredentialsManager.GetConfluentCredentialsFromPrompt(&cobra.Command{})()
	suite.require.NoError(err)
	suite.compareCredentials(promptCredentials, creds)
}

func (suite *LoginCredentialsManagerTestSuite) TestGetCredentialsFunction() {
	noCredentialsNetrcHandler := &mock.MockNetrcHandler{
		GetMatchingNetrcMachineFunc: func(params netrc.GetMatchingNetrcMachineParams) (*netrc.Machine, error) {
			return nil, nil
		},
		GetFileNameFunc: func() string {
			return netrcFileName
		},
	}

	// No credentials in env var and netrc so should look for prompt
	loginCredentialsManager := NewLoginCredentialsManager(noCredentialsNetrcHandler, suite.prompt, suite.logger)
	creds, err := GetLoginCredentials(
		loginCredentialsManager.GetCCloudCredentialsFromEnvVar(&cobra.Command{}),
		loginCredentialsManager.GetCredentialsFromNetrc(&cobra.Command{}, netrc.GetMatchingNetrcMachineParams{
			CLIName: "ccloud",
			IsSSO:   false,
		}),
		loginCredentialsManager.GetCCloudCredentialsFromPrompt(&cobra.Command{}, suite.ccloudClient),
	)
	suite.require.NoError(err)
	suite.compareCredentials(promptCredentials, creds)

	// No credentials in env var but credentials in netrc so netrc credentials should be returned
	loginCredentialsManager = NewLoginCredentialsManager(suite.netrcHandler, suite.prompt, suite.logger)
	creds, err = GetLoginCredentials(
		loginCredentialsManager.GetCCloudCredentialsFromEnvVar(&cobra.Command{}),
		loginCredentialsManager.GetCredentialsFromNetrc(&cobra.Command{}, netrc.GetMatchingNetrcMachineParams{
			CLIName: "ccloud",
			IsSSO:   false,
		}),
		loginCredentialsManager.GetCCloudCredentialsFromPrompt(&cobra.Command{}, suite.ccloudClient),
	)
	suite.require.NoError(err)
	suite.compareCredentials(netrcCredentials, creds)

	// Credentials in environment variables has highest order of precedence
	suite.setCCloudEnvironmentVariables()
	loginCredentialsManager = NewLoginCredentialsManager(suite.netrcHandler, suite.prompt, suite.logger)
	creds, err = GetLoginCredentials(
		loginCredentialsManager.GetCCloudCredentialsFromEnvVar(&cobra.Command{}),
		loginCredentialsManager.GetCredentialsFromNetrc(&cobra.Command{}, netrc.GetMatchingNetrcMachineParams{
			CLIName: "ccloud",
			IsSSO:   false,
		}),
		loginCredentialsManager.GetCCloudCredentialsFromPrompt(&cobra.Command{}, suite.ccloudClient),
	)
	suite.require.NoError(err)
	suite.compareCredentials(envCredentials, creds)
}

func (suite *LoginCredentialsManagerTestSuite) compareCredentials(expect, actual *Credentials) {
	suite.require.Equal(expect.Username, actual.Username)
	suite.require.Equal(expect.Password, actual.Password)
	suite.require.Equal(expect.IsSSO, actual.IsSSO)
}

func (suite *LoginCredentialsManagerTestSuite) clearCCloudEnvironmentVariables() {
	suite.require.NoError(os.Setenv(CCloudEmailEnvVar, ""))
	suite.require.NoError(os.Setenv(CCloudPasswordEnvVar, ""))
	suite.require.NoError(os.Setenv(CCloudEmailDeprecatedEnvVar, ""))
	suite.require.NoError(os.Setenv(CCloudPasswordDeprecatedEnvVar, ""))
}

func (suite *LoginCredentialsManagerTestSuite) setCCloudEnvironmentVariables() {
	suite.require.NoError(os.Setenv(CCloudEmailEnvVar, envUsername))
	suite.require.NoError(os.Setenv(CCloudPasswordEnvVar, envPassword))
}

func (suite *LoginCredentialsManagerTestSuite) setCCloudDeprecatedEnvironmentVariables() {
	suite.require.NoError(os.Setenv(CCloudEmailDeprecatedEnvVar, deprecatedEnvUser))
	suite.require.NoError(os.Setenv(CCloudPasswordDeprecatedEnvVar, deprecatedEnvPassword))
}

func (suite *LoginCredentialsManagerTestSuite) setConfluentEnvironmentVariables() {
	suite.require.NoError(os.Setenv(ConfluentUsernameEnvVar, envUsername))
	suite.require.NoError(os.Setenv(ConfluentPasswordEnvVar, envPassword))
}

func (suite *LoginCredentialsManagerTestSuite) setConfluentDeprecatedEnvironmentVariables() {
	suite.require.NoError(os.Setenv(ConfluentUsernameDeprecatedEnvVar, deprecatedEnvUser))
	suite.require.NoError(os.Setenv(ConfluentPasswordDeprecatedEnvVar, deprecatedEnvPassword))
}

func (suite *LoginCredentialsManagerTestSuite) clearConfluentEnvironmentVariables() {
	suite.require.NoError(os.Setenv(ConfluentUsernameEnvVar, ""))
	suite.require.NoError(os.Setenv(ConfluentPasswordEnvVar, ""))
	suite.require.NoError(os.Setenv(ConfluentUsernameDeprecatedEnvVar, ""))
	suite.require.NoError(os.Setenv(ConfluentPasswordDeprecatedEnvVar, ""))
}

func TestLoginCredentialsManager(t *testing.T) {
	suite.Run(t, new(LoginCredentialsManagerTestSuite))
}
