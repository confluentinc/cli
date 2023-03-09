package auth

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	ccloudv1 "github.com/confluentinc/ccloud-sdk-go-v1-public"
	ccloudv1mock "github.com/confluentinc/ccloud-sdk-go-v1-public/mock"

	"github.com/confluentinc/cli/internal/pkg/errors"
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
	}
	confluentMachine = &netrc.Machine{
		Name:     "confluent",
		User:     netrcUsername,
		Password: netrcPassword,
	}

	confluentNetrcPrerunMachine = &netrc.Machine{
		Name:     fmt.Sprintf("confluent-cli:mds-username-password:login-%s-%s", prerunNetrcUsername, prerunURL),
		User:     prerunNetrcUsername,
		Password: prerunNetrPassword,
	}
	confluentNetrcPrerunMachineWithCaCertPath = &netrc.Machine{
		Name:     fmt.Sprintf("confluent-cli:mds-username-password:login-%s-%s?cacertpath=%s", prerunNetrcUsername, prerunURL, caCertPath),
		User:     prerunNetrcUsername,
		Password: prerunNetrPassword,
	}
)

type LoginCredentialsManagerTestSuite struct {
	suite.Suite
	require *require.Assertions

	ccloudClient *ccloudv1.Client
	netrcHandler netrc.NetrcHandler
	prompt       *mock.Prompt

	loginCredentialsManager LoginCredentialsManager
}

func (suite *LoginCredentialsManagerTestSuite) SetupSuite() {
	params := &ccloudv1.Params{
		BaseURL: "https://devel.cpdev.cloud",
	}
	suite.ccloudClient = &ccloudv1.Client{
		Params: params,
		User: &ccloudv1mock.UserInterface{
			LoginRealmFunc: func(ctx context.Context, req *ccloudv1.GetLoginRealmRequest) (*ccloudv1.GetLoginRealmReply, error) {
				if req.Email == "test+sso@confluent.io" {
					return &ccloudv1.GetLoginRealmReply{IsSso: true, Realm: "ccloud-local"}, nil
				}
				return &ccloudv1.GetLoginRealmReply{IsSso: false, Realm: "ccloud-local"}, nil
			},
		},
	}
	suite.netrcHandler = &mock.NetrcHandler{
		GetMatchingNetrcMachineFunc: func(params netrc.NetrcMachineParams) (*netrc.Machine, error) {
			if params.IsCloud {
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
	suite.clearCCEnvVars()
	suite.clearCPEnvVars()
	suite.loginCredentialsManager = NewLoginCredentialsManager(suite.netrcHandler, suite.prompt, suite.ccloudClient)
}

func (suite *LoginCredentialsManagerTestSuite) TestGetCCloudCredentialsFromEnvVar() {
	// incomplete credentials, setting on username but not password
	suite.require.NoError(os.Setenv(ConfluentCloudEmail, envUsername))
	creds, err := suite.loginCredentialsManager.GetCloudCredentialsFromEnvVar("")()
	suite.require.NoError(err)
	suite.require.Nil(creds)

	suite.require.NoError(os.Setenv(ConfluentCloudEmail, "test+sso@confluent.io"))
	creds, err = suite.loginCredentialsManager.GetCloudCredentialsFromEnvVar("")()
	suite.require.NoError(err)
	suite.compareCredentials(&Credentials{Username: "test+sso@confluent.io", IsSSO: true, Password: ""}, creds)

	suite.setCCEnvVars()
	creds, err = suite.loginCredentialsManager.GetCloudCredentialsFromEnvVar("")()
	suite.require.NoError(err)
	suite.compareCredentials(envCredentials, creds)
}

func (suite *LoginCredentialsManagerTestSuite) TestGetCCloudCredentialsFromDeprecatedEnvVar() {
	// incomplete credentials, setting on username but not password
	suite.require.NoError(os.Setenv(DeprecatedConfluentCloudEmail, deprecatedEnvUser))
	creds, err := suite.loginCredentialsManager.GetCloudCredentialsFromEnvVar("")()
	suite.require.NoError(err)
	suite.require.Nil(creds)

	suite.setDeprecatedCCEnvVars()
	creds, err = suite.loginCredentialsManager.GetCloudCredentialsFromEnvVar("")()
	suite.require.NoError(err)
	suite.compareCredentials(deprecateEnvCredentials, creds)
}

func (suite *LoginCredentialsManagerTestSuite) TestGetCCloudCredentialsFromEnvVarOrderOfPrecedence() {
	suite.setCCEnvVars()
	suite.setDeprecatedCCEnvVars()
	creds, err := suite.loginCredentialsManager.GetCloudCredentialsFromEnvVar("")()
	suite.require.NoError(err)
	suite.compareCredentials(envCredentials, creds)
}

func (suite *LoginCredentialsManagerTestSuite) TestGetConfluentTokenAndCredentialsFromEnvVar() {
	// incomplete credentials, setting on username but not password
	suite.require.NoError(os.Setenv(ConfluentPlatformUsername, deprecatedEnvUser))
	creds, err := suite.loginCredentialsManager.GetOnPremCredentialsFromEnvVar()()
	suite.require.NoError(err)
	suite.require.Nil(creds)

	suite.setCPEnvVars()
	creds, err = suite.loginCredentialsManager.GetOnPremCredentialsFromEnvVar()()
	suite.require.NoError(err)
	suite.compareCredentials(envCredentials, creds)
}

func (suite *LoginCredentialsManagerTestSuite) TestGetConfluentCredentialsFromDeprecatedEnvVar() {
	// incomplete credentials, setting on username but not password
	suite.require.NoError(os.Setenv(DeprecatedConfluentPlatformUsername, deprecatedEnvUser))
	creds, err := suite.loginCredentialsManager.GetOnPremCredentialsFromEnvVar()()
	suite.require.NoError(err)
	suite.require.Nil(creds)

	suite.setDeprecatedCPEnvVars()
	creds, err = suite.loginCredentialsManager.GetOnPremCredentialsFromEnvVar()()
	suite.require.NoError(err)
	suite.compareCredentials(deprecateEnvCredentials, creds)
}

func (suite *LoginCredentialsManagerTestSuite) TestGetConfluentCredentialsFromEnvVarOrderOfPrecedence() {
	suite.setCPEnvVars()
	suite.setDeprecatedCPEnvVars()
	creds, err := suite.loginCredentialsManager.GetOnPremCredentialsFromEnvVar()()
	suite.require.NoError(err)
	suite.compareCredentials(envCredentials, creds)
}

func (suite *LoginCredentialsManagerTestSuite) TestCCloudUsernamePasswordGetCredentialsFromNetrc() {
	cmd := &cobra.Command{}
	cmd.Flags().Bool("save", false, "test")
	creds, err := suite.loginCredentialsManager.GetCredentialsFromNetrc(netrc.NetrcMachineParams{
		IsCloud: true,
	})()
	suite.require.NoError(err)
	suite.compareCredentials(netrcCredentials, creds)
}

func (suite *LoginCredentialsManagerTestSuite) TestConfluentGetCredentialsFromNetrc() {
	cmd := &cobra.Command{}
	cmd.Flags().Bool("save", false, "test")
	creds, err := suite.loginCredentialsManager.GetCredentialsFromNetrc(netrc.NetrcMachineParams{
		IsCloud: false,
		URL:     "http://hi",
	})()
	suite.require.NoError(err)
	suite.compareCredentials(netrcCredentials, creds)
}

func (suite *LoginCredentialsManagerTestSuite) TestGetCCloudCredentialsFromPrompt() {
	creds, err := suite.loginCredentialsManager.GetCloudCredentialsFromPrompt("")()
	suite.require.NoError(err)
	suite.compareCredentials(promptCredentials, creds)
}

func (suite *LoginCredentialsManagerTestSuite) TestGetConfluentCredentialsFromPrompt() {
	creds, err := suite.loginCredentialsManager.GetOnPremCredentialsFromPrompt()()
	suite.require.NoError(err)
	suite.compareCredentials(promptCredentials, creds)
}

func (suite *LoginCredentialsManagerTestSuite) TestGetConfluentPrerunCredentialsFromEnvVar() {
	// incomplete and should, as there is no credentials set
	suite.require.NoError(os.Setenv(ConfluentPlatformMDSURL, prerunURL))
	creds, err := suite.loginCredentialsManager.GetOnPremPrerunCredentialsFromEnvVar()()
	suite.require.Error(err)
	suite.require.Equal(errors.NoCredentialsFoundErrorMsg, err.Error())
	suite.require.Nil(creds)

	// incomplete and should return nil cred, as there is no password set even though username is set
	suite.require.NoError(os.Setenv(ConfluentPlatformUsername, envUsername))
	creds, err = suite.loginCredentialsManager.GetOnPremPrerunCredentialsFromEnvVar()()
	suite.require.Error(err)
	suite.require.Equal(errors.NoCredentialsFoundErrorMsg, err.Error())
	suite.require.Nil(creds)

	// incomplete as this only sets username and password but not URL which is needed for Prerun login
	suite.require.NoError(os.Unsetenv(ConfluentPlatformMDSURL))
	suite.setCPEnvVars()
	creds, err = suite.loginCredentialsManager.GetOnPremPrerunCredentialsFromEnvVar()()
	suite.require.Error(err)
	suite.require.Equal(errors.NoURLEnvVarErrorMsg, err.Error())
	suite.require.Nil(creds)

	// Set URL
	suite.require.NoError(os.Setenv(ConfluentPlatformMDSURL, prerunURL))
	creds, err = suite.loginCredentialsManager.GetOnPremPrerunCredentialsFromEnvVar()()
	suite.require.NoError(err)
	suite.compareCredentials(envPrerunCredentials, creds)

	// Set ca-cert-path
	suite.require.NoError(os.Setenv(ConfluentPlatformCACertPath, caCertPath))
	creds, err = suite.loginCredentialsManager.GetOnPremPrerunCredentialsFromEnvVar()()
	suite.require.NoError(err)
	suite.compareCredentials(envPrerunCredentialsWithCaCertPath, creds)
}

func (suite *LoginCredentialsManagerTestSuite) TestGetConfluentPrerunCredentialsFromNetrc() {
	var params netrc.NetrcMachineParams

	// no cacertpath
	netrcHandler := &mock.NetrcHandler{
		GetMatchingNetrcMachineFunc: func(_ netrc.NetrcMachineParams) (*netrc.Machine, error) {
			return confluentNetrcPrerunMachine, nil
		},
		GetFileNameFunc: func() string {
			return netrcFileName
		},
	}
	loginCredentialsManager := NewLoginCredentialsManager(netrcHandler, suite.prompt, suite.ccloudClient)
	creds, err := loginCredentialsManager.GetOnPremPrerunCredentialsFromNetrc(&cobra.Command{}, params)()
	suite.require.NoError(err)
	suite.compareCredentials(netrcPrerunCredentials, creds)

	// with cacertpath
	netrcHandler = &mock.NetrcHandler{
		GetMatchingNetrcMachineFunc: func(_ netrc.NetrcMachineParams) (*netrc.Machine, error) {
			return confluentNetrcPrerunMachineWithCaCertPath, nil
		},
		GetFileNameFunc: func() string {
			return netrcFileName
		},
	}
	loginCredentialsManager = NewLoginCredentialsManager(netrcHandler, suite.prompt, suite.ccloudClient)
	creds, err = loginCredentialsManager.GetOnPremPrerunCredentialsFromNetrc(&cobra.Command{}, params)()
	suite.require.NoError(err)
	suite.compareCredentials(netrcPrerunCredentialsWithCaCertPath, creds)
}

func (suite *LoginCredentialsManagerTestSuite) TestGetCredentialsFunction() {
	noCredentialsNetrcHandler := &mock.NetrcHandler{
		GetMatchingNetrcMachineFunc: func(params netrc.NetrcMachineParams) (*netrc.Machine, error) {
			return nil, nil
		},
		GetFileNameFunc: func() string {
			return netrcFileName
		},
	}
	cmd := &cobra.Command{}
	cmd.Flags().Bool("save", false, "test")
	// No credentials in env var and netrc so should look for prompt
	loginCredentialsManager := NewLoginCredentialsManager(noCredentialsNetrcHandler, suite.prompt, suite.ccloudClient)
	creds, err := GetLoginCredentials(
		loginCredentialsManager.GetCloudCredentialsFromEnvVar(""),
		loginCredentialsManager.GetCredentialsFromNetrc(netrc.NetrcMachineParams{IsCloud: true}),
		loginCredentialsManager.GetCloudCredentialsFromPrompt(""),
	)
	fmt.Println("") // HACK: Newline needed to parse test output
	suite.require.NoError(err)
	suite.compareCredentials(promptCredentials, creds)

	// No credentials in env var but credentials in netrc so netrc credentials should be returned
	loginCredentialsManager = NewLoginCredentialsManager(suite.netrcHandler, suite.prompt, suite.ccloudClient)
	creds, err = GetLoginCredentials(
		loginCredentialsManager.GetCloudCredentialsFromEnvVar(""),
		loginCredentialsManager.GetCredentialsFromNetrc(netrc.NetrcMachineParams{IsCloud: true}),
		loginCredentialsManager.GetCloudCredentialsFromPrompt(""),
	)
	suite.require.NoError(err)
	suite.compareCredentials(netrcCredentials, creds)

	// Credentials in environment variables has highest order of precedence
	suite.setCCEnvVars()
	loginCredentialsManager = NewLoginCredentialsManager(suite.netrcHandler, suite.prompt, suite.ccloudClient)
	creds, err = GetLoginCredentials(
		loginCredentialsManager.GetCloudCredentialsFromEnvVar(""),
		loginCredentialsManager.GetCredentialsFromNetrc(netrc.NetrcMachineParams{IsCloud: true}),
		loginCredentialsManager.GetCloudCredentialsFromPrompt(""),
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

func (suite *LoginCredentialsManagerTestSuite) setCCEnvVars() {
	suite.require.NoError(os.Setenv(ConfluentCloudEmail, envUsername))
	suite.require.NoError(os.Setenv(ConfluentCloudPassword, envPassword))
}

func (suite *LoginCredentialsManagerTestSuite) setDeprecatedCCEnvVars() {
	suite.require.NoError(os.Setenv(DeprecatedConfluentCloudEmail, deprecatedEnvUser))
	suite.require.NoError(os.Setenv(DeprecatedConfluentCloudPassword, deprecatedEnvPassword))
}

func (suite *LoginCredentialsManagerTestSuite) clearCCEnvVars() {
	suite.require.NoError(os.Unsetenv(ConfluentCloudEmail))
	suite.require.NoError(os.Unsetenv(ConfluentCloudPassword))
	suite.require.NoError(os.Unsetenv(DeprecatedConfluentCloudEmail))
	suite.require.NoError(os.Unsetenv(DeprecatedConfluentCloudPassword))
}

func (suite *LoginCredentialsManagerTestSuite) setCPEnvVars() {
	suite.require.NoError(os.Setenv(ConfluentPlatformUsername, envUsername))
	suite.require.NoError(os.Setenv(ConfluentPlatformPassword, envPassword))
}

func (suite *LoginCredentialsManagerTestSuite) setDeprecatedCPEnvVars() {
	suite.require.NoError(os.Setenv(DeprecatedConfluentPlatformUsername, deprecatedEnvUser))
	suite.require.NoError(os.Setenv(DeprecatedConfluentPlatformPassword, deprecatedEnvPassword))
}

func (suite *LoginCredentialsManagerTestSuite) clearCPEnvVars() {
	suite.require.NoError(os.Unsetenv(ConfluentPlatformUsername))
	suite.require.NoError(os.Unsetenv(ConfluentPlatformPassword))
	suite.require.NoError(os.Unsetenv(DeprecatedConfluentPlatformUsername))
	suite.require.NoError(os.Unsetenv(DeprecatedConfluentPlatformPassword))
}

func TestLoginCredentialsManager(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(LoginCredentialsManagerTestSuite))
}
