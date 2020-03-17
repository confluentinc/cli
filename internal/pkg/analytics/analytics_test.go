package analytics_test

// NOTE: All cobra commands must have "Use" field (name of the command) so that in
// analytics cmdCalled is not "", which will allow CatchHelpCalls to know that prerun is already run,
// so it will skip help flag catching.
// This prevents confusion in help flags catching because "make test" has flags like
// -test.testlogfile=/var/folders/n4/c3r14cc15zn9xfylw_gpdkh00000gp/T/go-build070382070/b377/testlog.txt
// which contains the "-" shorthand flag symbol, and can also contain an "h".

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/jonboulle/clockwork"
	segment "github.com/segmentio/analytics-go"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	orgv1 "github.com/confluentinc/ccloudapis/org/v1"

	"github.com/confluentinc/cli/internal/cmd"
	"github.com/confluentinc/cli/internal/pkg/analytics"
	v0 "github.com/confluentinc/cli/internal/pkg/config/v0"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	v2 "github.com/confluentinc/cli/internal/pkg/config/v2"
	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
	"github.com/confluentinc/cli/mock"
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

	otherUserId      = int32(111)
	otherUserEmail   = "other@confluent.io"
	otherUserContext = "login-other@confluent.io"
	otherUserCred    = "username-other@confluent.io"

	ccloudName   = "ccloud"
	flagName     = "flag"
	flagArg      = "flagArg"
	arg1         = "arg1"
	arg2         = "arg2"
	errorMessage = "error message"
	unknownCmd   = "unknown"

	version = "1.1.1.1.1.1"

	testTime = time.Date(1999, time.December, 31, 23, 59, 59, 0, time.UTC)
)

type AnalyticsTestSuite struct {
	suite.Suite
	config          *v3.Config
	analyticsClient analytics.Client
	mockClient      *mock.SegmentClient
	output          []segment.Message
}

func (suite *AnalyticsTestSuite) SetupSuite() {
	suite.config = v3.AuthenticatedCloudConfigMock()
	suite.config.CLIName = ccloudName
	suite.createContexts()
	suite.createStates()
	suite.createCredentials()
}

func (suite *AnalyticsTestSuite) SetupTest() {
	suite.output = make([]segment.Message, 0)
	suite.mockClient = &mock.SegmentClient{
		EnqueueFunc: func(m segment.Message) error {
			suite.output = append(suite.output, m)
			return nil
		},
		CloseFunc: func() error { return nil },
	}
	suite.analyticsClient = analytics.NewAnalyticsClient(suite.config.CLIName, suite.config, version, suite.mockClient, clockwork.NewFakeClockAt(testTime))
}

func (suite *AnalyticsTestSuite) TestHelpCall() {
	// assume user already logged in
	suite.loginUser()

	req := require.New(suite.T())
	cobraCmd := &cobra.Command{
		Use: suite.config.CLIName,
		Run: func(cmd *cobra.Command, args []string) {},
	}
	command := cmd.Command{
		Command:   cobraCmd,
		Analytics: suite.analyticsClient,
	}
	err := command.Execute([]string{"ccloud", "--help"})
	req.NoError(err)

	req.Equal(1, len(suite.output))
	page, ok := suite.output[0].(segment.Page)
	req.True(ok)

	suite.checkPageBasic(page)
	suite.checkPageLoggedIn(page)
	suite.checkPageSuccess(page)
	suite.checkHelpFlag(page)
}

func (suite *AnalyticsTestSuite) TestSuccessWithFlagAndArgs() {
	// assume user already logged in
	suite.loginUser()

	req := require.New(suite.T())
	cobraCmd := &cobra.Command{
		Use:    suite.config.CLIName,
		Run:    func(cmd *cobra.Command, args []string) {},
		PreRun: suite.preRunFunc(),
	}
	cobraCmd.Flags().String(flagName, "", "")
	command := cmd.Command{
		Command:   cobraCmd,
		Analytics: suite.analyticsClient,
	}
	err := command.Execute([]string{arg1, arg2, "--" + flagName, flagArg})
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

func (suite *AnalyticsTestSuite) TestHelpWithFlagAndArgs() {

	// assume user already logged in
	suite.loginUser()

	req := require.New(suite.T())
	cobraCmd := &cobra.Command{
		Use:    suite.config.CLIName,
		Run:    func(cmd *cobra.Command, args []string) {},
		PreRun: suite.preRunFunc(),
	}
	cobraCmd.Flags().String(flagName, "", "")
	command := cmd.Command{
		Command:   cobraCmd,
		Analytics: suite.analyticsClient,
	}
	err := command.Execute([]string{arg1, arg2, "--" + flagName, flagArg, "-h"})
	req.NoError(err)

	req.Equal(1, len(suite.output))
	page, ok := suite.output[0].(segment.Page)
	req.True(ok)

	suite.checkPageBasic(page)
	suite.checkPageLoggedIn(page)
	suite.checkPageSuccess(page)

	flags, ok := (page.Properties[analytics.FlagsPropertiesKey]).(map[string]string)
	req.True(ok)
	req.Equal(2, len(flags))
	flagVal, ok := flags[flagName]
	req.True(ok)
	req.Equal(flagArg, flagVal)
	req.Equal(flags["help"], "true")

	args, ok := (page.Properties[analytics.ArgsPropertiesKey]).([]string)
	req.True(ok)
	req.Equal(2, len(args))
	req.Equal(arg1, args[0])
	req.Equal(arg2, args[1])
}

func (suite *AnalyticsTestSuite) TestHelpWithFlagAndArgsSwapOrder() {
	req := require.New(suite.T())

	// make sure user is logged out
	suite.loginUser()
	rootCmd := &cobra.Command{
		Use: suite.config.CLIName,
	}
	loginCmd := &cobra.Command{
		Use:    "login",
		PreRun: suite.preRunFunc(),
	}

	loginUserCmd := &cobra.Command{
		Use: "user",
		Run: func(cmd *cobra.Command, args []string) {
			suite.loginUser()
		},
		PreRun: func(cmd *cobra.Command, args []string) {
			suite.preRunFunc()(cmd, args)
		},
	}
	loginUserCmd.Flags().String(flagName, "", "")
	loginCmd.AddCommand(loginUserCmd)

	rootCmd.AddCommand(loginCmd)
	command := cmd.Command{
		Command:   rootCmd,
		Analytics: suite.analyticsClient,
	}
	err := command.Execute([]string{"login", "--" + flagName, flagArg, "user", arg1, arg2, "--help"})
	req.NoError(err)

	req.Equal(1, len(suite.output))
	page, ok := suite.output[0].(segment.Page)
	req.True(ok)

	suite.checkPageBasic(page)
	suite.checkPageLoggedIn(page)
	suite.checkPageSuccess(page)

	flags, ok := (page.Properties[analytics.FlagsPropertiesKey]).(map[string]string)
	req.True(ok)
	req.Equal(2, len(flags))
	flagVal, ok := flags[flagName]
	req.True(ok)
	req.Equal(flagArg, flagVal)
	req.Equal(flags["help"], "true")

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
		Use: "login",
		Run: func(cmd *cobra.Command, args []string) {
			suite.loginUser()
		},
		PreRun: func(cmd *cobra.Command, args []string) {
			suite.analyticsClient.SetCommandType(analytics.Login)
			suite.preRunFunc()(cmd, args)
		},
	}
	rootCmd.AddCommand(loginCmd)
	command := cmd.Command{
		Command:   rootCmd,
		Analytics: suite.analyticsClient,
	}
	err := command.Execute([]string{"login"})
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
			suite.checkIdentify(identify, strconv.Itoa(int(userId)))
		default:
			suite.T().Error("Must be either Page or Identify event.")
		}
	}
}

func (suite *AnalyticsTestSuite) TestAnonymousIdResetOnLogin() {
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
		Use: "user",
		Run: func(cmd *cobra.Command, args []string) {
			suite.loginUser()
		},
		PreRun: func(cmd *cobra.Command, args []string) {
			suite.analyticsClient.SetCommandType(analytics.Login)
			suite.preRunFunc()(cmd, args)
		},
	}
	loginCmd.AddCommand(loginUserCmd)

	loginOtherCmd := &cobra.Command{
		Use: "other",
		Run: func(cmd *cobra.Command, args []string) {
			suite.loginOtherUser()
		},
		PreRun: func(cmd *cobra.Command, args []string) {
			suite.analyticsClient.SetCommandType(analytics.Login)
			suite.preRunFunc()(cmd, args)
		},
	}
	loginCmd.AddCommand(loginOtherCmd)

	rootCmd.AddCommand(loginCmd)
	command := cmd.Command{
		Command:   rootCmd,
		Analytics: suite.analyticsClient,
	}
	err := command.Execute([]string{"login", "user"})
	req.NoError(err)

	req.Equal(2, len(suite.output))
	var firstAnonId string
	for _, msg := range suite.output {
		switch msg.(type) {
		case segment.Page:
			page, ok := msg.(segment.Page)
			req.True(ok)
			firstAnonId = page.AnonymousId
		case segment.Identify:
			identify, ok := msg.(segment.Identify)
			req.True(ok)
			suite.checkIdentify(identify, strconv.Itoa(int(userId)))
		default:
			suite.T().Error("Must be Page or Identify event.")
		}
	}

	err = command.Execute([]string{"login", "other"})
	req.NoError(err)

	req.Equal(4, len(suite.output))
	var secondAnonId string
	for i := 2; i < 4; i++ {
		switch suite.output[i].(type) {
		case segment.Page:
			page, ok := suite.output[i].(segment.Page)
			req.True(ok)
			secondAnonId = page.AnonymousId
		case segment.Identify:
			identify, ok := suite.output[i].(segment.Identify)
			req.True(ok)
			suite.checkIdentify(identify, strconv.Itoa(int(otherUserId)))
		default:
			suite.T().Error("Must be Page or Identify event.")
		}
	}

	req.NotEqual(firstAnonId, secondAnonId)
}

func (suite *AnalyticsTestSuite) TestAnonymousIdResetOnContextSwitch() {
	req := require.New(suite.T())

	// log in with username cred
	suite.loginUser()

	firstAnonId := suite.config.AnonymousId

	contextUseCmd := &cobra.Command{
		Use: suite.config.CLIName,
		Run: func(cmd *cobra.Command, args []string) {
			suite.apiKeyCredContext()
		},
		PreRun: func(cmd *cobra.Command, args []string) {
			suite.analyticsClient.SetCommandType(analytics.ContextUse)
			suite.preRunFunc()(cmd, args)
		},
	}

	command := cmd.Command{
		Command:   contextUseCmd,
		Analytics: suite.analyticsClient,
	}

	err := command.Execute([]string{})
	req.NoError(err)

	req.Equal(2, len(suite.output))
	var secondAnonId string
	for _, msg := range suite.output {
		switch msg.(type) {
		case segment.Page:
			page, ok := msg.(segment.Page)
			req.True(ok)
			secondAnonId = page.AnonymousId
		case segment.Identify:
			identify, ok := msg.(segment.Identify)
			req.True(ok)
			suite.checkIdentify(identify, "")
		default:
			suite.T().Error("Must be Page or Identify event.")
		}
	}

	req.NotEqual(firstAnonId, secondAnonId)
}

func (suite *AnalyticsTestSuite) TestUserNotLoggedIn() {
	// make sure user is logged out
	suite.logOut()

	req := require.New(suite.T())
	cobraCmd := &cobra.Command{
		Use:    suite.config.CLIName,
		Run:    func(cmd *cobra.Command, args []string) {},
		PreRun: suite.preRunFunc(),
	}
	command := cmd.Command{
		Command:   cobraCmd,
		Analytics: suite.analyticsClient,
	}
	err := command.Execute([]string{})
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
		Use: suite.config.CLIName,
		Run: func(cmd *cobra.Command, args []string) {},
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
	err := command.Execute([]string{})
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
	err := command.Execute([]string{})
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
	err := command.Execute([]string{unknownCmd})
	req.NotNil(err)

	req.Equal(1, len(suite.output))
	track, ok := suite.output[0].(segment.Track)
	req.True(ok)

	suite.checkMalformedCommandTrack(track)
}

func (suite *AnalyticsTestSuite) TestApiKeyStoreSecretHandler() {
	// login the user
	suite.loginUser()

	req := require.New(suite.T())
	rootCmd := &cobra.Command{
		Use: suite.config.CLIName,
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

	pipeSymbol := "-"
	filePathArg := "@/file.txt"
	secondArgs := []string{apiSecret, pipeSymbol, filePathArg}
	for _, secondArg := range secondArgs {
		err := command.Execute([]string{"api-key", "store", apiKey, secondArg})
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
		if secondArg == apiSecret {
			req.Equal(analytics.SecretValueString, args[1])
		} else {
			req.Equal(secondArg, args[1])
		}
		suite.output = make([]segment.Message, 0)
	}
}

// --------------------------- setup helper functions -------------------------------
func (suite *AnalyticsTestSuite) createContexts() {
	platform := &v2.Platform{
		Name:       "test-platform",
		Server:     "test",
		CaCertPath: "",
	}
	apiContext := &v3.Context{
		Name:         apiKeyContext,
		Platform:     platform,
		PlatformName: platform.Name,
		KafkaClusterContext: &v3.KafkaClusterContext{
			EnvContext: true,
		},
		Config: suite.config,
	}
	apiContext.KafkaClusterContext.Context = apiContext
	userContext := &v3.Context{
		Name:         userNameContext,
		Platform:     platform,
		PlatformName: platform.Name,
		KafkaClusterContext: &v3.KafkaClusterContext{
			EnvContext: true,
		},
		Config: suite.config,
	}
	userContext.KafkaClusterContext.Context = userContext
	otherContext := &v3.Context{
		Name:         otherUserContext,
		Platform:     platform,
		PlatformName: platform.Name,
		KafkaClusterContext: &v3.KafkaClusterContext{
			EnvContext: true,
		},
		Config: suite.config,
	}
	otherContext.KafkaClusterContext.Context = otherContext
	contexts := make(map[string]*v3.Context)
	contexts[apiKeyContext] = apiContext
	contexts[userNameContext] = userContext
	contexts[otherUserContext] = otherContext
	suite.config.Contexts = contexts

	platforms := make(map[string]*v2.Platform)
	platforms[platform.Name] = platform
	suite.config.Platforms = platforms
}

func (suite *AnalyticsTestSuite) createStates() {
	contexts := suite.config.Contexts
	account := &orgv1.Account{
		Id:             "1",
		Name:           "env1",
		OrganizationId: organizationId,
	}
	userState := &v2.ContextState{
		Auth: &v1.AuthConfig{
			User: &orgv1.User{
				Id:             userId,
				Email:          userEmail,
				OrganizationId: organizationId,
			},
			Account: account,
		},
		AuthToken: "user-token",
	}
	contexts[userNameContext].State = userState
	otherUserState := &v2.ContextState{
		Auth: &v1.AuthConfig{
			User: &orgv1.User{
				Id:             otherUserId,
				Email:          userEmail,
				OrganizationId: organizationId,
			},
			Account: account,
		},
		AuthToken: "other-user-token",
	}
	contexts[otherUserContext].State = otherUserState
	contextStates := make(map[string]*v2.ContextState)
	contextStates[userNameContext] = contexts[userNameContext].State
	contextStates[otherUserContext] = contexts[otherUserContext].State
	suite.config.ContextStates = contextStates
}

func (suite *AnalyticsTestSuite) createCredentials() {
	credentials := make(map[string]*v2.Credential)
	apiCred := &v2.Credential{
		Name: apiKeyCred,
		APIKeyPair: &v0.APIKeyPair{
			Key:    apiKey,
			Secret: apiSecret,
		},
		CredentialType: v2.APIKey,
	}
	userCred := &v2.Credential{
		Name:           userNameCred,
		Username:       userEmail,
		CredentialType: v2.Username,
	}
	otherCred := &v2.Credential{
		Name:           otherUserCred,
		Username:       otherUserEmail,
		CredentialType: v2.Username,
	}
	contexts := suite.config.Contexts
	contexts[apiKeyContext].Credential = apiCred
	contexts[apiKeyContext].CredentialName = apiCred.Name
	contexts[userNameContext].Credential = userCred
	contexts[userNameContext].CredentialName = userCred.Name
	contexts[otherUserContext].Credential = otherCred
	contexts[otherUserContext].CredentialName = otherCred.Name
	credentials[apiKeyCred] = apiCred
	credentials[userNameCred] = userCred
	credentials[otherUserCred] = otherCred
	suite.config.Credentials = credentials
}

// --------------------------- login, logout, context switching helpers -------------------------------
func (suite *AnalyticsTestSuite) loginUser() {
	suite.config.CurrentContext = userNameContext
}

func (suite *AnalyticsTestSuite) loginOtherUser() {
	suite.config.CurrentContext = otherUserContext
}

func (suite *AnalyticsTestSuite) logOut() {
	suite.config.CurrentContext = ""
}

func (suite *AnalyticsTestSuite) apiKeyCredContext() {
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

func (suite *AnalyticsTestSuite) checkIdentify(identify segment.Identify, expectedUserId string) {
	req := require.New(suite.T())
	req.Equal(expectedUserId, identify.UserId)
	req.NotEqual("", identify.AnonymousId)
}

func (suite *AnalyticsTestSuite) checkMalformedCommandTrack(track segment.Track) {
	req := require.New(suite.T())
	errMsg, ok := track.Properties[analytics.ErrorMsgPropertiesKey]
	req.True(ok)
	req.Equal(fmt.Sprintf("unknown command \"%s\" for \"%s\"", unknownCmd, ccloudName), errMsg)
}

func (suite *AnalyticsTestSuite) checkHelpFlag(page segment.Page) {
	req := require.New(suite.T())
	flags, ok := (page.Properties[analytics.FlagsPropertiesKey]).(map[string]string)
	req.True(ok)
	flagVal, ok := flags["help"]
	req.True(ok)
	req.Equal("true", flagVal)
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
