package analytics_test

import (
	"fmt"
	"github.com/jonboulle/clockwork"
	"strconv"
	"testing"
	"time"

	segment "github.com/segmentio/analytics-go"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	v1 "github.com/confluentinc/ccloudapis/org/v1"
	"github.com/confluentinc/cli/internal/cmd"
	"github.com/confluentinc/cli/internal/pkg/analytics"
	"github.com/confluentinc/cli/internal/pkg/config"
	mock "github.com/confluentinc/cli/mock/analytics"
)

var (
	userNameContext = "login-tester@confluent.io"
	userNameCred    = "username-tester@confluent.io"
	apiKeyContext   = "api-key-context"
	apiKeyCred      = "api-key-ABCD1234"
	apiKey          = "ABCD1234"
	apiSecret       = "abcdABCD"
	userId          = int32(123)
	organizationId  = int32(321)
	userEmail       = "tester@confluent.io"

	otherUserId     = int32(111)
	otherUserEmail  = "other@confluent.io"
	otherUserContext    = "login-other@confluent.io"
	otherUserCred   = "username-other@confluent.io"

	ccloudName    = "ccloud"
	flagName      = "flag"
	flagArg       = "flagArg"
	arg1          = "arg1"
	arg2          = "arg2"
	errorMessage  = "error message"
	unknownCmd    = "unknown"

	version = "1.1.1.1.1.1"

	testTime = time.Date(1999, time.December, 31, 23, 59, 59, 0, time.UTC)
)

type AnalyticsTestSuite struct {
	suite.Suite
	config          *config.Config
	auth            *config.AuthConfig
	authOther       *config.AuthConfig

	analyticsClient analytics.Client
	mockClient      *mock.SegmentClient
	output          []segment.Message
}

func (suite *AnalyticsTestSuite) SetupSuite() {
	suite.config = config.New()
	suite.config.CLIName = ccloudName
	suite.createAuth()
	suite.createContexts()
	suite.createCredentials()
}

func (suite *AnalyticsTestSuite) SetupTest() {
	suite.output = make([]segment.Message, 0)
	suite.mockClient = &mock.SegmentClient{
		EnqueueFunc: func(m segment.Message) error {
			suite.output = append(suite.output, m)
			return nil
		},
		CloseFunc: func() error {return nil},
	}
	suite.analyticsClient = analytics.NewAnalyticsClient(suite.config.CLIName, suite.config, version, suite.mockClient, clockwork.NewFakeClockAt(testTime))
}

func (suite *AnalyticsTestSuite) TestSuccessWithFlagAndArgs() {
	// assume user already logged in
	suite.loginUser()

	req := require.New(suite.T())
	cobraCmd := &cobra.Command{
		Run:    func(cmd *cobra.Command, args []string) {},
		PreRun: suite.preRunFunc(),
	}
	cobraCmd.Flags().String(flagName, "", "")
	cobraCmd.SetArgs([]string{arg1, arg2, "--" + flagName + "=" + flagArg})
	command := cmd.Command{
		Command:   cobraCmd,
		Analytics: suite.analyticsClient,
	}
	err := command.Execute()
	req.NoError(err)

	req.Equal(1, len(suite.output))
	page, ok := suite.output[0].(segment.Page)
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
	req := require.New(suite.T())

	// make sure user is logged out
	suite.logOut()
	rootCmd := &cobra.Command{
		Use: suite.config.CLIName,
	}
	loginCmd := &cobra.Command{
		Use:    "login",
		Run:    func(cmd *cobra.Command, args []string) {
			suite.analyticsClient.SetCommandType(analytics.Login)
			suite.loginUser()
		},
		PreRun: suite.preRunFunc(),
	}
	rootCmd.AddCommand(loginCmd)
	command := cmd.Command{
		Command:   rootCmd,
		Analytics: suite.analyticsClient,
	}
	rootCmd.SetArgs([]string{"login"})
	err := command.Execute()
	req.NoError(err)

	req.Equal(2, len(suite.output))
	for _, msg := range suite.output {
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

func (suite *AnalyticsTestSuite) TestAnonymousIdReset() {
	req := require.New(suite.T())

	// make sure user is logged out
	suite.logOut()
	rootCmd := &cobra.Command{
		Use: suite.config.CLIName,
	}
	loginCmd := &cobra.Command{
		Use:    "login",
		PreRun: suite.preRunFunc(),
	}

	loginUserCmd := &cobra.Command{
		Use:    "user",
		Run:    func(cmd *cobra.Command, args []string) {
			suite.analyticsClient.SetCommandType(analytics.Login)
			suite.loginUser()
		},
		PreRun: suite.preRunFunc(),
	}
	loginCmd.AddCommand(loginUserCmd)

	loginOtherCmd := &cobra.Command{
		Use:    "other",
		Run:    func(cmd *cobra.Command, args []string) {
			suite.analyticsClient.SetCommandType(analytics.Login)
			suite.loginOtherUser()
		},
		PreRun: suite.preRunFunc(),
	}
	loginCmd.AddCommand(loginOtherCmd)

	rootCmd.AddCommand(loginCmd)
	command := cmd.Command{
		Command:   rootCmd,
		Analytics: suite.analyticsClient,
	}
	rootCmd.SetArgs([]string{"login", "user"})
	err := command.Execute()
	req.NoError(err)

	req.Equal(2, len(suite.output))
	var firstAnonId string
	for _, msg := range suite.output {
		switch msg.(type) {
		case segment.Page:
			page, ok := msg.(segment.Page)
			req.True(ok)
			firstAnonId = page.AnonymousId
		}
	}

	rootCmd.SetArgs([]string{"login", "other"})
	err = command.Execute()
	req.NoError(err)

	req.Equal(4, len(suite.output))
	var secondAnonId string
	for i := 2; i < 4; i++ {
		switch suite.output[i].(type) {
		case segment.Page:
			page, ok := suite.output[i].(segment.Page)
			req.True(ok)
			secondAnonId = page.AnonymousId
		}
	}

	req.NotEqual(firstAnonId, secondAnonId)
}

func (suite *AnalyticsTestSuite) TestUserNotLoggedIn() {
	// make sure user is logged out
	suite.logOut()

	req := require.New(suite.T())
	cobraCmd := &cobra.Command{
		Run:    func(cmd *cobra.Command, args []string) {},
		PreRun: suite.preRunFunc(),
	}
	command := cmd.Command{
		Command:   cobraCmd,
		Analytics: suite.analyticsClient,
	}
	err := command.Execute()
	req.NoError(err)

	req.Equal(1, len(suite.output))
	page, ok := suite.output[0].(segment.Page)
	req.True(ok)

	suite.checkPageBasic(page)
	suite.checkPageNotLoggedIn(page)
	suite.checkPageSuccess(page)
}

func (suite *AnalyticsTestSuite) TestSessionTimedOut() {
	req := require.New(suite.T())
	suite.loginUser()
	prevAnonId := suite.config.AnonymousId
	cobraCmd := &cobra.Command{
		Run:    func(cmd *cobra.Command, args []string) {},
		PreRun: func(cmd *cobra.Command, args []string) {
			err := suite.analyticsClient.SessionTimedOut()
			req.NoError(err)
			suite.logOut()
			suite.preRunFunc()(cmd, args)
		},
	}
	command := cmd.Command{
		Command:   cobraCmd,
		Analytics: suite.analyticsClient,
	}
	err := command.Execute()
	req.NoError(err)

	req.Equal(1, len(suite.output))
	page, ok := suite.output[0].(segment.Page)
	req.True(ok)

	suite.checkPageBasic(page)
	suite.checkPageNotLoggedIn(page)
	suite.checkPageSuccess(page)
	req.NotEqual(prevAnonId, suite.config.AnonymousId)
}

func (suite *AnalyticsTestSuite) TestErrorReturnedByCommand() {
	// assume user is logged in
	suite.loginUser()

	req := require.New(suite.T())
	cobraCmd := &cobra.Command{
		Use: "command",
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf(errorMessage)
		},
		PreRun: suite.preRunFunc(),
	}
	command := cmd.Command{
		Command:   cobraCmd,
		Analytics: suite.analyticsClient,
	}
	err := command.Execute()
	req.NotNil(err)

	req.Equal(1, len(suite.output))
	page, ok := suite.output[0].(segment.Page)
	req.True(ok)

	suite.checkPageBasic(page)
	suite.checkPageLoggedIn(page)
	suite.checkPageError(page)
}

func (suite *AnalyticsTestSuite) TestMalformedCommand() {
	req := require.New(suite.T())
	rootCmd := &cobra.Command{
		Use: suite.config.CLIName,
	}
	randomCmd := &cobra.Command{
		Use:    "random",
		Run:    func(cmd *cobra.Command, args []string) {},
		PreRun: suite.preRunFunc(),
	}
	rootCmd.AddCommand(randomCmd)
	command := cmd.Command{
		Command:   rootCmd,
		Analytics: suite.analyticsClient,
	}
	rootCmd.SetArgs([]string{unknownCmd})
	err := command.Execute()
	req.NotNil(err)

	req.Equal(1, len(suite.output))
	track, ok := suite.output[0].(segment.Track)
	req.True(ok)

	suite.checkMalformedCommandTrack(track)
}

func (suite *AnalyticsTestSuite) TestHideSecretForApiStore() {
	// login the user
	suite.loginUser()

	req := require.New(suite.T())
	rootCmd := &cobra.Command{
		Use: "ccloud",
	}
	apiCmd := &cobra.Command{
		Use: "api-key",
	}
	storeCmd := &cobra.Command{
		Use:    "store",
		Run:    func(cmd *cobra.Command, args []string) {},
		PreRun: suite.preRunFunc(),
	}
	apiCmd.AddCommand(storeCmd)
	rootCmd.AddCommand(apiCmd)
	command := cmd.Command{
		Command:   rootCmd,
		Analytics: suite.analyticsClient,
	}
	rootCmd.SetArgs([]string{"api-key", "store", apiKey, apiSecret})
	err := command.Execute()
	req.NoError(err)

	req.Equal(1, len(suite.output))
	page, ok := suite.output[0].(segment.Page)
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


// --------------------------- setup helper functions -------------------------------
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
		User:    user,
		Account: account,
	}
	suite.auth = auth

	otherUser := &v1.User{
		Id:             otherUserId,
		Email:          userEmail,
		OrganizationId: organizationId,
	}
	authOther := &config.AuthConfig{
		User:    otherUser,
		Account: account,
	}
	suite.authOther = authOther
}

func (suite *AnalyticsTestSuite) createContexts() {
	contexts := make(map[string]*config.Context)
	apiContext := &config.Context{
		Name:       apiKeyContext,
		Credential: apiKeyCred,
	}
	userContext := &config.Context{
		Name:       userNameContext,
		Credential: userNameCred,
	}
	otherContext := &config.Context{
		Name:       otherUserContext,
		Credential: otherUserCred,
	}
	contexts[apiKeyContext] = apiContext
	contexts[userNameContext] = userContext
	contexts[otherUserContext] = otherContext
	suite.config.Contexts = contexts
}

func (suite *AnalyticsTestSuite) createCredentials() {
	credentials := make(map[string]*config.Credential)
	apiCred := &config.Credential{
		APIKeyPair: &config.APIKeyPair{
			Key:    apiKey,
			Secret: apiSecret,
		},
		CredentialType: config.APIKey,
	}
	userCred := &config.Credential{
		Username:       userEmail,
		CredentialType: config.Username,
	}
	otherCred := &config.Credential{
		Username:       otherUserEmail,
		CredentialType: config.Username,
	}
	credentials[apiKeyCred] = apiCred
	credentials[userNameCred] = userCred
	credentials[otherUserCred] = otherCred
	suite.config.Credentials = credentials
}

// --------------------------- login, logout, context switching helpers -------------------------------
func (suite *AnalyticsTestSuite) loginUser() {
	suite.config.Auth = suite.auth
	suite.config.CurrentContext = userNameContext
}

func (suite *AnalyticsTestSuite) loginOtherUser() {
	suite.config.Auth = suite.authOther
	suite.config.CurrentContext = otherUserContext
}

func (suite *AnalyticsTestSuite) logOut() {
	suite.config.AuthToken = ""
	suite.config.Auth = nil
}

func (suite *AnalyticsTestSuite) apiKeyCredContext() {
	suite.config.Auth = nil
	suite.config.CurrentContext = apiKeyContext
}

// --------------------------- Check helpers -------------------------------
func (suite *AnalyticsTestSuite) checkPageBasic(page segment.Page) {
	req := require.New(suite.T())
	req.NotEqual("", page.AnonymousId)
	startTime, ok := page.Properties[analytics.StartTimePropertiesKey]
	req.True(ok)
	req.Equal(testTime, startTime)
	finishTime, ok := page.Properties[analytics.FinishTimePropertiesKey]
	req.True(ok)
	req.Equal(testTime, finishTime)
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
}

func (suite *AnalyticsTestSuite) checkPageSuccess(page segment.Page) {
	req := require.New(suite.T())
	_, ok := page.Properties[analytics.ErrorMsgPropertiesKey]
	req.False(ok)
	succeeded, ok := page.Properties[analytics.SucceededPropertiesKey]
	req.True(ok)
	req.True(succeeded.(bool))
}

func (suite *AnalyticsTestSuite) checkIdentify(identify segment.Identify) {
	req := require.New(suite.T())
	req.Equal(strconv.Itoa(int(userId)), identify.UserId)
	req.NotEqual("", identify.AnonymousId)
}

func (suite *AnalyticsTestSuite) checkMalformedCommandTrack(track segment.Track) {
	req := require.New(suite.T())
	errMsg, ok := track.Properties[analytics.ErrorMsgPropertiesKey]
	req.True(ok)
	req.Equal(fmt.Sprintf("unknown command \"%s\" for \"%s\"", unknownCmd, ccloudName), errMsg)
}

// ------------------------- PreRun --------------------------
func (suite *AnalyticsTestSuite) preRunFunc() func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		suite.analyticsClient.TrackCommand(cmd, args)
	}
}

func TestAnalyticsTestSuite(t *testing.T) {
	suite.Run(t, new(AnalyticsTestSuite))
}
