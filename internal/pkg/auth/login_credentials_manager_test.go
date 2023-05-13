package auth

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	flowv1 "github.com/confluentinc/cc-structs/kafka/flow/v1"
	"github.com/confluentinc/ccloud-sdk-go-v1"
	sdkMock "github.com/confluentinc/ccloud-sdk-go-v1/mock"

	"github.com/confluentinc/cli/internal/pkg/errors"
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

	prerunNetrcUsername = "csreesangkom"
	prerunNetrPassword  = "password"

	prerunURL  = "http://test"
	caCertPath = "cert-path"
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
	envPrerunCredentials = &Credentials{
		Username:              envUsername,
		Password:              envPassword,
		PrerunLoginURL:        prerunURL,
		PrerunLoginCaCertPath: "",
	}
	envPrerunCredentialsWithCaCertPath = &Credentials{
		Username:              envUsername,
		Password:              envPassword,
		PrerunLoginURL:        prerunURL,
		PrerunLoginCaCertPath: caCertPath,
	}
	netrcPrerunCredentials = &Credentials{
		Username:              prerunNetrcUsername,
		Password:              prerunNetrPassword,
		PrerunLoginURL:        prerunURL,
		PrerunLoginCaCertPath: "",
	}
	netrcPrerunCredentialsWithCaCertPath = &Credentials{
		Username:              prerunNetrcUsername,
		Password:              prerunNetrPassword,
		PrerunLoginURL:        prerunURL,
		PrerunLoginCaCertPath: caCertPath,
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

	confluentNetrcPrerunMachine = &netrc.Machine{
		Name:     fmt.Sprintf("confluent-cli:mds-username-password:login-%s-%s", prerunNetrcUsername, prerunURL),
		User:     prerunNetrcUsername,
		Password: prerunNetrPassword,
		IsSSO:    false,
	}
	confluentNetrcPrerunMachineWithCaCertPath = &netrc.Machine{
		Name:     fmt.Sprintf("confluent-cli:mds-username-password:login-%s-%s?cacertpath=%s", prerunNetrcUsername, prerunURL, caCertPath),
		User:     prerunNetrcUsername,
		Password: prerunNetrPassword,
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
	params := &ccloud.Params{
		BaseURL: "https://devel.cpdev.cloud",
	}
	suite.ccloudClient = &ccloud.Client{
		Params: params,
		User: &sdkMock.User{
			LoginRealmFunc: func(ctx context.Context, req *flowv1.GetLoginRealmRequest) (*flowv1.GetLoginRealmReply, error) {
				if req.Email == "test+sso@confluent.io" {
					return &flowv1.GetLoginRealmReply{IsSso: true, Realm: "ccloud-local"}, nil
				}
				return &flowv1.GetLoginRealmReply{IsSso: false, Realm: "ccloud-local"}, nil
			},
		},
	}
	suite.logger = log.New()
	suite.netrcHandler = &mock.NetrcHandler{
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
	suite.loginCredentialsManager = NewLoginCredentialsManager(suite.netrcHandler, suite.prompt, suite.logger, suite.ccloudClient)
}

func (suite *LoginCredentialsManagerTestSuite) TestGetCCloudCredentialsFromEnvVar() {
	// incomplete credentials, setting on username but not password
	suite.require.NoError(os.Setenv(CCloudEmailEnvVar, envUsername))
	creds, err := suite.loginCredentialsManager.GetCCloudCredentialsFromEnvVar(&cobra.Command{})()
	suite.require.NoError(err)
	suite.require.Nil(creds)

	suite.require.NoError(os.Setenv(CCloudEmailEnvVar, "test+sso@confluent.io"))
	creds, err = suite.loginCredentialsManager.GetCCloudCredentialsFromEnvVar(&cobra.Command{})()
	suite.require.NoError(err)
	suite.compareCredentials(&Credentials{Username: "test+sso@confluent.io", IsSSO: true, Password: ""}, creds)

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
		URL:     "http://hi",
	})()
	suite.require.NoError(err)
	suite.compareCredentials(netrcCredentials, creds)
}

func (suite *LoginCredentialsManagerTestSuite) TestGetCCloudCredentialsFromPrompt() {
	creds, err := suite.loginCredentialsManager.GetCCloudCredentialsFromPrompt(&cobra.Command{})()
	suite.require.NoError(err)
	suite.compareCredentials(promptCredentials, creds)
}

func (suite *LoginCredentialsManagerTestSuite) TestGetConfluentCredentialsFromPrompt() {
	creds, err := suite.loginCredentialsManager.GetConfluentCredentialsFromPrompt(&cobra.Command{})()
	suite.require.NoError(err)
	suite.compareCredentials(promptCredentials, creds)
}

func (suite *LoginCredentialsManagerTestSuite) TestGetConfluentPrerunCredentialsFromEnvVar() {
	// incomplete and should, as there is no credentials set
	suite.require.NoError(os.Setenv(ConfluentURLEnvVar, prerunURL))
	creds, err := suite.loginCredentialsManager.GetConfluentPrerunCredentialsFromEnvVar(&cobra.Command{})()
	suite.require.Error(err)
	suite.require.Equal(errors.NoCredentialsFoundErrorMsg, err.Error())
	suite.require.Nil(creds)

	// incomplete and should return nil cred, as there is no password set even though username is set
	suite.require.NoError(os.Setenv(ConfluentUsernameEnvVar, envUsername))
	creds, err = suite.loginCredentialsManager.GetConfluentPrerunCredentialsFromEnvVar(&cobra.Command{})()
	suite.require.Error(err)
	suite.require.Equal(errors.NoCredentialsFoundErrorMsg, err.Error())
	suite.require.Nil(creds)

	// incomplete as this only sets username and password but not URL which is needed for Prerun login
	suite.require.NoError(os.Setenv(ConfluentURLEnvVar, ""))
	suite.setConfluentEnvironmentVariables()
	creds, err = suite.loginCredentialsManager.GetConfluentPrerunCredentialsFromEnvVar(&cobra.Command{})()
	suite.require.Error(err)
	suite.require.Equal(errors.NoURLEnvVarErrorMsg, err.Error())
	suite.require.Nil(creds)

	// Set URL
	suite.require.NoError(os.Setenv(ConfluentURLEnvVar, prerunURL))
	creds, err = suite.loginCredentialsManager.GetConfluentPrerunCredentialsFromEnvVar(&cobra.Command{})()
	suite.require.NoError(err)
	suite.compareCredentials(envPrerunCredentials, creds)

	// Set ca-cert-pat
	suite.require.NoError(os.Setenv(ConfluentCaCertPathEnvVar, caCertPath))
	creds, err = suite.loginCredentialsManager.GetConfluentPrerunCredentialsFromEnvVar(&cobra.Command{})()
	suite.require.NoError(err)
	suite.compareCredentials(envPrerunCredentialsWithCaCertPath, creds)

}

func (suite *LoginCredentialsManagerTestSuite) TestGetConfluentPrerunCredentialsFromNetrc() {
	// no cacertpath
	netrcHandler := &mock.NetrcHandler{
		GetMatchingNetrcMachineFunc: func(params netrc.GetMatchingNetrcMachineParams) (*netrc.Machine, error) {
			return confluentNetrcPrerunMachine, nil
		},
		GetFileNameFunc: func() string {
			return netrcFileName
		},
	}
	loginCredentialsManager := NewLoginCredentialsManager(netrcHandler, suite.prompt, suite.logger, suite.ccloudClient)
	creds, err := loginCredentialsManager.GetConfluentPrerunCredentialsFromNetrc(&cobra.Command{})()
	suite.require.NoError(err)
	suite.compareCredentials(netrcPrerunCredentials, creds)

	// with cacertpath
	netrcHandler = &mock.NetrcHandler{
		GetMatchingNetrcMachineFunc: func(params netrc.GetMatchingNetrcMachineParams) (*netrc.Machine, error) {
			return confluentNetrcPrerunMachineWithCaCertPath, nil
		},
		GetFileNameFunc: func() string {
			return netrcFileName
		},
	}
	loginCredentialsManager = NewLoginCredentialsManager(netrcHandler, suite.prompt, suite.logger, suite.ccloudClient)
	creds, err = loginCredentialsManager.GetConfluentPrerunCredentialsFromNetrc(&cobra.Command{})()
	suite.require.NoError(err)
	suite.compareCredentials(netrcPrerunCredentialsWithCaCertPath, creds)
}

func (suite *LoginCredentialsManagerTestSuite) TestGetCredentialsFunction() {
	noCredentialsNetrcHandler := &mock.NetrcHandler{
		GetMatchingNetrcMachineFunc: func(params netrc.GetMatchingNetrcMachineParams) (*netrc.Machine, error) {
			return nil, nil
		},
		GetFileNameFunc: func() string {
			return netrcFileName
		},
	}

	// No credentials in env var and netrc so should look for prompt
	loginCredentialsManager := NewLoginCredentialsManager(noCredentialsNetrcHandler, suite.prompt, suite.logger, suite.ccloudClient)
	creds, err := GetLoginCredentials(
		loginCredentialsManager.GetCCloudCredentialsFromEnvVar(&cobra.Command{}),
		loginCredentialsManager.GetCredentialsFromNetrc(&cobra.Command{}, netrc.GetMatchingNetrcMachineParams{
			CLIName: "ccloud",
			IsSSO:   false,
		}),
		loginCredentialsManager.GetCCloudCredentialsFromPrompt(&cobra.Command{}),
	)
	suite.require.NoError(err)
	suite.compareCredentials(promptCredentials, creds)

	// No credentials in env var but credentials in netrc so netrc credentials should be returned
	loginCredentialsManager = NewLoginCredentialsManager(suite.netrcHandler, suite.prompt, suite.logger, suite.ccloudClient)
	creds, err = GetLoginCredentials(
		loginCredentialsManager.GetCCloudCredentialsFromEnvVar(&cobra.Command{}),
		loginCredentialsManager.GetCredentialsFromNetrc(&cobra.Command{}, netrc.GetMatchingNetrcMachineParams{
			CLIName: "ccloud",
			IsSSO:   false,
		}),
		loginCredentialsManager.GetCCloudCredentialsFromPrompt(&cobra.Command{}),
	)
	suite.require.NoError(err)
	suite.compareCredentials(netrcCredentials, creds)

	// Credentials in environment variables has highest order of precedence
	suite.setCCloudEnvironmentVariables()
	loginCredentialsManager = NewLoginCredentialsManager(suite.netrcHandler, suite.prompt, suite.logger, suite.ccloudClient)
	creds, err = GetLoginCredentials(
		loginCredentialsManager.GetCCloudCredentialsFromEnvVar(&cobra.Command{}),
		loginCredentialsManager.GetCredentialsFromNetrc(&cobra.Command{}, netrc.GetMatchingNetrcMachineParams{
			CLIName: "ccloud",
			IsSSO:   false,
		}),
		loginCredentialsManager.GetCCloudCredentialsFromPrompt(&cobra.Command{}),
	)
	suite.require.NoError(err)
	suite.compareCredentials(envCredentials, creds)
}

func (suite *LoginCredentialsManagerTestSuite) compareCredentials(expect, actual *Credentials) {
	suite.require.Equal(expect.Username, actual.Username)
	suite.require.Equal(expect.Password, actual.Password)
	suite.require.Equal(expect.IsSSO, actual.IsSSO)

	// For prerun credentials only, for others theses should be empty strings
	suite.require.Equal(expect.PrerunLoginURL, actual.PrerunLoginURL)
	suite.require.Equal(expect.PrerunLoginCaCertPath, actual.PrerunLoginCaCertPath)
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
