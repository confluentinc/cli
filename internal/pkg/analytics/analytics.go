//go:generate go run github.com/travisjeffery/mocker/cmd/mocker --prefix "" --dst mock/analytics.go --pkg mock --selfpkg github.com/confluentinc/cli analytics.go Client
package analytics

import (
	"github.com/jonboulle/clockwork"
	segment "github.com/segmentio/analytics-go"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"strconv"
	"strings"

	"github.com/confluentinc/cli/internal/pkg/config"
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
	FlagsPropertiesKey      = "flags"
	ArgsPropertiesKey       = "args"
	OrgIdPropertiesKey      = "organization_id"
	EmailPropertiesKey      = "email"
	ErrorMsgPropertiesKey   = "error_message"
	StartTimePropertiesKey  = "start_time"
	FinishTimePropertiesKey = "finish_time"
	SucceededPropertiesKey  = "succeeded"
	CredentialPropertiesKey = "credential"
	ApiKeyPropertiesKey     = "apikey"
	VersionPropertiesKey    = "version"
	CliNameTraitsKey        = "cli_name"

	apiKeyCred   = "apikey"
	userNameCred = "username"
)

type Client interface {
	TrackCommand(cmd *cobra.Command, args []string)
	FlushCommandSucceeded() error
	FlushCommandFailed(e error) error
}

type ClientObj struct {
	cliName string
	Client  segment.Client
	config  *config.Config
	clock   clockwork.Clock

	// cache data until we flush events to segment (when each cmd call finishes)
	cmdCalled   string
	properties  segment.Properties
	user        userInfo
	cliVersion  string
}

type userInfo struct {
	credential     string
	id             string
	email          string
	organizationId string
	apiKey         string
}

func NewAnalyticsClient(cliName string, cfg *config.Config, version string, segmentClient segment.Client, clock clockwork.Clock) *ClientObj {
	client := &ClientObj{
		cliName:    cliName,
		Client:     segmentClient,
		config:     cfg,
		properties: make(segment.Properties),
		cliVersion: version,
		clock:      clock,
	}
	return client
}

func (a *ClientObj) TrackCommand(cmd *cobra.Command, args []string) {
	a.cmdCalled = cmd.CommandPath()
	a.addArgsProperties(cmd, args)
	a.addFlagProperties(cmd)
	a.properties.Set(StartTimePropertiesKey, a.clock.Now())
	a.properties.Set(VersionPropertiesKey, a.cliVersion)
	a.user = a.getUser()
}

func (a *ClientObj) FlushCommandSucceeded() error {
	defer a.Client.Close()
	err := a.loginHandler()
	if err != nil {
		return err
	}
	a.properties.Set(SucceededPropertiesKey, true)
	a.properties.Set(FinishTimePropertiesKey, a.clock.Now())
	if err := a.sendPage(); err != nil {
		return err
	}
	if strings.Contains(a.cmdCalled, a.config.CLIName + " logout") {
		if err := a.config.ResetAnonymousId(); err != nil {
			return err
		}
	}
	return nil
}

func (a *ClientObj) FlushCommandFailed(e error) error {
	defer a.Client.Close()
	if a.cmdCalled == "" {
		return a.malformedCommandError(e)
	}
	a.properties.Set(SucceededPropertiesKey, false)
	a.properties.Set(FinishTimePropertiesKey, a.clock.Now())
	a.properties.Set(ErrorMsgPropertiesKey, e.Error())
	if err := a.sendPage(); err != nil {
		return err
	}
	return nil
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
	return a.Client.Enqueue(page)
}

func (a *ClientObj) identify() error {
	identify := segment.Identify{
		AnonymousId: a.config.AnonymousId,
		UserId:      a.user.id,
	}
	traits := segment.Traits{}
	traits.Set(VersionPropertiesKey, a.cliVersion)
	traits.Set(CliNameTraitsKey, a.config.CLIName)
	traits.Set(CredentialPropertiesKey, a.user.credential)
	if a.user.credential == apiKeyCred {
		traits.Set(ApiKeyPropertiesKey, a.user.apiKey)
	}
	identify.Traits = traits
	return a.Client.Enqueue(identify)
}

func (a *ClientObj) malformedCommandError(e error) error {
	defer a.Client.Close()
	a.user = a.getUser()
	a.properties.Set(ErrorMsgPropertiesKey, e.Error())
	track := segment.Track{
		AnonymousId: a.config.AnonymousId,
		Event:       malformedCmdEventName,
		Properties:  a.properties,
		UserId:      a.user.id,
	}
	a.addUserProperties()
	err := a.Client.Enqueue(track)
	if err != nil {
		return err
	}
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
	a.properties.Set(CredentialPropertiesKey, a.user.credential)
	if a.user.organizationId != "" {
		a.properties.Set(OrgIdPropertiesKey, a.user.organizationId)
		a.properties.Set(EmailPropertiesKey, a.user.email)
	}
	if a.user.credential == apiKeyCred {
		a.properties.Set(ApiKeyPropertiesKey, a.user.apiKey)
	}
}

func (a *ClientObj) getUser() userInfo {
	var user userInfo
	user.credential = a.getCredentialType()
	if user.credential == "none" {
		return user
	}
	if user.credential == apiKeyCred {
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

func (a *ClientObj) getCloudUserInfo() (userId string, organizationId string, email string) {
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
		return "none"
	}
	switch credType {
	case config.Username:
		if a.config.CheckLogin() == nil {
			return userNameCred
		}
	case config.APIKey:
		return apiKeyCred
	}
	return "none"
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
	// do nothing for non login commands
	if !(strings.Contains(a.cmdCalled, a.config.CLIName + " login") ||
	     strings.Contains(a.cmdCalled, "ccloud init") ||
		 strings.Contains(a.cmdCalled, "ccloud config context use")) {
		return nil
	}
	prevUser := a.user
	a.user = a.getUser()
	// previous not logged in need to identify but no need for anonymous id reset
	if prevUser.credential == "none" {
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
	if prevUser.credential != a.user.credential {
		return true
	}
	if a.user.credential == userNameCred {
		if prevUser.id != a.user.id {
			return true
		}
	} else if a.user.credential == apiKeyCred {
		if a.user.apiKey != a.user.apiKey {
			return true
		}
	}
	return false
}
