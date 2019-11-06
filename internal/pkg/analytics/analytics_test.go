package analytics_test

import (
	"fmt"
	"strconv"
	"testing"

	segment "github.com/segmentio/analytics-go"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	v1 "github.com/confluentinc/ccloudapis/org/v1"
	"github.com/confluentinc/cli/internal/cmd"
	"github.com/confluentinc/cli/internal/pkg/analytics"
	"github.com/confluentinc/cli/internal/pkg/config"
)

var (
	userNameContext = "login-tester@confluent.io"
	userNameCred = "username-tester@confluent.io"
	apiKeyContext = "api-key-context"
	apiKeyCred = "api-key-ABCD1234"
	apiKey = "ABCD1234"
	apiSecret = "abcdABCD"
	userId = int32(123)
	organizationId = int32(321)
	userEmail = "tester@confluent.io"

	ccloudCliName = "ccloud"
	flagName = "flag"
	flagArg = "flagArg"
	arg1 = "arg1"
	arg2 = "arg2"
	errorMessage = "error message"
)

type AnalyticsTestSuite struct {
	suite.Suite
	config *config.Config
	auth   *config.AuthConfig
}

func (suite *AnalyticsTestSuite) SetupSuite() {
	suite.config = &config.Config{
		CLIName: ccloudCliName,
		Auth:    nil,
	}
	suite.createAuth()
	suite.createContexts()
	suite.createCredentials()

}

func (suite *AnalyticsTestSuite) createAuth() {
	user := &v1.User{
		Id:             userId,
		Email:          userEmail,
		OrganizationId: organizationId,
	}
	account := &v1.Account{
		Id:             "1",
		Name:           "env1",
		OrganizationId: organizationId,
	}
	auth := &config.AuthConfig{
		User: user,
		Account: account,
	}
	suite.auth = auth
}

func (suite *AnalyticsTestSuite) createContexts() {
	contexts := make(map[string]*config.Context)
	apiContext := config.Context{
		Name:                   apiKeyContext,
		Credential:             apiKeyCred,
	}
	userContext := config.Context{
		Name:       userNameContext,
		Credential: userNameCred,
	}
	contexts[apiKeyContext] = &apiContext
	contexts[userNameContext] = &userContext
	suite.config.Contexts = contexts
}

func (suite *AnalyticsTestSuite) createCredentials() {
	credentials := make(map[string]*config.Credential)
	apiCred := config.Credential{
		APIKeyPair:     &config.APIKeyPair{
			Key:    apiKey,
			Secret: apiSecret,
		},
		CredentialType: 1,
	}
	userCred := config.Credential{
		Username:       "tester@confluent.io",
		CredentialType: 0,
	}
	credentials[apiKeyCred] = &apiCred
	credentials[userNameCred] = &userCred
	suite.config.Credentials = credentials
}

func (suite *AnalyticsTestSuite) setLoginConfig() {
	suite.config.Auth = suite.auth
	suite.config.CurrentContext = userNameContext
}

func (suite *AnalyticsTestSuite) logOut() {
	suite.config.Auth = nil
}

func (suite *AnalyticsTestSuite) apiKeyCredContext() {
	suite.config.CurrentContext = apiKeyContext
}

func (suite *AnalyticsTestSuite) TestSuccessWithFlagAndArgs() {
	// assume user already logged in
	suite.setLoginConfig()

	req := require.New(suite.T())
	l := make([]segment.Message, 0)
	out := &l
	mockClient := &analytics.MockSegmentClient{Out: out}
	analyticsClient := analytics.NewAnalyticsClient(suite.config, mockClient)
	cobraCmd := &cobra.Command{
		Run:    func(cmd *cobra.Command, args []string) {},
		PreRun: analyticsClient.TrackCommand,
	}
	cobraCmd.Flags().String(flagName, "", "")
	cobraCmd.SetArgs([]string{arg1, arg2, "--" + flagName + "=" + flagArg})
	command := cmd.Command{
		Command:   cobraCmd,
		Analytics: analyticsClient,
	}
	err := command.Execute()
	req.Nil(err)
	req.Equal(1, len(*out))
	page, ok := (*out)[0].(segment.Page)
	req.True(ok)

	suite.checkPageBasic(page)
	suite.checkPageLoggedIn(page)
	suite.checkPageSuccess(page)

	flags, ok := (page.Properties[analytics.FlagsPropertiesKey]).(map[string]string)
	req.True(ok)
	req.Equal(1, len(flags))
	flagVal, ok := flags[flagName]
	req.True(ok)
	req.Equal(flagArg, flagVal)

	args, ok := (page.Properties[analytics.ArgsPropertiesKey]).([]string)
	req.True(ok)
	req.Equal(2, len(args))
	req.Equal(arg1, args[0])
	req.Equal(arg2, args[1])
}

func (suite *AnalyticsTestSuite) TestLogin() {
	l := make([]segment.Message, 0)
	out := &l
	mockClient := &analytics.MockSegmentClient{Out: out}
	analyticsClient := analytics.NewAnalyticsClient(suite.config, mockClient)
	req := require.New(suite.T())

	suite.setLoginConfig()
	rootCmd := &cobra.Command{
		Use: suite.config.CLIName,
	}
	loginCmd := &cobra.Command{
		Use:    "login",
		Run:    func(cmd *cobra.Command, args []string) {},
		PreRun: analyticsClient.TrackCommand,
	}
	rootCmd.AddCommand(loginCmd)
	command := cmd.Command{
		Command:   rootCmd,
		Analytics: analyticsClient,
	}
	rootCmd.SetArgs([]string{"login"})
	err := command.Execute()
	req.Nil(err)
	req.Equal(2, len(*out))
	for _,msg := range *out {
		switch msg.(type) {
		case segment.Page:
			page, ok := msg.(segment.Page)
			req.True(ok)
			suite.checkPageSuccess(page)
			suite.checkPageBasic(page)
			suite.checkPageLoggedIn(page)
		case segment.Identify:
			identify, ok := msg.(segment.Identify)
			req.True(ok)
			suite.checkIdentify(identify)
		}
	}
}

func (suite *AnalyticsTestSuite) TestUserNotLoggedIn() {
	// make sure user is logged out
	suite.logOut()

	req := require.New(suite.T())
	l := make([]segment.Message, 0)
	out := &l
	mockClient := &analytics.MockSegmentClient{Out: out}
	analyticsClient := analytics.NewAnalyticsClient(suite.config, mockClient)
	cobraCmd := &cobra.Command{
		Run:    func(cmd *cobra.Command, args []string) {},
		PreRun: analyticsClient.TrackCommand,
	}
	command := cmd.Command{
		Command:   cobraCmd,
		Analytics: analyticsClient,
	}
	err := command.Execute()
	req.Nil(err)

	req.Equal(1, len(*out))
	page, ok := (*out)[0].(segment.Page)
	req.True(ok)

	suite.checkPageBasic(page)
	suite.checkPageNotLoggedIn(page)
	suite.checkPageSuccess(page)
}

func (suite *AnalyticsTestSuite) TestInternalError() {
	// assume user is logged in
	suite.setLoginConfig()

	req := require.New(suite.T())
	l := make([]segment.Message, 0)
	out := &l
	mockClient := &analytics.MockSegmentClient{Out: out}
	analyticsClient := analytics.NewAnalyticsClient(suite.config, mockClient)
	cobraCmd := &cobra.Command{
		Use:    "command",
		RunE:   func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf(errorMessage)
		},
		PreRun: analyticsClient.TrackCommand,
	}
	command := cmd.Command{
		Command:   cobraCmd,
		Analytics: analyticsClient,
	}
	err := command.Execute()
	req.NotNil(err)

	req.Equal(1, len(*out))
	page, ok := (*out)[0].(segment.Page)
	req.True(ok)

	suite.checkPageBasic(page)
	suite.checkPageLoggedIn(page)
	suite.checkPageError(page)
}

func (suite *AnalyticsTestSuite) TestMalformedCommand() {
	req := require.New(suite.T())
	l := make([]segment.Message, 0)
	out := &l
	mockClient := &analytics.MockSegmentClient{Out: out}
	analyticsClient := analytics.NewAnalyticsClient(suite.config, mockClient)
	rootCmd := &cobra.Command{
		Use: suite.config.CLIName,
	}
	randomCmd := &cobra.Command{
		Use:    "random",
		Run:    func(cmd *cobra.Command, args []string) {},
		PreRun: analyticsClient.TrackCommand,
	}
	rootCmd.AddCommand(randomCmd)
	command := cmd.Command{
		Command:   rootCmd,
		Analytics: analyticsClient,
	}
	rootCmd.SetArgs([]string{"notrandom"})
	err := command.Execute()
	req.NotNil(err)

	req.Equal(1, len(*out))
	track, ok := (*out)[0].(segment.Track)
	req.True(ok)

	suite.checkMalformedCommandTrack(track)
}

func (suite *AnalyticsTestSuite) TestHideSecretForApiStore() {
	// login the user
	suite.setLoginConfig()

	req := require.New(suite.T())
	l := make([]segment.Message, 0)
	out := &l
	mockClient := &analytics.MockSegmentClient{Out: out}
	analyticsClient := analytics.NewAnalyticsClient(suite.config, mockClient)
	rootCmd := &cobra.Command{
		Use: "ccloud",
	}
	apiCmd := &cobra.Command{
		Use: "api-key",
	}
	storeCmd := &cobra.Command{
		Use:    "store",
		Run:    func(cmd *cobra.Command, args []string) {},
		PreRun: analyticsClient.TrackCommand,
	}
	apiCmd.AddCommand(storeCmd)
	rootCmd.AddCommand(apiCmd)
	command := cmd.Command{
		Command:   rootCmd,
		Analytics: analyticsClient,
	}
	rootCmd.SetArgs([]string{"api-key", "store", apiKey, apiSecret})
	err := command.Execute()
	req.Nil(err)

	req.Equal(1, len(*out))
	page, ok := (*out)[0].(segment.Page)
	req.True(ok)

	suite.checkPageBasic(page)
	suite.checkPageLoggedIn(page)
	suite.checkPageSuccess(page)

	args, ok := (page.Properties[analytics.ArgsPropertiesKey]).([]string)
	req.True(ok)
	req.Equal(2, len(args))
	req.Equal(apiKey, args[0])
	req.Equal(analytics.SecretValueString, args[1])
}


func (suite *AnalyticsTestSuite) checkPageBasic(page segment.Page) {
	req := require.New(suite.T())
	req.NotEqual("", page.AnonymousId)
	_, ok := page.Properties[analytics.StartTimePropertiesKey]
	req.True(ok)
	_, ok = page.Properties[analytics.ArgsPropertiesKey]
	req.True(ok)
	_, ok = page.Properties[analytics.FlagsPropertiesKey]
	req.True(ok)
}

func (suite *AnalyticsTestSuite) checkPageLoggedIn(page segment.Page) {
	req := require.New(suite.T())

	req.Equal(strconv.Itoa(int(userId)), page.UserId)

	orgId, ok := page.Properties[analytics.OrgIdPropertiesKey]
	req.True(ok)
	req.Equal(strconv.Itoa(int(organizationId)), orgId)

	email, ok := page.Properties[analytics.EmailPropertiesKey]
	req.True(ok)
	req.Equal(userEmail, email)
}

func (suite *AnalyticsTestSuite) checkPageNotLoggedIn(page segment.Page) {
	req := require.New(suite.T())
	req.Equal("", page.UserId)
	_, ok := page.Properties[analytics.OrgIdPropertiesKey]
	req.False(ok)
	_, ok = page.Properties[analytics.EmailPropertiesKey]
	req.False(ok)
}

func (suite *AnalyticsTestSuite) checkPageError(page segment.Page) {
	req := require.New(suite.T())
	errorMsg, ok := page.Properties[analytics.ErrorMsgPropertiesKey]
	req.True(ok)
	req.Equal(errorMessage, errorMsg)
	succeeded, ok := page.Properties[analytics.SucceededPropertiesKey]
	req.True(ok)
	req.False(succeeded.(bool))
	_, ok = page.Properties[analytics.FinishTimePropertiesKey]
	req.False(ok)
}

func (suite *AnalyticsTestSuite) checkPageSuccess(page segment.Page) {
	req := require.New(suite.T())
	_, ok := page.Properties[analytics.ErrorMsgPropertiesKey]
	req.False(ok)
	succeeded, ok := page.Properties[analytics.SucceededPropertiesKey]
	req.True(ok)
	req.True(succeeded.(bool))
	_, ok = page.Properties[analytics.FinishTimePropertiesKey]
	req.True(ok)
}

func (suite *AnalyticsTestSuite) checkIdentify(identify segment.Identify) {
	req := require.New(suite.T())
	req.Equal(strconv.Itoa(int(userId)), identify.UserId)
	req.NotEqual("", identify.AnonymousId)
}

func (suite *AnalyticsTestSuite) checkMalformedCommandTrack(track segment.Track) {
	req := require.New(suite.T())
	_, ok := track.Properties[analytics.ErrorMsgPropertiesKey]
	req.True(ok)
}

func TestAnalyticsTestSuite(t *testing.T) {
	suite.Run(t, new(AnalyticsTestSuite))
}
