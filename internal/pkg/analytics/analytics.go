//go:generate go run github.com/travisjeffery/mocker/cmd/mocker --prefix "Analytics" --dst ../../../mock/analytics.go --pkg mock --selfpkg github.com/confluentinc/cli analytics.go Client
package analytics

import (
	"strconv"
	"strings"

	"github.com/jonboulle/clockwork"
	segment "github.com/segmentio/analytics-go"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	v2 "github.com/confluentinc/cli/internal/pkg/config/v2"
	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/log"
)

type CommandType int

const (
	Other CommandType = iota
	Login
	Init
	ContextUse
	Logout
)

var (
	secretCommandFlags = map[string][]string{
		"ccloud init":                   {"api-secret"},
		"confluent master-key generate": {"passphrase", "local-secrets-file"},
		"confluent file rotate":         {"passphrase", "passphrase-new"},
	}
	// map command string to secret handler func
	secretCommandArgs     = map[string]func([]string) []string{"ccloud api-key store": apiKeyStoreSecretHandler}
	SecretValueString     = "<secret_value>"
	malformedCmdEventName = "Malformed Command Error"
	nonUser               = "no-user-info"

	// key used in tracking created and deleted resources
	ResourceIDPropertiesKey = "resource_id"

	// these are exported to avoid import cycle with test (test is in package analytics_test)
	// @VisibleForTesting
	FlagsPropertiesKey              = "flags"
	ArgsPropertiesKey               = "args"
	OrgIdPropertiesKey              = "organization_id"
	EmailPropertiesKey              = "email"
	ErrorMsgPropertiesKey           = "error_message"
	StartTimePropertiesKey          = "start_time"
	FinishTimePropertiesKey         = "finish_time"
	SucceededPropertiesKey          = "succeeded"
	CredentialPropertiesKey         = "credential_type"
	ApiKeyPropertiesKey             = "api-key"
	VersionPropertiesKey            = "version"
	CliNameTraitsKey                = "cli_name"
	ReleaseNotesErrorPropertiesKeys = "release_notes_error"
	FeedbackPropertiesKey           = "feedback"
)

// Logger struct that implements Segment's logger and redirects segments error log to debug log
type Logger struct {
	logger *log.Logger
}

func NewLogger(logger *log.Logger) *Logger {
	return &Logger{logger: logger}
}

func (l *Logger) Logf(format string, args ...interface{}) {
	l.logger.Debugf(format, args...)
}

func (l *Logger) Errorf(format string, args ...interface{}) {
	l.logger.Debugf("[Segment Error] "+format, args...)
}

type Client interface {
	SetStartTime()
	TrackCommand(cmd *cobra.Command, args []string)
	SetCommandType(commandType CommandType)
	SessionTimedOut() error
	SendCommandAnalytics(cmd *cobra.Command, args []string, cmdExecutionError error) error
	Close() error
	SetSpecialProperty(propertiesKey string, value interface{})
}

type cmdPage struct {
	cmdCalled   string
	properties  segment.Properties
	user        userInfo
	cliVersion  string
	commandType CommandType
}
type ClientObj struct {
	cliName string
	client  segment.Client
	config  *v3.Config
	clock   clockwork.Clock

	// cache data until we flush events to segment (when each cmd call finishes)
	cmdPages   []*cmdPage
	activeCmd  *cmdPage
	cliVersion string
}

type userInfo struct {
	credentialType string
	id             string
	email          string
	organizationId string
	apiKey         string
	anonymousId    string
}

func NewAnalyticsClient(cliName string, cfg *v3.Config, version string, segmentClient segment.Client, clock clockwork.Clock) *ClientObj {
	client := &ClientObj{
		cliName:    cliName,
		config:     cfg,
		client:     segmentClient,
		cliVersion: version,
		clock:      clock,
	}
	return client
}

// not in prerun because help calls do not trigger prerun
func (a *ClientObj) SetStartTime() {
	cmdPage := &cmdPage{
		cliVersion: a.cliVersion,
		properties: make(segment.Properties),
	}
	cmdPage.properties.Set(StartTimePropertiesKey, a.clock.Now())
	a.cmdPages = append(a.cmdPages, cmdPage)
	a.activeCmd = cmdPage
}

func (a *ClientObj) TrackCommand(cmd *cobra.Command, args []string) {
	a.activeCmd.cmdCalled = cmd.CommandPath()
	a.addArgsProperties(cmd, args)
	a.addFlagProperties(cmd)
	a.activeCmd.properties.Set(VersionPropertiesKey, a.activeCmd.cliVersion)
	a.activeCmd.user = a.getUser()
}

func (a *ClientObj) SessionTimedOut() error {
	// just in case; redundant if config.DeleteUserAuth called before TrackCommand in prerunner.Anonymous()
	a.activeCmd.user = userInfo{}
	if a.config != nil {
		err := a.resetAnonymousId()
		if err != nil {
			return err
		}
	}
	return nil
}

// Cobra does not trigger prerun and postrun when help flag is used
func (a *ClientObj) catchHelpCall(rootCmd *cobra.Command, args []string) {
	// non-help calls would already have triggered preruns
	if a.activeCmd.cmdCalled != "" {
		return
	}
	cmd, flags, err := rootCmd.Find(args)
	if err != nil {
		return
	}
	for _, flag := range flags {
		if isHelpFlag(flag) {
			a.TrackCommand(cmd, cmd.Flags().Args())
			break
		}
	}
}

func (a *ClientObj) SendCommandAnalytics(cmd *cobra.Command, args []string, cmdExecutionError error) error {
	a.catchHelpCall(cmd, args)
	if cmdExecutionError != nil {
		err := a.sendCommandFailed(cmdExecutionError)
		a.updateCmdPages()
		return err
	}
	err := a.sendCommandSucceeded()
	a.updateCmdPages()
	return err
}

func (a *ClientObj) updateCmdPages() {
	if len(a.cmdPages) > 1 {
		a.cmdPages = a.cmdPages[:len(a.cmdPages)-1]
		a.activeCmd = a.cmdPages[len(a.cmdPages)-1]
	} else {
		a.cmdPages = nil
		a.activeCmd = nil
	}
}

func (a *ClientObj) sendCommandSucceeded() error {
	if a.activeCmd.commandType == Login || a.activeCmd.commandType == Init || a.activeCmd.commandType == ContextUse {
		err := a.loginHandler()
		if err != nil {
			return err
		}
	}
	a.activeCmd.properties.Set(SucceededPropertiesKey, true)
	a.activeCmd.properties.Set(FinishTimePropertiesKey, a.clock.Now())
	if err := a.sendPage(); err != nil {
		return err
	}
	// only reset anonymous id if logout from a username credential
	// preventing logouts that have no effects from resetting anonymous id
	if a.activeCmd.commandType == Logout && a.activeCmd.user.credentialType == v2.Username.String() {
		if err := a.resetAnonymousId(); err != nil {
			return err
		}
	}
	return nil
}

func (a *ClientObj) sendCommandFailed(e error) error {
	a.activeCmd.properties.Set(SucceededPropertiesKey, false)
	a.activeCmd.properties.Set(FinishTimePropertiesKey, a.clock.Now())
	a.activeCmd.properties.Set(ErrorMsgPropertiesKey, e.Error())
	if a.activeCmd.cmdCalled == "" {
		return a.malformedCommandError()
	}
	if err := a.sendPage(); err != nil {
		return err
	}
	return nil
}

func (a *ClientObj) SetCommandType(commandType CommandType) {
	a.activeCmd.commandType = commandType
}

func (a *ClientObj) Close() error {
	return a.client.Close()
}

// for commands that need extra properties other than the common ones already set
func (a *ClientObj) SetSpecialProperty(propertiesKey string, value interface{}) {
	a.activeCmd.properties.Set(propertiesKey, value)
}

// Helper Functions

func (a *ClientObj) sendPage() error {
	page := segment.Page{
		Name:        a.activeCmd.cmdCalled,
		Properties:  a.activeCmd.properties,
		UserId:      a.activeCmd.user.id,
		AnonymousId: a.activeCmd.user.anonymousId,
	}
	if a.config != nil {
		a.addUserProperties()
	}
	return a.client.Enqueue(page)
}

func (a *ClientObj) identify() error {
	identify := segment.Identify{
		AnonymousId: a.activeCmd.user.anonymousId,
		UserId:      a.activeCmd.user.id,
	}
	traits := segment.Traits{}
	traits.Set(VersionPropertiesKey, a.activeCmd.cliVersion)
	traits.Set(CliNameTraitsKey, a.cliName)
	traits.Set(CredentialPropertiesKey, a.activeCmd.user.credentialType)
	if a.activeCmd.user.credentialType == v2.APIKey.String() {
		traits.Set(ApiKeyPropertiesKey, a.activeCmd.user.apiKey)
	}
	identify.Traits = traits
	return a.client.Enqueue(identify)
}

func (a *ClientObj) malformedCommandError() error {
	track := segment.Track{
		Event:      malformedCmdEventName,
		Properties: a.activeCmd.properties,
	}
	if a.config != nil {
		a.activeCmd.user = a.getUser()
		track.AnonymousId = a.activeCmd.user.anonymousId
		track.UserId = a.activeCmd.user.id
		a.addUserProperties()
	}
	return a.client.Enqueue(track)
}

func (a *ClientObj) resetAnonymousId() error {
	err := a.config.ResetAnonymousId()
	if err != nil {
		return errors.Wrap(err, "Unable to reset anonymous id")
	}
	a.activeCmd.user.anonymousId = a.config.AnonymousId
	return nil
}

func (a *ClientObj) addFlagProperties(cmd *cobra.Command) {
	flags := make(map[string]string)
	cmd.Flags().Visit(func(f *pflag.Flag) {
		if flagNames, ok := secretCommandFlags[cmd.CommandPath()]; ok {
			for _, flagName := range flagNames {
				if f.Name == flagName {
					flags[f.Name] = SecretValueString
					break
				}
			}
		}
		if _, ok := flags[f.Name]; !ok {
			flags[f.Name] = f.Value.String()
		}
	})
	a.activeCmd.properties.Set(FlagsPropertiesKey, flags)
}

func (a *ClientObj) addArgsProperties(cmd *cobra.Command, args []string) {
	argsLog := args
	if secretHandler, ok := secretCommandArgs[cmd.CommandPath()]; ok {
		argsLog = secretHandler(args)
	}
	a.activeCmd.properties.Set(ArgsPropertiesKey, argsLog)
}

func (a *ClientObj) addUserProperties() {
	a.activeCmd.properties.Set(CredentialPropertiesKey, a.activeCmd.user.credentialType)
	if a.cliName == "ccloud" && a.activeCmd.user.credentialType == v2.Username.String() {
		a.activeCmd.properties.Set(OrgIdPropertiesKey, a.activeCmd.user.organizationId)
		a.activeCmd.properties.Set(EmailPropertiesKey, a.activeCmd.user.email)
	}
	if a.activeCmd.user.credentialType == v2.APIKey.String() {
		a.activeCmd.properties.Set(ApiKeyPropertiesKey, a.activeCmd.user.apiKey)
	}
}

func (a *ClientObj) getUser() userInfo {
	var user userInfo
	if a.config == nil {
		return userInfo{
			id:          nonUser,
			anonymousId: nonUser,
		}
	}
	user.anonymousId = a.config.AnonymousId
	user.credentialType = a.getCredentialType()
	// If the user is not logged in
	if user.credentialType == "" {
		return user
	}
	if user.credentialType == v2.APIKey.String() {
		user.apiKey = a.getCredApiKey()
	}
	if a.cliName == "ccloud" {
		userId, organizationId, email := a.getCloudUserInfo()
		user.id = userId
		user.organizationId = organizationId
		user.email = email
	} else {
		user.id = a.getCPUsername()
	}
	return user
}

func (a *ClientObj) getCloudUserInfo() (userId, organizationId, email string) {
	if !a.config.HasLogin() {
		return "", "", ""
	}
	user := a.config.Context().State.Auth.User
	userId = strconv.Itoa(int(user.Id))
	organizationId = strconv.Itoa(int(user.OrganizationId))
	email = user.Email
	return userId, organizationId, email
}

func (a *ClientObj) getCPUsername() string {
	if !a.config.HasLogin() {
		return ""
	}
	ctx := a.config.Context()
	return ctx.Credential.Username
}

func (a *ClientObj) getCredentialType() string {
	switch a.config.CredentialType() {
	case v2.Username:
		if a.config.HasLogin() {
			return v2.Username.String()
		}
	case v2.APIKey:
		return v2.APIKey.String()
	}
	return ""
}

func (a *ClientObj) getCredApiKey() string {
	ctx := a.config.Context()
	if ctx == nil || ctx.Credential.APIKeyPair == nil {
		return ""
	}
	return ctx.Credential.APIKeyPair.Key
}

func (a *ClientObj) loginHandler() error {
	prevUser := a.activeCmd.user
	a.activeCmd.user = a.getUser()
	// prevUser not logged in, need to identify but no anonymous id reset
	if prevUser.credentialType == "" {
		return a.identify()
	}

	if a.isSwitchUserLogin(prevUser) {
		if err := a.resetAnonymousId(); err != nil {
			return err
		}
		return a.identify()
	}
	return nil
}

func (a *ClientObj) isSwitchUserLogin(prevUser userInfo) bool {
	if prevUser.credentialType != a.activeCmd.user.credentialType {
		return true
	}
	if a.activeCmd.user.credentialType == v2.Username.String() {
		if prevUser.id != a.activeCmd.user.id {
			return true
		}
	} else if a.activeCmd.user.credentialType == v2.APIKey.String() {
		if a.activeCmd.user.apiKey != a.activeCmd.user.apiKey {
			return true
		}
	}
	return false
}

func isHelpFlag(flag string) bool {
	if strings.HasPrefix(flag, "--") {
		return flag == "--help"
	} else if strings.HasPrefix(flag, "-") {
		return strings.Contains(flag, "h")
	}
	return false
}

func apiKeyStoreSecretHandler(args []string) []string {
	if len(args) < 2 {
		return args
	}
	if !(args[1] == "-" || strings.HasPrefix(args[1], "@")) {
		argsCopy := make([]string, len(args))
		copy(argsCopy, args)
		argsCopy[1] = SecretValueString
		return argsCopy
	}
	return args
}

func SendAnalyticsAndLog(cmd *cobra.Command, args []string, err error, client Client, logger *log.Logger) {
	analyticsError := client.SendCommandAnalytics(cmd, args, err)
	if analyticsError != nil {
		logger.Debugf("segment analytics sending event failed: %s\n", analyticsError.Error())
	}
}
