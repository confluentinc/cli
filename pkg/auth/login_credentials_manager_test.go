package auth

import (
	"fmt"
	"os"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	ccloudv1 "github.com/confluentinc/ccloud-sdk-go-v1-public"
	ccloudv1mock "github.com/confluentinc/ccloud-sdk-go-v1-public/mock"

	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/mock"
)

const (
	deprecatedEnvUser     = "deprecated-chrissy"
	deprecatedEnvPassword = "deprecated-password"

	envUsername = "env-username"
	envPassword = "env-password"

	promptUsername = "prompt-chrissy"
	promptPassword = "  prompt-password  "

	prerunURL  = "http://test"
	caCertPath = "cert-path"
)

var (
	envCredentials = &Credentials{
		Username: envUsername,
		Password: envPassword,
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
)

type LoginCredentialsManagerTestSuite struct {
	suite.Suite
	require *require.Assertions

	ccloudClient *ccloudv1.Client
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
			LoginRealmFunc: func(req *ccloudv1.GetLoginRealmRequest) (*ccloudv1.GetLoginRealmReply, error) {
				if req.Email == "test+sso@confluent.io" {
					return &ccloudv1.GetLoginRealmReply{IsSso: true, MfaRequired: false, Realm: "ccloud-local"}, nil
				}
				if req.Email == "test+mfa@confluent.io" {
					return &ccloudv1.GetLoginRealmReply{MfaRequired: true, Realm: "ccloud-local"}, nil
				}
				return &ccloudv1.GetLoginRealmReply{IsSso: false, MfaRequired: false, Realm: "ccloud-local"}, nil
			},
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
	suite.loginCredentialsManager = NewLoginCredentialsManager(suite.prompt, suite.ccloudClient)
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

	suite.require.NoError(os.Setenv(ConfluentCloudEmail, "test+mfa@confluent.io"))
	creds, err = suite.loginCredentialsManager.GetCloudCredentialsFromEnvVar("")()
	suite.require.NoError(err)
	suite.compareCredentials(&Credentials{Username: "test+mfa@confluent.io", IsMFA: true, Password: ""}, creds)

	suite.setCCEnvVars()
	creds, err = suite.loginCredentialsManager.GetCloudCredentialsFromEnvVar("")()
	suite.require.NoError(err)
	suite.compareCredentials(envCredentials, creds)
}

func (suite *LoginCredentialsManagerTestSuite) TestGetCCloudCredentialsFromEnvVarOrderOfPrecedence() {
	suite.setCCEnvVars()
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

func (suite *LoginCredentialsManagerTestSuite) TestGetConfluentCredentialsFromEnvVarOrderOfPrecedence() {
	suite.setCPEnvVars()
	creds, err := suite.loginCredentialsManager.GetOnPremCredentialsFromEnvVar()()
	suite.require.NoError(err)
	suite.compareCredentials(envCredentials, creds)
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
	suite.require.Equal(errors.NoUrlEnvVarErrorMsg, err.Error())
	suite.require.Nil(creds)

	// Set URL
	suite.require.NoError(os.Setenv(ConfluentPlatformMDSURL, prerunURL))
	creds, err = suite.loginCredentialsManager.GetOnPremPrerunCredentialsFromEnvVar()()
	suite.require.NoError(err)
	suite.compareCredentials(envPrerunCredentials, creds)

	// Set certificate-authority-path
	suite.require.NoError(os.Setenv(ConfluentPlatformCertificateAuthorityPath, caCertPath))
	creds, err = suite.loginCredentialsManager.GetOnPremPrerunCredentialsFromEnvVar()()
	suite.require.NoError(err)
	suite.compareCredentials(envPrerunCredentialsWithCaCertPath, creds)
}

func (suite *LoginCredentialsManagerTestSuite) TestGetCredentialsFunction() {
	cmd := &cobra.Command{}
	cmd.Flags().Bool("save", false, "test")
	// No credentials in CONFLUENT_CLOUD_EMAIL and CONFLUENT_CLOUD_PASSWORD env vars, so should look for prompt
	loginCredentialsManager := NewLoginCredentialsManager(suite.prompt, suite.ccloudClient)
	creds, err := GetLoginCredentials(
		loginCredentialsManager.GetCloudCredentialsFromEnvVar(""),
		loginCredentialsManager.GetCloudCredentialsFromPrompt(""),
	)
	fmt.Println("") // HACK: Newline needed to parse test output
	suite.require.NoError(err)
	suite.compareCredentials(promptCredentials, creds)

	// Credentials in CONFLUENT_CLOUD_EMAIL and CONFLUENT_CLOUD_PASSWORD environment variables has highest order of precedence
	suite.setCCEnvVars()
	loginCredentialsManager = NewLoginCredentialsManager(suite.prompt, suite.ccloudClient)
	creds, err = GetLoginCredentials(
		loginCredentialsManager.GetCloudCredentialsFromEnvVar(""),
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

func (suite *LoginCredentialsManagerTestSuite) clearCCEnvVars() {
	suite.require.NoError(os.Unsetenv(ConfluentCloudEmail))
	suite.require.NoError(os.Unsetenv(ConfluentCloudPassword))
}

func (suite *LoginCredentialsManagerTestSuite) setCPEnvVars() {
	suite.require.NoError(os.Setenv(ConfluentPlatformUsername, envUsername))
	suite.require.NoError(os.Setenv(ConfluentPlatformPassword, envPassword))
}

func (suite *LoginCredentialsManagerTestSuite) clearCPEnvVars() {
	suite.require.NoError(os.Unsetenv(ConfluentPlatformUsername))
	suite.require.NoError(os.Unsetenv(ConfluentPlatformPassword))
}

func TestLoginCredentialsManager(t *testing.T) {
	suite.Run(t, new(LoginCredentialsManagerTestSuite))
}
