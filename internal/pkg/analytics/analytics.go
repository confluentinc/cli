//go:generate go run github.com/travisjeffery/mocker/cmd/mocker --prefix "Analytics" --dst ../../../mock/analytics.go --pkg mock --selfpkg github.com/confluentinc/cli analytics.go Client
package analytics

import (
	"strconv"
	"strings"

	"github.com/jonboulle/clockwork"
	segment "github.com/segmentio/analytics-go"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/confluentinc/cli/internal/pkg/config"
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
	secretCommandArgs     = map[string][]int{"ccloud api-key store": {1}}
	SecretValueString     = "<secret_value>"
	malformedCmdEventName = "Malformed Command Error"

	// these are exported to avoid import cycle with test (test is in package analytics_test)
	// @VisibleForTesting
	FlagsPropertiesKey      = "flags"
	ArgsPropertiesKey       = "args"
	OrgIdPropertiesKey      = "organization_id"
	EmailPropertiesKey      = "email"
	ErrorMsgPropertiesKey   = "error_message"
	StartTimePropertiesKey  = "start_time"
	FinishTimePropertiesKey = "finish_time"
	SucceededPropertiesKey  = "succeeded"
	CredentialPropertiesKey = "credential_type"
	ApiKeyPropertiesKey     = "api-key"
	VersionPropertiesKey    = "version"
	CliNameTraitsKey        = "cli_name"
)

// Logger struct that implements Segment's logger and redirects segments error log to debug log
type Logger struct {
	logger *log.Logger
}

func NewLogger(logger *log.Logger) *Logger {
	return &Logger{logger: logger}
}

func (l *Logger) Logf(format string, args ...interface{}) {
	l.logger.Debugf(format, args ...)
}

func (l *Logger) Errorf(format string, args ...interface{}) {
	l.logger.Debugf("[Segment Error] " + format, args ...)
}

type Client interface {
	SetStartTime()
	TrackCommand(cmd *cobra.Command, args []string)
	CatchHelpCall(rootCmd *cobra.Command, args []string)
	SendCommandSucceeded() error
	SendCommandFailed(e error) error
	SetCommandType(commandType CommandType)
	SessionTimedOut() error
	Close() error
}

type ClientObj struct {
	cliName string
	client  segment.Client
	config  *config.Config
	clock   clockwork.Clock

	// cache data until we flush events to segment (when each cmd call finishes)
	cmdCalled   string
	properties  segment.Properties
	user        userInfo
	cliVersion  string
	commandType CommandType
}

type userInfo struct {
	credentialType string
	id             string
	email          string
	organizationId string
	apiKey         string
}

func NewAnalyticsClient(cliName string, cfg *config.Config, version string, segmentClient segment.Client, clock clockwork.Clock) *ClientObj {
	client := &ClientObj{
		cliName:     cliName,
		client:      segmentClient,
		config:      cfg,
		properties:  make(segment.Properties),
		cliVersion:  version,
		clock:       clock,
		commandType: Other,
	}
	return client
}

// not in prerun because help calls do not trigger prerun
func (a *ClientObj) SetStartTime() {
	a.properties.Set(StartTimePropertiesKey, a.clock.Now())
}

func (a *ClientObj) TrackCommand(cmd *cobra.Command, args []string) {
	a.cmdCalled = cmd.CommandPath()
	a.addArgsProperties(cmd, args)
	a.addFlagProperties(cmd)
	a.properties.Set(VersionPropertiesKey, a.cliVersion)
	a.user = a.getUser()
}

func (a *ClientObj) SessionTimedOut() error {
	// just in case; redundant if config.DeleteUserAuth called before TrackCommand in prerunner.Anonymous()
	a.user = userInfo{}
	err := a.config.ResetAnonymousId()
	if err != nil {
		return errors.Wrap(err, "Unable to reset anonymous id")
	}
	return nil
}

// Cobra does not trigger prerun and postrun when help flag is true
func (a *ClientObj) CatchHelpCall(rootCmd *cobra.Command, args []string) {
	// non-help calls would already have triggered preruns
	if a.cmdCalled != "" {
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

func (a *ClientObj) SendCommandSucceeded() error {
	if a.commandType == Login || a.commandType == Init || a.commandType == ContextUse {
		err := a.loginHandler()
		if err != nil {
			return err
		}
	}
	a.properties.Set(SucceededPropertiesKey, true)
	a.properties.Set(FinishTimePropertiesKey, a.clock.Now())
	if err := a.sendPage(); err != nil {
		return err
	}
	// only reset anonymous id if logout from a username credential
	// preventing logouts that have no effects from resetting anonymous id
	if a.commandType == Logout && a.user.credentialType == config.Username.String() {
		if err := a.config.ResetAnonymousId(); err != nil {
			return err
		}
	}
	return nil
}

func (a *ClientObj) SendCommandFailed(e error) error {
	a.properties.Set(SucceededPropertiesKey, false)
	a.properties.Set(FinishTimePropertiesKey, a.clock.Now())
	a.properties.Set(ErrorMsgPropertiesKey, e.Error())
	if a.cmdCalled == "" {
		return a.malformedCommandError(e)
	}
	if err := a.sendPage(); err != nil {
		return err
	}
	return nil
}

func (a *ClientObj) SetCommandType(commandType CommandType) {
	a.commandType = commandType
}

func (a *ClientObj) Close() error {
	return a.client.Close()
}

// Helper Functions

func (a *ClientObj) sendPage() error {
	page := segment.Page{
		AnonymousId: a.config.AnonymousId,
		Name:        a.cmdCalled,
		Properties:  a.properties,
		UserId:      a.user.id,
	}
	a.addUserProperties()
	return a.client.Enqueue(page)
}

func (a *ClientObj) identify() error {
	identify := segment.Identify{
		AnonymousId: a.config.AnonymousId,
		UserId:      a.user.id,
	}
	traits := segment.Traits{}
	traits.Set(VersionPropertiesKey, a.cliVersion)
	traits.Set(CliNameTraitsKey, a.config.CLIName)
	traits.Set(CredentialPropertiesKey, a.user.credentialType)
	if a.user.credentialType == config.APIKey.String() {
		traits.Set(ApiKeyPropertiesKey, a.user.apiKey)
	}
	identify.Traits = traits
	return a.client.Enqueue(identify)
}

func (a *ClientObj) malformedCommandError(e error) error {
	a.user = a.getUser()
	track := segment.Track{
		AnonymousId: a.config.AnonymousId,
		Event:       malformedCmdEventName,
		Properties:  a.properties,
		UserId:      a.user.id,
	}
	a.addUserProperties()
	return a.client.Enqueue(track)
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
	a.properties.Set(FlagsPropertiesKey, flags)
}

func (a *ClientObj) addArgsProperties(cmd *cobra.Command, args []string) {
	argsCopy := make([]string, len(args))
	copy(argsCopy, args)
	if ids, ok := secretCommandArgs[cmd.CommandPath()]; ok {
		for _, i := range ids {
			argsCopy[i] = SecretValueString
		}
	}
	a.properties.Set(ArgsPropertiesKey, argsCopy)
}

func (a *ClientObj) addUserProperties() {
	a.properties.Set(CredentialPropertiesKey, a.user.credentialType)
	if a.config.CLIName == "ccloud" && a.user.credentialType == config.Username.String() {
		a.properties.Set(OrgIdPropertiesKey, a.user.organizationId)
		a.properties.Set(EmailPropertiesKey, a.user.email)
	}
	if a.user.credentialType == config.APIKey.String() {
		a.properties.Set(ApiKeyPropertiesKey, a.user.apiKey)
	}
}

func (a *ClientObj) getUser() userInfo {
	var user userInfo
	user.credentialType = a.getCredentialType()
	// If the user is not logged in
	if user.credentialType == "" {
		return user
	}
	if user.credentialType == config.APIKey.String() {
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
	if err := a.config.CheckLogin(); err != nil {
		return "", "", ""
	}
	user := a.config.Auth.User
	userId = strconv.Itoa(int(user.Id))
	organizationId = strconv.Itoa(int(user.OrganizationId))
	email = user.Email
	return userId, organizationId, email
}

func (a *ClientObj) getCPUsername() string {
	if err := a.config.CheckLogin(); err != nil {
		return ""
	}
	ctx := a.config.Contexts[a.config.CurrentContext]
	cred := a.config.Credentials[ctx.Credential]
	return cred.Username
}

func (a *ClientObj) getCredentialType() string {
	credType, err := a.config.CredentialType()
	if err != nil {
		return ""
	}
	switch credType {
	case config.Username:
		if a.config.CheckLogin() == nil {
			return config.Username.String()
		}
	case config.APIKey:
		return config.APIKey.String()
	}
	return ""
}

func (a *ClientObj) getCredApiKey() string {
	context, err := a.config.Context()
	if err != nil {
		return ""
	}
	if cred, ok := a.config.Credentials[context.Credential]; ok {
		return cred.APIKeyPair.Key
	}
	return ""
}

func (a *ClientObj) loginHandler() error {
	prevUser := a.user
	a.user = a.getUser()
	// prevUser not logged in, need to identify but no anonymous id reset
	if prevUser.credentialType == "" {
		return a.identify()
	}

	if a.isSwitchUserLogin(prevUser) {
		if err := a.config.ResetAnonymousId(); err != nil {
			return err
		}
		return a.identify()
	}
	return nil
}

func (a *ClientObj) isSwitchUserLogin(prevUser userInfo) bool {
	if prevUser.credentialType != a.user.credentialType {
		return true
	}
	if a.user.credentialType == config.Username.String() {
		if prevUser.id != a.user.id {
			return true
		}
	} else if a.user.credentialType == config.APIKey.String() {
		if a.user.apiKey != a.user.apiKey {
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
